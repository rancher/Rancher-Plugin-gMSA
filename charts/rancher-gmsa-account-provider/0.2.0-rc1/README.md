rancher-gmsa-account-provider
========

The Rancher gMSA Account Provider is a **core component** of the Container Credential Guard Rancher Kubernetes Cluster Plugin (CCGRKC Plugin).

It is intended to be deployed onto a cluster whose Windows nodes have already installed the CCGRKC Plugin DLL.

> **Note**: More than one instance of this chart may be installed per cluster, but only one can be installed per namespace. Resources are named according to release name.

## Who needs the Rancher GMSA Account Provider?

Anyone who would like to use the CCGRKC Plugin to deploy container workloads that require Active Directory credentials using the [non-domain-joined gMSA authorization process](https://learn.microsoft.com/en-us/virtualization/windowscontainers/manage-containers/manage-serviceaccounts#gmsa-architecture-and-improvements). 

> **Note**: To install the CCGRKC Plugin DLL, it is recommended for users to deploy the Rancher gMSA Plugin Installer Helm chart onto the cluster.

> **Note**: This chart is **required** for the CCGRKC Plugin to work. Without this chart installed onto your cluster, the CCG Plugin will not be able to access any impersonation account credentials that have been stored on your cluster.

## Before You Install This Chart

### Install `cert-manager` or provide certificates

By default, this chart assumes that [`cert-manager`](https://cert-manager.io/docs/installation/) has already been installed onto the cluster.

> **Note**: Why is `cert-manager` required?
>
> The Account Provider establishes an `mTLS` connection between the CCGRKC DLL (invoked by CCG on the host) and the Account Provider Pod deployed onto the host to communicate the impersonation account credentials through a secure tunnel.
>
> This requires client (CCGRKC DLL) and server (Account Provider) certificates to be created to establish the connection.

#### Using `cert-manager`

If `cert-manager` is used, the following resources are automatically created by this chart:
- A [Self-Signed `Issuer`](https://cert-manager.io/docs/configuration/selfsigned/) (`bootstrap`)
- A CA `Certificate` (`ca-cert`) that uses the "bootstrap" `Issuer`
- A [CA `Issuer`](https://cert-manager.io/docs/configuration/ca/) (`ca`) that uses the `ca-cert` Certificate
- A Client `Certificate` (`client-cert`) that uses the `ca` `Issuer`
- A Server `Certificate` (`server-cert`) that uses the `ca` `Issuer`

> **Note:** Why do we need two `Issuers`?
>
> If we used the `bootstrap` Issuer to directly sign `client-cert` and `server-cert`, they would each be used to **self-sign their own certificate independently** (i.e. the CA certificate for `client-cert` would be `client-cert`).
>
> The expectation of the CCGRKC DLL is that **both certificates must be signed by the same certificate authority**, so we need to make a self-signed certificate first (issued by the `boostrap` Issuer) that corresponds to a single certificate authority (used by the `ca` Issuer) and then use that certificate to sign both the client and server certificates.

After `cert-manager` creates these `Certificates`, it places the contents of the certificate into `Secrets` identified by the following fields in the `values.yaml` of this chart:
- `.Values.certificates.caSecretName`
- `.Values.certificates.serverSecretName`
- `.Values.certificates.clientSecretName`

#### Using provided certificates

If you wanted to externally manage the certificates that are used by this Account Provider, disable `cert-manager` by setting `.Values.certificates.certManager.enabled=false`.

Then, you can externally create [TLS `Secrets`](https://kubernetes.io/docs/concepts/configuration/secret/#tls-secrets) in the release namespace of this chart corresponding to the CA, Client, and Server certificates and provide their names to the following fields:
- `.Values.certificates.caSecretName`
- `.Values.certificates.serverSecretName`
- `.Values.certificates.clientSecretName`

### Create an Impersonation Account Secret

This chart expects a `Secret` to be deployed into the **same namespace** as this chart that contains the following fields:
- `domainName`: the DNS of the Active Directory domain
- `username`: The username of the "impersonation account"
- `password`: The password of the "impersonation account"

> **Note**: Why does it have to be deployed onto the same namespace?
>
> By default, we configure the Account Provider to only watch for Secrets within its own namespace. This makes it easier for a cluster administrator to enforce security policies on impersonation account secrets.
>
> If this plugin is used in a Rancher environment, it is expected that this namespace would live in the System Project.

> **Note**: If you are using [`Rancher Backups`](https://github.com/rancher/backup-restore-operator) and would like your `Secret` to be backed up along with the cluster, make sure you also add the label `resources.cattle.io/backup: "true"` to the `Secret`.

#### What is an impersonation account?

An "impersonation account" is a **user** account in your Active Directory domain that has permissions to retrieve the managed password of the gMSA.

Typically, this is implemented by setting the `PrincipalsAllowedToRetrieveManagedPassword` field on your gMSAs to an "impersonation group" that contains a single entity: the "impersonation account".

```powershell
# The gMSAs you would like to use for workloads in your Kubernetes cluster
$GMSAs = @("gmsa1", "gmsa2")

# An Active Directory group that contains a single user: the "impersonation account"
$ImpersonationGroup = "gmsaImpersonator"

foreach($GMSA in $GMSAs) {
    # This ensures that members of the group can retrieve the gMSAs managed password, which provides permissions to the impersonation account to do so.
    # This impersonation account is assumed by CCG to retrieve and inject the password into your container on initialization.
    Set-AdServiceAccount -Identity $GMSA -PrincipalsAllowedToRetrieveManagedPassword $ImpersonationGroup
}
```

This user account is assumed by Container Credential Guard (CCG) on your Windows host to inject the gMSA's credentials into the Windows container, which is what allows the container to access the Active Directory domain.

The credentials are provided to CCG via a CCG Plugin (i.e. the CCGRKC Plugin), which retrieves it from an external secret store (i.e. Azure KeyVault, the Kubernetes cluster, etc.).

### Create a GMSACredentialSpec

A `GMSACredentialSpec` that was used to deploy the Windows workload will contain a `credspec.ActiveDirectoryConfig.HostAccountConfig.PluginInput` that specifies a `Secret` within the Kubernetes cluster in the format `<secret-namespace>:<secret-name>`.

> **Note**: The following values should also be specified in your `GMSACredentialSpec` under the `HostAccountConfig` to target the CCGRKC Plugin:
> - `PortableCcgVersion`: `"1"`
> - `PluginGUID`: `{e4781092-f116-4b79-b55e-28eb6a224e26}`

## Components

### Install / Upgrade

On install / upgrade, a `DaemonSet` of `HostProcess` containers will be scheduled onto every Windows host in your cluster.

These `HostProcess` containers will run `ccg-account-provider.exe`, which will copy over the mounted certificates to a path on the host and then start up an HTTPs server that only servers `mTLS` connections on a randomly selected host port.

On receiving requests to the `/provider` endpoint, the Account Provider will retrieve the Secret that corresponds to the input encoded in the request and return the credentials. It's expected that this endpoint is only intended to be invoked by the CCGRKC Plugin DLL on behalf of CCG.

### Uninstall

On an uninstall, two workloads will be deployed as Helm [`post-delete` hooks](https://helm.sh/docs/topics/charts_hooks/): a `DaemonSet` of `HostProcess` containers and a `kubectl` Job.

The `HostProcess` containers will run `ccg-account-provider.exe uninstall` as an `initContainer` to uninstall the Account Provider certificates from the Windows host.

After the Account Provider certificates are successfully uninstalled, the main container of each `Pod` will run no more operations (i.e. pause).

On the other hand, the `kubectl` Job will run `kubectl rollout status` on the DaemonSet. Once the `DaemonSet` is rolled out, the hook will allow Helm to continue to remove all resources from the cluster.

> **Note**: Why do we need the `kubectl` Job?
>
> Based on the [Helm docs](https://helm.sh/docs/topics/charts_hooks), Helm will only wait for `Pods` and `Jobs` to complete before marking a hook as completion; all other resources will be marked as complete the moment they are deployed.
>
> So Helm cannot wait for a DaemonSet to rollout to identify when a `post-delete` hook is successful.
>
> Therefore, the `kubectl` Job is used to force Helm to wait for the DaemonSet to "complete" before allowing the Helm chart to be uninstalled from the cluster.

> **Note**: Why is it a `post-delete` Job?
>
> This ensures that the uninstallation process does not encounter a race condition where a node that is added **during the uninstallation of this chart** has both the install and uninstall scripts running on the host at the same time. If this happens, it can have unforeseen side effects.

Once the rollout is completed, the hook will also clean up the `Secrets` left behind by `cert-manager`, since [these Secrets aren't cleaned up by default](https://cert-manager.io/docs/usage/certificate/#cleaning-up-secrets-when-certificates-are-deleted).

If cert-manager is not enabled, these Secrets are expected to be externally managed, so no action will be taken.

## Debugging

### How do I identify whether the Account Provider is working as expected?

Upon installation of the Rancher gMSA Account Provider each Windows Node should have a hostProcess container running which will expose an HTTP/s API on `localhost`. This API server will listen on a port assigned to it by the host. To retrieve the port that the server is listening on, find and view the contents of the `port.txt` file located within the `/var/lib/rancher/gmsa/<ACCOUNT_PROVIDER_NAMESPACE>` directory. The API server will respond to GET requests made against the `/provider` endpoint.

Before you can test the API endpoint, you must ensure that you have an [impersonation account secret](#create-an-impersonation-account-secret) created within the account provider Kubernetes namespace. For example, if you have deployed an Account Provider into the `cattle-windows-workload-system` namespace, you can create a test impersonation account secret using the following command

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: test-gmsa-impersonation-account-secret
  namespace: cattle-windows-workload-system 
type: Opaque
data:
  username: bXktYWQtZG9tYWluLmNvbQ== # username
  password: cGFzc3dvcmQ= # password
  domainName: dXNlcm5hbWU= # my-ad-domain.com
```

Once the secret has been created, you can use the following PowerShell command to test the API endpoint. This example assumes that the account provider has been installed into the `cattle-windows-workload-system` namespace

```powershell
$PORT = $(type C:\var\lib\rancher\gmsa\cattle-windows-workload-system\port.txt)
$PFX_PATH="C:\var\lib\rancher\gmsa\cattle-windows-workload-system\ssl\client\tls.pfx"
Invoke-WebRequest -Method GET -UseBasicParsing -Uri https://localhost:$PORT/provider -Certificate (Get-PfxCertificate $PFX_PATH)  -Headers @{'object' = 'test-gmsa-impersonation-account-secret'}
```

If everything has been configured properly, the command will produce an output similar to the following:

```
StatusCode        : 200
StatusDescription : OK
Content           : {"username":"<username>","password":"<passowrd>","domainName":"<ad-domain-url>"}
RawContent        : HTTP/1.1 200 OK
                    Content-Length: 85
                    Content-Type: application/json; charset=utf-8
                    Date: <redacted>

                    {"username":"<username>","password":"<password>","domainName":"<ad-domain-url>....
Forms             :
Headers           : {[Content-Length, 85], [Content-Type, application/json; charset=utf-8], [Date, <redacted>]}
Images            : {}
InputFields       : {}
Links             : {}
ParsedHtml        :
RawContentLength  : 85
```

If you receive a response like the above, then the account provider API is functioning as expected!

## Certificate Rotation

To rotate the certificates used by the Rancher gMSA Account Provider when certificates are managed by `cert-manager`, users can reference the following script:

```bash
GMSA_SYSTEM_NAMESPACE="cattle-windows-gmsa-system"
CERT_MANAGER_ANNOTATION="cert-manager.io/certificate-name"

# Note:
# It is very important that the "ca-cert" certificate gets deleted **and** recreated before
# you attempt to delete all other certificates. This is because the other certificates depend on
# "ca-cert" existing, so if you delete them at the same time as ca-cert, the CertificateRequest resource
# issued by cert-manager will be stuck.
#
# To fix this, you can simply delete the pending CertificateRequest resource and cert-manager will be
# able to reconcile the secret again.
for cert in "ca-cert" "ccg-dll-cert" "account-provider-cert"; do
    kubectl delete -n $GMSA_SYSTEM_NAMESPACE secret $cert
    while true; do
        echo "Waiting for secret $GMSA_SYSTEM_NAMESPACE/$cert" to be recreated...
        if kubectl -n $GMSA_SYSTEM_NAMESPACE get secret $cert >/dev/null 2>&1; then
            break
        fi
        sleep 2
    done
done

echo "Restarting gMSA Account Provider..."
kubectl -n $GMSA_SYSTEM_NAMESPACE rollout restart daemonset/rancher-gmsa-account-provider

# Note:
# You do not need to restart the CCG DLL Installer or any gMSA workload pods when you update
# the gMSA Account Provider's certificates. Once you redeploy the gMSA Account Provider onto each host,
# the Account Provider will automatically update the hostPath certificates, which will be what is referenced
# by the DLL (and subsequently all gMSA workload pods) when they attempt to retrieve the certificates.
```

If you brought your own certificates, simply update the relevant `Secret` objects in your cluster and run the final `kubectl rollout restart` command in the above script to achieve the same result.

## License
Copyright (c) 2023 [Rancher Labs, Inc.](http://rancher.com)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
