package skupper_ocp_smoke

import (
	"context"
	"fmt"
	v12 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

func (cli *Client) WaitForNSDeletion(ctx context.Context, limit int, step string) error {

	timeused := 0
	for {

		if cli.NameSpaceExists(ctx) == false {
			PrintIfDebug(step, ": Namespace", cli.Namespace, "removed. Moving on")
			return nil
		}

		time.Sleep(5 * time.Second)
		PrintIfDebug(step, ": Waiting 5 seconds until namespace get removed")
		timeused += 5

		if timeused >= limit {
			return fmt.Errorf("time limit achieved waiting for namespace %s to be removed", cli.Namespace)
		}
	}
}

func (cli *Client) NameSpaceExists(ctx context.Context) bool {

	_, err := cli.KubeClient.CoreV1().Namespaces().Get(ctx, cli.Namespace, v1.GetOptions{})
	if err != nil {
		return false
	}
	return true
}

func (cli *Client) DeleteNamespace(ctx context.Context, step string) error {
	if cli.NameSpaceExists(ctx) == false {
		PrintIfDebug(step, ": Namespace", cli.Namespace, "DOES NOT Exists. Skipping deletion")
		return nil
	}

	PrintIfDebug(step, ": Namespace", cli.Namespace, "Exists. Trying to remove it")
	err := cli.KubeClient.CoreV1().Namespaces().Delete(ctx, cli.Namespace, v1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("unable to delete namespace %s : %s", cli.Namespace, err.Error())
	}
	PrintIfDebug(step, ": Namespace", cli.Namespace, "removed")
	return nil
}

func (cli *Client) CreateNamespace(ctx context.Context, step string) error {
	if cli.NameSpaceExists(ctx) == true {
		PrintIfDebug(step, ": Namespace", cli.Namespace, "already exists")
	} else {
		PrintIfDebug(step, ": Namespace", cli.Namespace, "DOES NOT exists. Creating")
		nsSpec := v12.Namespace{ObjectMeta: v1.ObjectMeta{Name: cli.Namespace}}
		ns, err := cli.KubeClient.CoreV1().Namespaces().Create(ctx, &nsSpec, v1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("unable to create namespace %s - %s", cli.Namespace, err.Error())
		}
		PrintIfDebug(step, ": Namespace", ns.Name, "created")
	}
	return nil
}
