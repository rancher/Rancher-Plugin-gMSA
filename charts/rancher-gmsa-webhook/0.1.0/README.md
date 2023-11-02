# Windows GMSA Admission Webhook

This chart deploys the [Windows GMSA Admission Webhook](https://github.com/kubernetes-sigs/windows-gmsa/tree/master) onto a Windows cluster.

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
