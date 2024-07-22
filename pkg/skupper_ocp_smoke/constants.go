package skupper_ocp_smoke

import (
	"os"
	"strconv"
)

const (
	PUBKUBECONFIGFILE       string = "/config/kubeconfig"
	PRIVKUBECONFIGFILE      string = "/config/kubeconfig"
	PUBNS                   string = "pub-test-ns"
	PRIVNS                  string = "priv-test-ns"
	OPERATORGROUP           string = "test-operator-group"
	STARTINGCSV             string = ""
	CHANNEL                 string = "stable"
	SUBSCRIPTION            string = "skupper-operator"
	WAITLIMIT               int    = 120
	OPERATORNAME            string = "skupper-operator"
	OPERATORNAMESPACE       string = "openshift-marketplace"
	OPERATORCATALOG         string = "redhat-operators"
	SITECONFIGRUNASUSERKEY  string = "run-as-user"
	SITECONFIGRUNASGROUPKEY string = "run-as-group"
	SITECONFIGRUNASUSER     string = "1000"
	SITECONFIGRUNASGROUP    string = "2000"
)

var (
	debug                   = os.Getenv("QUIET") == ""
	operatorgroupname       = StrDefault(os.Getenv("OPERATORGROUP"), OPERATORGROUP)
	Startingcsv             = StrDefault(os.Getenv("STARTINGCSV"), STARTINGCSV)
	subscriptionname        = StrDefault(os.Getenv("SUBSCRIPTION"), SUBSCRIPTION)
	channel                 = StrDefault(os.Getenv("CHANNEL"), CHANNEL)
	Waitlimit               = IntDefault(os.Getenv("WAITLIMIT"), WAITLIMIT)
	operatorname            = StrDefault(os.Getenv("OPERATORNAME"), OPERATORNAME)
	operatornamespace       = StrDefault(os.Getenv("OPERATORNAMESPACE"), OPERATORNAMESPACE)
	operatorcatalog         = StrDefault(os.Getenv("OPERATORCATALOG"), OPERATORCATALOG)
	SiteConfigRunAsUserKey  = StrDefault(os.Getenv("SITECONFIGRUNASUSERKEY"), SITECONFIGRUNASUSERKEY)
	SiteConfigRunAsGroupKey = StrDefault(os.Getenv("SITECONFIGRUNASGROUPKEY"), SITECONFIGRUNASGROUPKEY)
	SiteConfigRunAsUser     = StrDefault(os.Getenv("SITECONFIGRUNASUSER"), SITECONFIGRUNASUSER)
	SiteConfigRunAsGroup    = StrDefault(os.Getenv("SITECONFIGRUNASGROUP"), SITECONFIGRUNASGROUP)
)

func StrDefault(val string, dflt string) string {
	if val != "" {
		return val
	}
	return dflt
}

func IntDefault(val string, dflt int) int {
	var ret int

	if val != "" {
		return dflt
	}
	ret, err := strconv.Atoi(val)
	if err != nil {
		return WAITLIMIT
	}
	return ret
}
