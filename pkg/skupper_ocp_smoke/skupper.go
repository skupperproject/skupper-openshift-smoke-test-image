package skupper_ocp_smoke

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
	"time"

	//"testing"
	appsv1 "k8s.io/api/apps/v1"
)

func (cli *Client) DeploySkupper(ctx context.Context, skupperConfigMap *v1.ConfigMap) error {

	_, err := cli.KubeClient.CoreV1().ConfigMaps(cli.Namespace).Create(ctx, skupperConfigMap, v12.CreateOptions{})
	if err != nil {
		if !strings.Contains(err.Error(), "\"skupper-site\" already exists") {
			return fmt.Errorf("unable to deploy Skupper instance on namespace %s - reason : %s", cli.Namespace, err.Error())
		}
	}
	return nil
}

func (cli *Client) CreateSkupperToken(ctx context.Context) error {

	_, err := cli.KubeClient.CoreV1().Secrets(cli.Namespace).Create(ctx, &v1.Secret{
		ObjectMeta: v12.ObjectMeta{
			Name: "pub-secret",
			Labels: map[string]string{
				"skupper.io/type": "connection-token-request",
			},
		},
	}, v12.CreateOptions{})

	if err != nil {
		if !strings.Contains(err.Error(), "\"pub-secret\" already exists") {
			return fmt.Errorf("unable to create public token. Reason :  %s", err.Error())
		}
	}

	return nil
}

// Wait until the secret get populated by the site-controller
// It will add 3 sections to the Data field : ca.crt, tls.crt and tls.key
func (cli *Client) WaitSkupperTokenPopulated(ctx context.Context) (*v1.Secret, error) {

	t := time.NewTicker(time.Second * 1)
	var pubToken *v1.Secret
	for {
		select {
		case <-t.C:
			pubToken, err := cli.KubeClient.CoreV1().Secrets(cli.Namespace).Get(ctx, "pub-secret", v12.GetOptions{})
			if err != nil {
				return pubToken, fmt.Errorf("unable to retrieve pubToken details : %v", err.Error())
			}
			if len(pubToken.Data) >= 3 {
				return pubToken, nil
			}
		case <-ctx.Done():
			return pubToken, fmt.Errorf("timeout waiting for token creation")
		}
	}
}

func (cli *Client) CreateSkupperLink(ctx context.Context, populatedToken *v1.Secret) error {

	_, err := cli.KubeClient.CoreV1().Secrets(cli.Namespace).Create(ctx, &v1.Secret{
		ObjectMeta: v12.ObjectMeta{
			Name: "priv-secret",
			Labels: map[string]string{
				"skupper.io/type": "connection-token",
			},
			Annotations: populatedToken.Annotations,
		},
		Data: populatedToken.Data,
		Type: "kubernetes.io/tls",
	}, v12.CreateOptions{})
	if err != nil {
		if !strings.Contains(err.Error(), "\"priv-secret\" already exists") {
			return fmt.Errorf("unable to create link : %v", err.Error())
		}
	}
	return nil
}

func (cli *Client) WaitForLink(ctx context.Context, limit int, step string) error {

	// Check if the link is established
	timeused := 0
	for {
		privTokencheck, err := cli.KubeClient.CoreV1().ConfigMaps(cli.Namespace).Get(ctx, "skupper-internal", v12.GetOptions{})
		if err != nil {
			return fmt.Errorf("error while validating link creation : %s", err.Error())
		}

		if skroutercfg, ok := privTokencheck.Data["skrouterd.json"]; ok {
			if strings.Contains(skroutercfg, "\"name\": \"priv-secret\",") {
				PrintIfDebug(step, ": Skupper internal contains the definitions for the link.")
				return nil
			}
		}

		time.Sleep(5 * time.Second)
		if timeused >= limit {
			return fmt.Errorf("time limit achieved waiting for link validation")
		}
		timeused += 5
	}
}

// Create the deployment for the Frontend in public namespace
func (cli *Client) CreateDeploymentWithSkupper(ctx context.Context, name string, image string, port string) error {

	replicas := int32(1)
	dep := &appsv1.Deployment{
		ObjectMeta: v12.ObjectMeta{
			Name:      name,
			Namespace: cli.Namespace,
			Labels:    map[string]string{"app": name},
			Annotations: map[string]string{
				"skupper.io/proxy": "http",
				"skupper.io/port":  port,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &v12.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: v12.ObjectMeta{
					Labels: map[string]string{
						"app": name,
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:            name,
							Image:           image,
							ImagePullPolicy: v1.PullIfNotPresent},
					},
					RestartPolicy: v1.RestartPolicyAlways,
				},
			},
		},
	}

	// Deploying resource
	_, err := cli.KubeClient.AppsV1().Deployments(cli.Namespace).Create(ctx, dep, v12.CreateOptions{})
	if err != nil {
		if !strings.Contains(err.Error(), "\"priv-deploy\" already exists") {
			return fmt.Errorf("unable to create deployment %s on namespace %s. Reason : %s", name, cli.Namespace, err.Error())
		}
	}
	return nil
}
