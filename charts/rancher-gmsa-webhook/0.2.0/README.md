# Windows gMSA Admission Webhook

This chart deploys the [Windows gMSA Admission Webhook](https://github.com/kubernetes-sigs/windows-gmsa/tree/master) onto a Windows cluster.

This admission webhook allows workloads in the cluster to specify `securityContext.windowsOptions.gmsaCredentialSpecName` to target an existing `GMSACredentialSpec` CRD that the workload's `ServiceAccount` is permitted to `use`.
If this requirement is met, the webhook will populate `securityContext.windowsOptions.gmsaCredentialSpec` with the contents of the `GMSACredentialSpec` resource.


The official documentation and tutorials can be found [here](https://github.com/kubernetes-sigs/windows-gmsa).
> **Note**: If `credential.enabled` is set to `true`, a default `GMSACredentialSpec` that can be used by workloads will also be created.
>
> However, workloads will need to introduce their own RBAC that gives their `ServiceAccount` permissions to `use` this `GMSACredentialSpec`.

## Prerequisites

- Active Directory that supports Group Managed Service Accounts
- A Group Managed Service Account
- Kubernetes v1.21+

## Setting up workloads to use a `GMSACredentialSpec`

In order to allow workloads to assume gMSAs, users must create a `Role` for each `GMSACredentialSpec` permitting the `use` verb on that resource.
Then, for each workload that uses that `GMSACredentialSpec` in the same namespace, users must create a `RoleBinding` allowing the workload's service account to be bound to that role.
> **Note**: If a service account is not specified on the workload, the `RoleBinding` should target the `ServiceAccount` named **default** in the workload's namespace.

## Certificate Rotation

To rotate the certificates used by the Windows gMSA Admission Webhook when certificates are managed by `cert-manager`, users can reference the following script:

```bash
GMSA_SYSTEM_NAMESPACE="cattle-windows-gmsa-system"
CERT_MANAGER_ANNOTATION="cert-manager.io/certificate-name"

for cert in "gmsa-server-cert"; do
    kubectl delete -n $GMSA_SYSTEM_NAMESPACE secret $cert
    while true; do
        echo "Waiting for secret $GMSA_SYSTEM_NAMESPACE/$cert" to be recreated...
        if kubectl -n $GMSA_SYSTEM_NAMESPACE get secret $cert >/dev/null 2>&1; then
            break
        fi
        sleep 2
    done
done

echo "Restarting gMSA Webhook..."
kubectl -n $GMSA_SYSTEM_NAMESPACE rollout restart deployment/rancher-gmsa-webhook
```

If you brought your own certificates, simply update the relevant `Secret` object in your cluster and run the final `kubectl rollout restart` command in the above script to achieve the same result.
