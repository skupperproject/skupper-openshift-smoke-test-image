package skupper_ocp_smoke

import (
	"context"
	"fmt"
	v12 "github.com/operator-framework/api/pkg/operators/v1"
	operatorclient "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (cli *Client) CreateOperatorGroup(step string) error {

	configResources, err := cli.DiscoveryClient.ServerResourcesForGroupVersion("config.openshift.io/v1")
	if err != nil {
		return fmt.Errorf("unable to access cluster resources - ConfigResources")
	}

	if len(configResources.APIResources) > 0 {
		operatorCli, err := operatorclient.NewForConfig(cli.RestConfig)
		if err != nil {
			return fmt.Errorf("unable to access cluster resources - RestConfig")
		}

		PrintIfDebug(step, ": Creating Operator Group")
		opGrpDefinition := v12.OperatorGroup{
			TypeMeta: v1.TypeMeta{
				Kind:       "OperatorGroup",
				APIVersion: "operators.coreos.com/v1",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:      cli.OperatorGroupName(),
				Namespace: cli.Namespace,
			},
			Spec: v12.OperatorGroupSpec{
				TargetNamespaces: []string{cli.Namespace},
			},
		}

		opGroup, err := operatorCli.OperatorsV1().OperatorGroups(cli.Namespace).Create(context.Background(), &opGrpDefinition, v1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("unable to create Operator Group = %s", err.Error())
		}
		PrintIfDebug(step, ": Operator Group", opGroup.Name, "Created on namespace", cli.Namespace)
		return nil
	}
	return fmt.Errorf("unable to create OperatorGroup")
}

func (cli *Client) DeleteOperatorGroup(step string) error {

	configResources, err := cli.DiscoveryClient.ServerResourcesForGroupVersion("config.openshift.io/v1")
	if err != nil {
		return fmt.Errorf("unable to access cluster resources - ConfigResources")
	}

	if len(configResources.APIResources) > 0 {
		operatorCli, err := operatorclient.NewForConfig(cli.RestConfig)
		if err != nil {
			return fmt.Errorf("unable to access cluster resources - RestConfig")
		}

		PrintIfDebug(step, ": Deleting Operator Group")

		err = operatorCli.OperatorsV1().OperatorGroups(cli.Namespace).Delete(context.Background(), cli.OperatorGroupName(), v1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("unable to create Operator Group = %s", err.Error())
		}
		PrintIfDebug(step, ": Operator Group", cli.OperatorGroupName(), "Created")
		return nil
	}
	return fmt.Errorf("unable to create OperatorGroup")
}
