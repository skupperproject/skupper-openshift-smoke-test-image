package skupper_ocp_smoke

import (
	"context"
	"fmt"
	v12 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"strconv"
	"strings"
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

// Receives a default value, that will be overridden when running on Openshift.
// Returns the value to be used.
//
// OpenShift requires RunAsUser to be configured according to annotations
// present on the namespace, lest it will fail with SCC errors.
//
// If any errors found while tring to determine the correct user, this function
// will ignore them and simply use the default value.
func (cli *Client) GetRunAsUserOrDefault(runAsUser string, cctx context.Context) string {

	// OpenShift requires container user IDs to exist within a range; we try to satisfy it here.
	namespace, err := cli.KubeClient.CoreV1().Namespaces().Get(cctx, cli.Namespace, v1.GetOptions{})
	if err != nil {
		log.Printf("Unable to get namespace %q; using pre-defined runAsUser value %v", cli.Namespace, runAsUser)
	} else {
		ns_annotations := namespace.GetAnnotations()
		if users, ok := ns_annotations["openshift.io/sa.scc.uid-range"]; ok {
			log.Printf("OpenShift UID range annotation found: %q", users)
			// format is like 1000860000/10000, where the first number is the
			// range start, and the second its length
			split_users := strings.Split(users, "/")
			if split_users[0] != "" {
				if _, err := strconv.Atoi(split_users[0]); err == nil {
					runAsUser = split_users[0]
				} else {
					log.Printf("Failed to parse openshift uid-range annotation: using default value %v", runAsUser)
				}
			} else {
				log.Printf("openshift uid-range annotation is empty, which is unexpected: using default value %v", runAsUser)
			}
		} // if annotation not found, we're not on Openshift, and we can use the default value.
	}
	return runAsUser
}