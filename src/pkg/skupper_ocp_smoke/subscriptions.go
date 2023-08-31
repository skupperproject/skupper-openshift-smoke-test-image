package skupper_ocp_smoke

import (
	"context"
	"fmt"
	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	catalogclient "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
)

func (cli *Client) CreateSubscription(step string) error {

	configResources, err := cli.DiscoveryClient.ServerResourcesForGroupVersion("config.openshift.io/v1")
	if err != nil {
		return fmt.Errorf("unable to access cluster resources - ConfigResources")
	}

	if len(configResources.APIResources) > 0 {
		catalogCli, err := catalogclient.NewForConfig(cli.RestConfig)
		if err != nil {
			return fmt.Errorf("unable to access cluster resources - RestConfig")
		}

		subscriptionSpec := v1alpha1.Subscription{
			TypeMeta: v1.TypeMeta{
				Kind:       "Subscription",
				APIVersion: "operators.coreos.com/v1alpha1",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:      cli.SubscriptionName(),
				Namespace: cli.Namespace,
			},
			Spec: &v1alpha1.SubscriptionSpec{
				CatalogSource:          cli.OperatorCatalog(),
				CatalogSourceNamespace: cli.OperatorNameSpace(),
				Package:                cli.OperatorName(),
				Channel:                "alpha",
				StartingCSV:            cli.StartingCSV(),
				InstallPlanApproval:    "Automatic",
			},
		}
		subscription, err := catalogCli.OperatorsV1alpha1().Subscriptions(cli.Namespace).Create(context.Background(), &subscriptionSpec, v1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("unable to create Subscription = %s", err.Error())
		}
		PrintIfDebug(step, ": Subscription", subscription.Name, "Created on namespace", cli.Namespace)
		return nil
	}
	return fmt.Errorf("unable to create subscription")
}

func (cli *Client) DeleteSubscription(step string) error {

	configResources, err := cli.DiscoveryClient.ServerResourcesForGroupVersion("config.openshift.io/v1")
	if err != nil {
		return fmt.Errorf("unable to access cluster resources - ConfigResources")
	}

	if len(configResources.APIResources) > 0 {
		catalogCli, err := catalogclient.NewForConfig(cli.RestConfig)
		if err != nil {
			return fmt.Errorf("unable to access cluster resources - RestConfig")
		}

		log.Printf("%s : Deleting Subscription", step)

		err = catalogCli.OperatorsV1alpha1().Subscriptions(cli.Namespace).Delete(context.Background(), cli.SubscriptionName(), v1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("unable to Delete Subscription = %s", err.Error())
		}
		PrintIfDebug(step, ": Subscription", cli.SubscriptionName(), "Deleted")
		return nil
	}
	return fmt.Errorf("unable to delete Subscription")
}
