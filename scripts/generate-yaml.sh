#!/bin/bash

# This script can be used to automate the generation of a YAML file to help users
# to run the Skupper Openshift smoke test inside a Kubernetes cluster, as a Job.
#
# Users must answer if they want to use one or two Openshift clusters, and also,
# answer whether they want to enable the debug mode for the tests.
#
# Based on their answers, it will ask for one or two Kubeconfig files.
#
# This/These Kubeconfig file(s) will be used to create a configMap definition and
# this configMap will be then mounted inside the Job definition.
#
# Once the YAML file is created, the user only needs to create the elements defined
# on it, this way :
#
#  kubectl create -f YAML-FILE -n user-namespace
#
# After that, the user can follow the smoke test execution looking into the pod logs, like :
#
#  $  kubectl logs -f -n user-namespace skupper-ocp-smoke-test-pod
#
# In the end of the execution, it will display a XML with the test execution results
#

TEMPCMYAML=$(mktemp -u "/tmp/skupper-ocp-smoke-test-cm.yaml.XXX")
TEMPJOBYAML=$(mktemp -u "/tmp/skupper-ocp-smoke-test.yaml.XXX")
FINALYAML="skupper-ocp-smoke-test.yaml"

#
# Display a message inside a header
#
function message() {
    echo "=================================================="
    echo "=="
    echo "==   $1"
    echo "=="
    echo "=================================================="
}

#
# Display an error  message and exit
#
function errormsg() {
    echo "=="
    echo "== !! ERROR !! - $1"
    echo "=="
    if [ "x${2}" == "xexit" ]; then
        exit 0
    fi
}

#
# Display Question and prompt
#
function question() {
    echo -e "\nQuestion :  $1 ?"
    if [ "x${2}" == "xprompt" ]; then
        echo -n "==> "
    elif [ "x${2}" != "x" ]; then
        echo -n "Options are $2 ==> "
    fi
}

#
# Generate the YAML representation of a config Map based in a Kubeconfig File
#
function generateConfigMapYAML() {

    echo "---" > "${TEMPCMYAML}"

    # One Cluster Only
    if [ "x${2}" == "x" ]; then
        kubectl create configmap config-cm --from-file "${1}" -o yaml --dry-run=client > "${TEMPCMYAML}"
    # Two Clusters
    else
        kubectl create configmap config-cm --from-file "${1}" --from-file "${2}" -o yaml --dry-run=client > "${TEMPCMYAML}"
    fi

    if [ ${?} -ne 0 ]; then
        errormsg "Unable to create configmap. Aborting" "exit"
    fi

    echo -e "\nConfigMap definition created successfully"
}


#
# Generate the YAML representation of a JOB to run the Skupper Openshift smoke tests
#
function generateJobYAML() {

# If DEBUG is set, we comment the QUIET envvar
# But keep it in the YAML, just in case we need to turn it on
DEBUGMODE=""

if [ "x${3^^}" != "xYES" ]; then
DEBUGMODE=$(cat <<EOF

          - name: QUIET
            value: "TRUE"
EOF
)
fi

# Generate the JOB definition
cat <<EOF>> ${TEMPJOBYAML}
apiVersion: batch/v1
kind: Job
metadata:
  labels:
    run: skupper-ocp-smoke-test
  name: skupper-ocp-smoke-test
spec:
  template:
    spec:
      containers:
      - name: skupper-ocp-smoke-test
        image: quay.io/skupper/skupper-ocp-smoke-test-image
        imagePullPolicy: Always
        env:
          - name: PUBKUBECONFIG
            value: /config/${1}
          - name: PRIVKUBECONFIG
            value: /config/${2}
          - name: CLEANBEFORE
            value: "TRUE"${DEBUGMODE}
        volumeMounts:
        - mountPath: /config
          name: config-volume
      volumes:
      - name: config-volume
        configMap:
          name: config-cm
      dnsPolicy: ClusterFirst
      restartPolicy: Never
  backoffLimit: 4
EOF

    echo -e "\nJob definition created successfully"
}


#
# Main
#
message "Generating YAML to use skupper-ocp-smoke-test in OpenShift"

question "Are you going to use one or two clusters"
select NUMCLUSTER in ONE TWO; do [ -n "$NUMCLUSTER" ] && break || echo -e "invalid answer" ; done

question "Enable the DEBUG mode in the final YAML file"
select DEBUG in YES NO; do [ -n "$DEBUG" ] && break || echo -e "invalid answer" ; done

# Both namespaces in one cluster
if [ "x${NUMCLUSTER^^}" == "xONE" ]; then
    KUBEPATH=""
    while [ "x${KUBEPATH}" == "x" ]
    do
        question "What is the location (full or relative path) of the Kubeconfig File" "prompt"
        read KUBEPATH

        # Does the file exists ?
        if [ ! -f "${KUBEPATH}" ]; then
          errormsg "Unable to locate file \"${KUBEPATH}\". Retrying..."
          KUBEPATH=""
        fi
    done

    # Extract the filename, if a path was provided
    KUBEFILE=$(basename -- "${KUBEPATH}")

    generateConfigMapYAML "${KUBEPATH}"
    generateJobYAML "${KUBEFILE}" "${KUBEFILE}" "${DEBUG}"

# Two clusters, one namespace in a distinct cluster
else
    #
    # Public Namespace
    #
    KUBEPATHPUB=""
    while [ "x${KUBEPATHPUB}" == "x" ]
    do
        question "What is the location (full or relative path) of the first Kubeconfig File" "prompt"
        read KUBEPATHPUB

        # Does the file exists ?
        if [ ! -f "${KUBEPATHPUB}" ]; then
            errormsg "Unable to locate file \"${KUBEFILEPUB}\""
            KUBEPATHPRIV=""
        fi
    done

    # Extract the filename, if a path was provided
    KUBEFILEPUB=$(basename -- "${KUBEPATHPUB}")

    #
    # Private Namespace
    #
    KUBEPATHPRIV=""
    while [ "x${KUBEPATHPRIV}" == "x" ]
    do
        question "What is the location (full or relative path) of the second Kubeconfig File" "prompt"
        read KUBEPATHPRIV

        # Does the file exists ?
        if [ ! -f "${KUBEPATHPRIV}" ]; then
            errormsg "Unable to locate file \"${KUBEFILEPRIV}\""
            KUBEPATHPRIV=""
        fi
    done

    # Extract the filename, if a path was provided
    KUBEFILEPRIV=$(basename -- "${KUBEPATHPRIV}")

    generateConfigMapYAML "${KUBEPATHPUB}" "${KUBEPATHPRIV}"
    generateJobYAML "${KUBEFILEPUB}" "${KUBEFILEPRIV}" "${DEBUG}"
fi


echo "---" > ${FINALYAML}
cat "${TEMPCMYAML}" >> ${FINALYAML}
echo "---" >> ${FINALYAML}
cat "${TEMPJOBYAML}" >> ${FINALYAML}

message "YAML ready to use ==>  ${FINALYAML}"
