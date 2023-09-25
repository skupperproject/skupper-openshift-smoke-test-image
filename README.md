# Skupper Openshift Smoke Test image

  This **Skupper** Openshift Smoke Test image is designed to be used in **Red Hat OpenShift (OCP)**, once it uses the Skupper operator from Openshift Marketplace.

  It creates two Skupper instances, in the namespaces 'pub-test-ns' and 'priv-test-ns'. These namespaces can be created in a unique OCP cluster or each namespace can be created in a specific cluster.

  After the Skupper instances are up, the test image connects them, setup a simple web server in the priv-test-ns namespace, and then access this service through Skupper in the pub-test-ns namespace.

<br>

## Basic usage :

**docker**:

    docker run -v /home/localuser/kubeconfigs/:/config  -it --env PUBKUBECONFIG=/config/config-ocp49 --env PRIVKUBECONFIG=/config/config-ocp410 --env QUIET=TRUE quay.io/skupper/skupper-ocp-smoke-test-image

<br>
<br>

**podman**:

    podman run -v /home/localuser/kubeconfigs/:/config  -it --env PUBKUBECONFIG=/config/config-ocp49 --env PRIVKUBECONFIG=/config/config-ocp410 --env QUIET=TRUE quay.io/skupper/skupper-ocp-smoke-test-image

<br>
<br>

**Openshift**:

    oc create -f skupper-ocp-smoke-test.yaml -n [namespace]

<br>

Openshift
==========

 Use the script generate-yaml.sh to generate the YAML file to run the Skupper Openshift smoke tests.

 You can use one or two clusters.
 If you want to use only one cluster, you must provide the path to the kubeconfig file, and both namespaces will be created in that cluster.
 If you want to use two clusters, you must provide two kubeconfig files.

 Based on the provided kubeconfig file(s), the script create a configMap entry based on the file, and then will setup a job using that configMap as a parameter to the job

 Once you apply the yaml, you can check the logs in the skupper-ocp-smoke-test POD, like this one :

    # oc logs skupper-ocp-smoke-test-k8d7v -n my-skupper-ocp-smoke-test -f

    === RUN   TestSkupperOCPSmoke
    2023/02/03 19:38:28 main : Starting main test
    2023/02/03 19:38:28 main : Debug detail
    2023/02/03 19:38:28 main : Get Kubeconfig settings
    2023/02/03 19:38:28 main : Kubeconfig for public namespace =  /config/config-ocp410
    2023/02/03 19:38:28 main : Kubeconfig for private namespace =  /config/config-ocp411
    2023/02/03 19:38:28 setup : Starting Setup
    ...
    2023/02/03 19:40:04 teardown : Waiting 5 seconds until namespace get removed
    2023/02/03 19:40:04 teardown : Namespace priv-test-ns removed. Moving on
    2023/02/03 19:40:04 teardown : Teardown finished
    --- PASS: TestSkupperOCPSmoke (110.43s)
    PASS
    <?xml version="1.0" encoding="UTF-8"?>
    <testsuites>
        <testsuite tests="1" failures="0" time="96.420" name="">
            <properties>
                <property name="go.version" value="go1.19.12"></property>
            </properties>
            <testcase classname="" name="TestSkupperOCPSmoke" time="96.420"></testcase>
        </testsuite>
    </testsuites>

Please note that in the end of the POD logs, you can see the results in XML format.

**Important:** There is a set of configurations/options that can be used to customize the test behavior. You can also add them to the created YAML, if necessary.


Docker / Podman
================

## How to configure access rights to the cluster(s) :

  This image uses kubeconfig files hosted into the /config directory, and these files must provide credentials to access the cluster as **admin**.

  You need to provide the kubeconfig mapping a local directory to the /config directory inside the image, files this way :

    "-v /home/localuser/kubeconfigs/:/config"

  <br>  By default, it uses a kubeconfig file named **config**, for both namespaces.

  If you are using a kubeconfig file with a different name, you must specify it using the "--env" parameter, like this :

    "--env PUBKUBECONFIG=/config/config-ocp-4-9 --env PRIVKUBECONFIG=/config/config-ocp-4-9"
    ** In this case, we are using a unique cluster, note that both environment variables
       points to the same kubeconfig file

    "--env PUBKUBECONFIG=/config/config-ocp-4-9 --env PRIVKUBECONFIG=/config/config-ocp-4-10"
    ** In this case, we are using different clusters, note that each environment variables
       points to distinct kubeconfig file

  <br>  This test image creates some debug messages by default, and if you want to disable them, you must set the QUIET environment variable :

    "--env QUIET=TRUE"

  <br>  At the end of the test execution, it displays the output messages and a junit file. This junit file is also available inside the container in the /result folder.

  You can then retrieve it from the container :

**docker**:    "docker cp [CONTAINER_NAME]:/result/junit.xml /home/localuser/junits/"
<br>
<br>
**podman**:    "podman cp [CONTAINER_NAME]:/result/junit.xml /home/localuser/junits/"

  or you can map the /results from the image to a local directory, adding to the command line the sentence below :

    "-v /tmp/local-test-results/:/result"

<br>

## Advanced configuration

  The test runs in three steps : Setup, Test and Teardown. If you don't specify a STEP, the test will go through all of them.

  Usually you don't need to specify a step, but if you need, you can do it using the STEP environment variable :

    "-- env STEP=SETUP"
    ** This will create the namespaces, deploy the Skupper instances and the 
       services, but will not validate the test. It can be used for debugging
       purposes.

    "-- env STEP=RUNTEST"
    ** This will run the test, using the scenario from the SETUP step, and then
       it will run the TEARDOWN step, removing all the created resources from 
       the cluster.

    "-- env SKIPTEARDOWN=TRUE"
    ** This will skip the teardown phase, if you are running the RUNTEST step
       or the full test ( not specifying a STEP )
    ** This can be useful for debugging purposes

    "-- env STEP=TEARDOWN"
    ** This will remove all the created resources from the cluster.

    "-- env STARTINGCSV=1.2.3.4"
    ** By default, the test installs the latest operator from the Operator Hub.
       If you need to test a specific version, you can specify the version.

    "-- env CLEANBEFORE=TRUE"
    ** Use this environment variable to force the test setup to remove the 
       namespace used in the test before it starts, usually to avoid any 
       conflict with previous test runs.
    ** This options affects only the step SETUP or the full test run 
       ( without any STEP specified )

    "-- env WAITLIMIT=180"
    ** By default, the test waits until 120 seconds for the operations to get
       completed, like the link creation or the test to run.
       If you need more time, you can specify it using this variable. It can be
       useful when using multiple clusters.
