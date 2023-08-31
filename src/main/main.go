package main

import (
	"context"
	"fmt"
	v13 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"os"
	"skupper_ocp_smoke/pkg/skupper_ocp_smoke"
	"strconv"
	"strings"
	"time"
)

func setup(ctx context.Context, pubCli *skupper_ocp_smoke.Client, privCli *skupper_ocp_smoke.Client) error {

	cliList := []*skupper_ocp_smoke.Client{
		pubCli,
		privCli,
	}

	step := "setup"
	log.Printf("%s : Starting Setup", step)

	//
	// Delete the namespace, if set by EnvVar
	//
	if strings.ToUpper(os.Getenv("CLEANBEFORE")) == "TRUE" {
		skupper_ocp_smoke.PrintIfDebug(step, ": Removing namespaces")
		for _, cli := range cliList {
			skupper_ocp_smoke.PrintIfDebug(step, ": Remove namespace", cli.Namespace)
			if err := cli.DeleteNamespace(ctx, step); err != nil {
				return fmt.Errorf("%s : %s", step, err.Error())
			}
			if err := cli.WaitForNSDeletion(ctx, cli.WaitLimit(), step); err != nil {
				return fmt.Errorf("%s : %s", step, err.Error())
			}
		}
	}

	// Create the namespace
	log.Printf("%s : Creating namespaces", step)
	for _, cli := range cliList {
		skupper_ocp_smoke.PrintIfDebug(step, ": Creating namespace", cli.Namespace)
		if err := cli.CreateNamespace(ctx, step); err != nil {
			return fmt.Errorf("%s : %s", step, err.Error())
		}
	}

	// Check if the Skupper operator is available
	// By default it checks for the operator "skupper-operator" in namespace "openshift-marketplace",
	// But that can be configured using ENV_VARS
	if err := pubCli.IsSkupperOperatorAvailable(step, pubCli.OperatorName(), pubCli.OperatorNameSpace(), pubCli.OperatorCatalog()); err != nil {
		return fmt.Errorf("%s : %s", step, err.Error())
	}

	// Create the Catalog Groups
	log.Printf("%s : Create OperatorGroups", step)
	for _, cli := range cliList {
		// Create the operatorgroup
		if err := cli.CreateOperatorGroup(step); err != nil {
			return fmt.Errorf("%s : %s", step, err.Error())
		}
		skupper_ocp_smoke.PrintIfDebug(step, ": OperatorGroup created into", cli.Namespace)
	}

	// Create the subscriptions
	log.Printf("%s : Create Subscriptions", step)
	for _, cli := range cliList {
		if err := cli.CreateSubscription(step); err != nil {
			return fmt.Errorf("%s : %s", step, err.Error())
		}

		// Wait until the operator is up
		if err := cli.WaitPodsStatus(ctx, "Running", cli.WaitLimit(), "skupper-site-controller", step); err != nil {
			return fmt.Errorf("%s : %s", step, err.Error())
		}
		skupper_ocp_smoke.PrintIfDebug(step, ": subscription created for", cli.Namespace)
	}

	// Public Skupper instance
	log.Printf("%s : Deploy instances", step)
	pubConfigMapDefinition := v1.ConfigMap{
		TypeMeta: v12.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: v12.ObjectMeta{
			Name:      "skupper-site",
			Namespace: pubCli.Namespace,
		},
		Immutable: nil,
		Data: map[string]string{
			"router-mode":      "interior",
			"console-user":     "admin",
			"console-password": "changeme",
		},
	}

	// Deploy the Public instance
	if err := pubCli.DeploySkupper(ctx, &pubConfigMapDefinition); err != nil {
		return fmt.Errorf("%s : %s", step, err.Error())
	}

	if err := pubCli.WaitPodsStatus(ctx, "Running", pubCli.WaitLimit(), "skupper-service-controller", step); err != nil {
		return fmt.Errorf("%s : %s", step, err.Error())
	}
	skupper_ocp_smoke.PrintIfDebug(step, ": Public instance is running")

	// Private Skupper instance
	privConfigMapDefinition := v1.ConfigMap{
		TypeMeta: v12.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: v12.ObjectMeta{
			Name:      "skupper-site",
			Namespace: privCli.Namespace,
		},
		Immutable: nil,
		Data: map[string]string{
			"router-mode":      "interior",
			"console-user":     "admin",
			"console-password": "changeme",
		},
	}

	// Deploy the Private instance
	if err := privCli.DeploySkupper(ctx, &privConfigMapDefinition); err != nil {
		return fmt.Errorf("%s : %s", step, err.Error())
	}

	if err := privCli.WaitPodsStatus(ctx, "Running", privCli.WaitLimit(), "skupper-service-controller", step); err != nil {
		return fmt.Errorf("%s : %s", step, err.Error())
	}
	skupper_ocp_smoke.PrintIfDebug(step, ": Private instance is running")

	// Create Token
	if err := pubCli.CreateSkupperToken(ctx); err != nil {
		return fmt.Errorf("%s : %s", step, err.Error())
	}
	log.Printf("%s : Public Token Created", step)

	pubToken, err := pubCli.WaitSkupperTokenPopulated(ctx)
	if err != nil {
		return fmt.Errorf("%s : %s", step, err.Error())
	}
	skupper_ocp_smoke.PrintIfDebug(step, ": Public Token populated")

	if err = privCli.CreateSkupperLink(ctx, pubToken); err != nil {
		return fmt.Errorf("%s : %s", step, err.Error())
	}
	skupper_ocp_smoke.PrintIfDebug(step, ": Link created into private namespace, Validating")

	if err = privCli.WaitForLink(ctx, privCli.WaitLimit(), step); err != nil {
		return fmt.Errorf("%s : %s", step, err.Error())
	}
	log.Printf("%s : Link creation validated", step)

	// Create the deployment in Public
	if err = privCli.CreateDeploymentWithSkupper(ctx, "priv-deploy", "hello-world-frontend", "8080"); err != nil {
		return fmt.Errorf("%s : %s", step, err.Error())
	}
	skupper_ocp_smoke.PrintIfDebug(step, ": Deployment created in", privCli.Namespace)

	// Check if the deployment is available in public
	if err = privCli.WaitForService(ctx, privCli.WaitLimit(), "priv-deploy", step); err != nil {
		return fmt.Errorf("%s : %s", step, err.Error())
	}
	skupper_ocp_smoke.PrintIfDebug(step, ": Service is running in", privCli.Namespace)

	// Check if the deployment is available in public, via Skupper
	if err = pubCli.WaitForService(ctx, pubCli.WaitLimit(), "priv-deploy", step); err != nil {
		return fmt.Errorf("%s : %s", step, err.Error())
	}
	skupper_ocp_smoke.PrintIfDebug(step, ": Service is running in", pubCli.Namespace)

	return nil
}

func runTheJob(ctx context.Context, pubCli *skupper_ocp_smoke.Client, limit int) error {

	// Run curl and check if it is available in the public namespace
	backoffLimit := int32(0)

	// Container Definition
	container := []v1.Container{
		{
			Name:  "testjob",
			Image: "curlimages/curl",
			Command: []string{
				"curl",
				"http://priv-deploy:8080",
			},
		},
	}

	pubJob := v13.Job{
		TypeMeta: v12.TypeMeta{
			Kind:       "Job",
			APIVersion: "batch/v1",
		},
		ObjectMeta: v12.ObjectMeta{
			Name: "testjob",
		},
		Spec: v13.JobSpec{
			BackoffLimit: &backoffLimit,
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers:    container,
					RestartPolicy: "Never",
					DNSPolicy:     "ClusterFirst",
				},
			},
		},
	}

	log.Printf("runtest : Starting job to validate test execution")
	_, err := pubCli.KubeClient.BatchV1().Jobs(pubCli.Namespace).Create(ctx, &pubJob, v12.CreateOptions{})
	if err != nil {
		return fmt.Errorf("runtest : unable to run job on namespace %s : %s", pubCli.Namespace, err.Error())
	}

	timeused := 0
	for {
		jobRan, err := pubCli.KubeClient.BatchV1().Jobs(pubCli.Namespace).Get(ctx, pubJob.Name, v12.GetOptions{})
		if err != nil {
			skupper_ocp_smoke.PrintIfDebug("runtest : Unable to retrieve details from job")
			return fmt.Errorf("runtedt : Unable to retrieve details from job")
		}
		if jobRan.Status.Succeeded > 0 {
			skupper_ocp_smoke.PrintIfDebug("runtest : Job ran in", pubCli.Namespace)
			return nil
		}

		skupper_ocp_smoke.PrintIfDebug("runtest : Waiting 5 seconds for the job end")
		time.Sleep(5 * time.Second)

		if timeused >= limit {
			return fmt.Errorf("time limit achieved waiting for job to finish")
		}
		timeused += 5
	}
}

func teardown(ctx context.Context, pubCli, privCli *skupper_ocp_smoke.Client) error {

	step := "teardown"
	log.Printf("%s : Starting Teardown", step)

	cliList := []*skupper_ocp_smoke.Client{
		pubCli,
		privCli,
	}

	// Delete the subscription
	for _, cli := range cliList {
		err := cli.DeleteSubscription("teardown")
		if err != nil {
			return fmt.Errorf("%s : %s", step, err.Error())
		}
		skupper_ocp_smoke.PrintIfDebug(step, ": subscription deleted from namespace", cli.Namespace)
	}

	// Delete the operator group
	for _, cli := range cliList {
		err := cli.DeleteOperatorGroup("teardown")
		if err != nil {
			return fmt.Errorf("%s : %s", step, err.Error())
		}
		skupper_ocp_smoke.PrintIfDebug(step, ": OperatorGroup deleted from", cli.Namespace)
	}

	for _, cli := range cliList {
		if err := cli.DeleteNamespace(ctx, "teardown"); err != nil {
			return fmt.Errorf("%s : %s", step, err.Error())
		}
		skupper_ocp_smoke.PrintIfDebug(step, ": Namespace deletion scheduled for", cli.Namespace)

		if err := cli.WaitForNSDeletion(ctx, cli.WaitLimit(), "teardown"); err != nil {
			return fmt.Errorf("%s : %s", step, err.Error())
		}
	}
	log.Printf("%s : Teardown finished", step)
	return nil
}

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {

	log.Printf("main : Starting main test")

	skupper_ocp_smoke.PrintIfDebug("main : Debug detail")

	if skupper_ocp_smoke.Startingcsv != "" {
		skupper_ocp_smoke.PrintIfDebug("main : Using specific startingCSV = ", skupper_ocp_smoke.Startingcsv)
	}

	if skupper_ocp_smoke.Waitlimit != skupper_ocp_smoke.WAITLIMIT {
		skupper_ocp_smoke.PrintIfDebug("main : Using specific timeout value = ", strconv.Itoa(skupper_ocp_smoke.Waitlimit))
	}

	log.Printf("main : Get Kubeconfig settings\n")
	pubKubeconfigFile := skupper_ocp_smoke.KubeConfigDefault(skupper_ocp_smoke.PUBKUBECONFIGFILE, "PUBKUBECONFIG")
	if pubKubeconfigFile == "" {
		return fmt.Errorf("unable to determine the kubeconfig for Public instance. Aborting")
	}

	privKubeconfigFile := skupper_ocp_smoke.KubeConfigDefault(skupper_ocp_smoke.PRIVKUBECONFIGFILE, "PRIVKUBECONFIG")
	if privKubeconfigFile == "" {
		return fmt.Errorf("unable to determine the kubeconfig for Private instance. Aborting")
	}

	skupper_ocp_smoke.PrintIfDebug("main : Kubeconfig for public namespace = ", os.Getenv("PUBKUBECONFIG"))
	skupper_ocp_smoke.PrintIfDebug("main : Kubeconfig for private namespace = ", os.Getenv("PRIVKUBECONFIG"))

	// Context and CLI
	ctx, cn := context.WithTimeout(context.Background(), time.Minute*30)
	defer cn()

	pubCli, _ := skupper_ocp_smoke.NewClient("", pubKubeconfigFile, skupper_ocp_smoke.PUBNS)
	privCli, _ := skupper_ocp_smoke.NewClient("", privKubeconfigFile, skupper_ocp_smoke.PRIVNS)

	// Run the Setup and test
	if strings.ToUpper(os.Getenv("STEP")) == "SETUP" || os.Getenv("STEP") == "" {
		err := setup(ctx, pubCli, privCli)
		if err != nil {
			return fmt.Errorf("%s", err.Error())
		}
	}

	if strings.ToUpper(os.Getenv("STEP")) == "RUNTEST" || os.Getenv("STEP") == "" {

		if strings.ToUpper(os.Getenv("SKIPTEARDOWN")) == "" {
			defer func(ctx context.Context, pubCli, privCli *skupper_ocp_smoke.Client) {
				err := teardown(ctx, pubCli, privCli)
				if err != nil {
					fmt.Printf("%s", err.Error())
				}
			}(ctx, pubCli, privCli)
		} else {
			log.Printf("Skipping teardown due to ENV parameter")
		}

		err := runTheJob(ctx, pubCli, pubCli.WaitLimit())
		if err != nil {
			return fmt.Errorf("%s", err.Error())
		}
	}

	// Run the teardown
	if strings.ToUpper(os.Getenv("STEP")) == "TEARDOWN" {
		err := teardown(ctx, pubCli, privCli)
		if err != nil {
			return fmt.Errorf("%s", err.Error())
		}
	}
	return nil
}
