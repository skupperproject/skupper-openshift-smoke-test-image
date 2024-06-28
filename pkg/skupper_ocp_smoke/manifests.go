package skupper_ocp_smoke

import (
	"context"
	"fmt"
	packmanifest "github.com/operator-framework/operator-lifecycle-manager/pkg/package-server/client/clientset/versioned"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (cli *Client) IsSkupperOperatorAvailable(step string, operatorName string, opNamespace string, opCatalogSource string) error {

	configResources, err := cli.DiscoveryClient.ServerResourcesForGroupVersion("packages.operators.coreos.com/v1")
	if err != nil {
		return fmt.Errorf("unable to access cluster resources - Config Resources")
	}

	if len(configResources.APIResources) > 0 {
		packManifestCli, err := packmanifest.NewForConfig(cli.RestConfig)
		if err != nil {
			if debug {
				fmt.Printf("DEBUG : Unable to access pack manifest client - %v\n", err)
			}
			return fmt.Errorf("unable to access cluster resources - RestConfig")
		}

		PrintIfDebug(step, " : Looping over manifests")
		for i := 0; i < 5; i++ {
			packManifest, err := packmanifest.Interface.OperatorsV1(packManifestCli).PackageManifests(opNamespace).Get(context.Background(), operatorName, v1.GetOptions{})
			if err != nil {
				if debug {
					fmt.Printf("DEBUG : Unable to access package manifest client - %v\n", err)
				}
				return fmt.Errorf("unable to access cluster resources - OperatorList")
			}

			if packManifest.Status.CatalogSource == opCatalogSource {
				PrintIfDebug(step, " : Operator", operatorName, "available to be installed from ", packManifest.Status.CatalogSource)
				return nil
			} else {
				PrintIfDebug(step, " : Operator", operatorName, "not found. Checking next one")
			}
		}
	}
	return fmt.Errorf("unable to find Skupper Operator ( %s ). Please ensure that it is available on %s catalog", operatorName, opCatalogSource)
}
