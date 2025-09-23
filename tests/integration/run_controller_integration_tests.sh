#!/bin/bash

go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

if [ -z "$KUBEBUILDER_ASSETS" ];
then
    echo "kube builder assets not explicitly set, running setup-envtest command"
    KUBEBUILDER_ASSETS=$(setup-envtest use --use-env -p path $ENVTEST_K8S_VERSION)
    echo "$KUBEBUILDER_ASSETS"
    export KUBEBUILDER_ASSETS
fi

go test -v $(dirname "$0")/...