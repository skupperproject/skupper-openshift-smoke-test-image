package skupper_ocp_smoke

import (
	"context"
	"fmt"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"os"
	osPath "path"
	"strings"
	"time"
)

func KubeConfigDefault(kubefile string, OSENV string) string {

	// Otherwise use the default
	kubeconfig := os.Getenv(OSENV)
	if kubeconfig == "" {
		homedir, _ := os.UserHomeDir()
		kubeconfig = osPath.Join(homedir, ".kube", kubefile)
	}

	// Validate that it exists
	_, err := os.Stat(kubeconfig)
	if err != nil {
		return ""
	}
	return kubeconfig
}

func (cli *Client) WaitPodsStatus(ctx context.Context, podStatus string, limit int, partPodName string, step string) error {

	timeused := 0
	for {
		pods, err := cli.KubeClient.CoreV1().Pods(cli.Namespace).List(ctx, v1.ListOptions{})
		if err != nil {
			return fmt.Errorf("unable to list pods - %s", partPodName)
		}
		for _, pod := range pods.Items {
			if strings.HasPrefix(pod.Name, partPodName) && string(pod.Status.Phase) == podStatus {
				return nil
			}
		}
		time.Sleep(5 * time.Second)
		PrintIfDebug(step, ": Waiting 5 seconds for pod", partPodName, " to be in", podStatus, "state")

		if timeused >= limit {
			return fmt.Errorf("time limit achieved waiting for pod %s to be in %s state on namespace %s", partPodName, podStatus, cli.Namespace)
		}
		timeused += 5
	}
}

func (cli *Client) WaitForService(ctx context.Context, limit int, serviceName string, step string) error {

	timeused := 0
	for {
		svcs, err := cli.KubeClient.CoreV1().Services(cli.Namespace).List(ctx, v1.ListOptions{})
		if err != nil {
			return fmt.Errorf("unable to list service %s : %s", serviceName, err.Error())
		}
		for _, svc := range svcs.Items {
			if svc.Name == serviceName {
				return nil
			}
		}
		time.Sleep(5 * time.Second)
		PrintIfDebug(step, ": Waiting 5 seconds for service", serviceName, " to be running in", cli.Namespace, "namespace")

		if timeused >= limit {
			return fmt.Errorf("time limit achieved waiting for service %s to be running in %s namespace", serviceName, cli.Namespace)
		}
		timeused += 5
	}
}

func PrintIfDebug(msg ...string) {
	if debug {
		log.Println(strings.Join(msg, " "))
	}
}
