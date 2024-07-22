package skupper_ocp_smoke

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
)

type Client struct {
	Namespace       string
	KubeClient      *kubernetes.Clientset
	RestConfig      *rest.Config
	DiscoveryClient *discovery.DiscoveryClient
	debug           bool
}

func NewClient(context string, kubeConfigPath string, namespace string) (*Client, error) {

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if kubeConfigPath != "" {
		loadingRules = &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeConfigPath}
	}
	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		&clientcmd.ConfigOverrides{
			CurrentContext: context,
		},
	)
	restconfig, err := kubeconfig.ClientConfig()
	if err != nil {
		return nil, err
	}
	c := &Client{}
	if os.Getenv("DEBUG") != "" {
		c.debug = true
	}
	restconfig.ContentConfig.GroupVersion = &schema.GroupVersion{Version: "v1"}
	restconfig.APIPath = "/api"
	restconfig.NegotiatedSerializer = serializer.WithoutConversionCodecFactory{CodecFactory: scheme.Codecs}
	c.RestConfig = restconfig
	c.KubeClient, err = kubernetes.NewForConfig(restconfig)
	if err != nil {
		return c, err
	}

	c.DiscoveryClient, err = discovery.NewDiscoveryClientForConfig(c.RestConfig)

	c.Namespace = namespace

	if err != nil {
		return nil, err
	}
	return c, nil
}

func (cli *Client) OperatorGroupName() string {
	return operatorgroupname
}

func (cli *Client) StartingCSV() string {
	return Startingcsv
}

func (cli *Client) SubscriptionName() string {
	return subscriptionname
}

func (cli *Client) Channel() string {
	return channel
}

func (cli *Client) WaitLimit() int {
	return Waitlimit
}

func (cli *Client) OperatorName() string {
	return operatorname
}

func (cli *Client) OperatorNameSpace() string {
	return operatornamespace
}

func (cli *Client) OperatorCatalog() string {
	return operatorcatalog
}

func (cli *Client) SiteConfigRunAsUserKey() string {
	return SiteConfigRunAsUserKey
}

func (cli *Client) SiteConfigRunAsGroupKey() string {
	return SiteConfigRunAsGroupKey
}

func (cli *Client) SiteConfigRunAsUser() string {
	return SiteConfigRunAsUser
}

func (cli *Client) SiteConfigRunAsGroup() string {
	return SiteConfigRunAsGroup
}
