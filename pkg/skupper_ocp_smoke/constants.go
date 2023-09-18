package skupper_ocp_smoke

import (
	"os"
	"strconv"
)

const (
	PUBKUBECONFIGFILE  = "/config/kubeconfig"
	PRIVKUBECONFIGFILE = "/config/kubeconfig"
	PUBNS              = "pub-test-ns"
	PRIVNS             = "priv-test-ns"
	OPERATORGROUP      = "test-operator-group"
	STARTINGCSV        = ""
	SUBSCRIPTION       = "skupper-operator"
	WAITLIMIT          = 120
	OPERATORNAME       = "skupper-operator"
	OPERATORNAMESPACE  = "openshift-marketplace"
	OPERATORCATALOG    = "redhat-operators"
)

var (
	debug             = os.Getenv("QUIET") == ""
	operatorgroupname = StrDefault(os.Getenv("OPERATORGROUP"), OPERATORGROUP)
	Startingcsv       = StrDefault(os.Getenv("STARTINGCSV"), STARTINGCSV)
	subscriptionname  = StrDefault(os.Getenv("SUBSCRIPTION"), SUBSCRIPTION)
	Waitlimit         = IntDefault(os.Getenv("WAITLIMIT"), WAITLIMIT)
	operatorname      = StrDefault(os.Getenv("OPERATORNAME"), OPERATORNAME)
	operatornamespace = StrDefault(os.Getenv("OPERATORNAMESPACE"), OPERATORNAMESPACE)
	operatorcatalog   = StrDefault(os.Getenv("OPERATORCATALOG"), OPERATORCATALOG)
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
