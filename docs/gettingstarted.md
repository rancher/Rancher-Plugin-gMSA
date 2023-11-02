# Getting Started

Installing Rancher gMSA CCG Plugin requires the installation of two Helm charts onto your Windows cluster:

1. [`rancher-gmsa-plugin-installer`](../charts/rancher-gmsa-plugin-installer): install the CCG Plugin onto each Windows host in your cluster
2. [`rancher-gmsa-account-provider`](../charts/rancher-gmsa-account-provider): serves as a local proxy on each host for the CCG Plugin to grab a Secret from the Kubernetes cluster

Once installed, you may also want to install [`rancher-windows-gmsa`](../charts/rancher-windows-gmsa) (and `rancher-windows-gmsa-crd`), which adds [admission webhooks](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/) on Pods to fill out gMSA credentials based on configuration stored in a `GMSACredentialSpec`. This will allow your Pods to specify that they should run as a specific gMSA account.

## Prerequisites

- A Kubernetes cluster with Windows nodes
  - Must be 1.24+
  - Must come with [`cert-manager`](https://cert-manager.io/docs/installation) and its CRDs installed
  - Each node must be running [`containerd`](https://containerd.io) 1.7+ as the runtime
  - All Windows hosts must be running `Windows Server 2019` or `Windows Server 2022`. No other versions are supported.
- An Active Directory instance
  - Expected to be on the same local network as your Windows nodes (or in a [peered network](https://en.wikipedia.org/wiki/Peering))
  - Expected to contain one or more Active Directory gMSAs that are tied to a single user ("impersonation") account that is allowed to retrieve each gMSA's managed password

## Simple Installation

### In Rancher (via Apps & Marketplace)

1. Ensure `cert-manager` has already been installed onto your Windows cluster
2. Navigate to `Apps & Marketplace -> Repositories` in your target downstream cluster and create a Repository that points to a `Git repository containing Helm chart or cluster template definitions` where the `Git Repo URL` is `https://github.com/rancher/Rancher-Plugin-gMSA` and the `Git Branch` is `main`
3. Navigate to `Apps & Marketplace -> Charts`; you should see three charts under the Repository you created:  `Windows GMSA`, `Rancher gMSA CCG Plugin`, and `Rancher gMSA Account Provider`.
4. Install `Rancher gMSA CCG Plugin` 
5. Install `Rancher gMSA Account Provider`

> **Note**: You can also install `Windows GMSA` from this new Repository.

### In a normal Kubernetes cluster (via running Helm 3 locally)

1. Install `rancher-gmsa-plugin-installer` onto your cluster via Helm to install the CCG Plugin DLL onto your Windows worker nodes

```
helm install --create-namespace -n cattle-windows-gmsa-system rancher-gmsa-plugin-installer charts/rancher-gmsa-plugin-installer
```

2. Install `rancher-gmsa-account-provider` chart to deploy the account provider API onto your Windows worker nodes.  

```bash
helm install --create-namespace -n cattle-windows-gmsa-system rancher-gmsa-account-provider charts/rancher-gmsa-account-provider
```

You can also install `rancher-windows-gmsa` after the fact:

```bash
helm install --create-namespace -n cattle-windows-gmsa-system rancher-gmsa-webhook-crd charts/rancher-gmsa-webhook-crd
helm install --create-namespace -n cattle-windows-gmsa-system rancher-gmsa-webhook charts/rancher-gmsa-webhook
```

### Account Configuration

After installation, two resources need to be properly configured to leverage a gMSA account on a non-domain joined nodes per Active Directory Domain you would like to add to this cluster.

1. A Kubernetes secret that contains "impersonation" account credentials
2. A `GMSACredentialSpec` for **each** gMSA that your "impersonation" account can retrieve the managed password for

The Kubernetes secret will contain the credentials required to authorize with Active Directory on behalf of the container. The user specified within this secret must have permissions within Active Directory to retrieve the managed password of each gMSA in your domain that will be used by a workload.

Once a user is properly configured, a simple secret can be created to outline the username and password of the account, as well as the domain which the account exists in. This secret must be created in the same namespace as an instance of the Account Provider API.  

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: rancher-domain-user-credentials 
  namespace: rancher-domain-account-provider-namespace
type: Opaque
data:
  # This is just an example, in reality all data fields should be base64 encoded
  username: test-username
  password: Password@123!
  domainName: rancher.ad.com
```

The `GMSACredentialSpec` is used to define the Active Directory Domain and gMSA account that a workload should utilize. It also provides input to the Rancher CCG Plugin as to what user account should be used to authenticate with Active Directory. In a single domain environment, you may only need a single user account secret, but will need a number of `GMSACredentialSpec`'s depending on the number of gMSA accounts. 

Here's an example of a `GMSACredentialSpec` which targets a gMSA account named `GMSA1` in an Active Directory Domain titled `rancher.ad.com`. This example utilizes the Rancher CCG Plugin to perform the Active Directory authorization required to obtain the gMSA account password, through the values supplied in the `HostAccountConfig` field. 

```yaml
apiVersion: windows.k8s.io/v1
kind: GMSACredentialSpec
metadata:
  name: example-GMSA1-credspec 
credspec:
  ActiveDirectoryConfig:
    HostAccountConfig: # This section is required for non-domain-joined nodes
      PluginGUID: '{e4781092-f116-4b79-b55e-28eb6a224e26}' # This field indicates that the Rancher CCG Plugin should be used
      PluginInput: 'rancher-domain-account-provider-namespace:rancher-domain-user-credentials' # <ACCOUNT_PROVIDER_NAMESPACE>:<ACCOUNT_CREDENTIAL_SECRET_NAME> 
      PortableCcgVersion: '1' # This must always be '1', until a new CCG version is released 
    GroupManagedServiceAccounts:
    - Name: GMSA1        # Username of the GMSA account
      Scope: RANCHER     # NETBIOS Domain Name
    - Name: GMSA1        # Username of the GMSA account
      Scope: rancher.ad.com # DNS Domain Name
  CmsPlugins:
  - ActiveDirectory
  DomainJoinConfig:
    DnsName: rancher.ad.com     # DNS Domain Name
    DnsTreeName: rancher.ad.com # DNS Domain Name Root
    Guid: 244818ae-87ac-4fcd-92ec-e79e5252348a  # GUID Of the Active Directory Domain
    MachineAccountName: GMSA1 # Username of the GMSA account
    NetBiosName: RANCHER      # NETBIOS Domain Name
    Sid: S-1-5-21-2126449477-2524075714-3094792973 # SID of the Active Directory Domain
```

Once you've created the `GMSACredentialSpec` (as well as any RBAC required to utilize the object), you can start to create gMSA workloads. Each workload which intends to utilize a gMSA account must include a value within the pods security context. For example, if we wanted to utilize the example `GMSACredentialSpec` we've just created within a deployment, the pod security context would need to look like this:
```yaml 
securityContext:
  windowsOptions:
    gmsaCredentialSpecName: example-GMSA1-credspec
```
The gMSA web-hook will automatically expand this reference onto all pods rolled-out, allowing the gMSA authorization process to occur. 

Assuming everything has been installed and configured correctly for your desired Active Directory domain, the workload should deploy as normal and enjoy all the permissions given to the gMSA account.  

#### Verifying Account Configuration

If you are running into a problem with this setup process and would like an example of what resources need to be created for your cluster, you can **reference** the [`windows-ad-setup`](https://github.com/rancher/windows/blob/main/charts/windows-ad-setup) chart in the `rancher/windows` repository that is used by the Rancher Windows team for testing this feature.

However, please note the following disclaimers:

1. **This chart will never be supported by the Rancher Windows team in any capacity and is subject for breaking changes or removal at any point of time**.
2. It is **never** recommended to deploy Secrets (like your "impersonation" account credentials) using a Helm chart in a production setup.

```bash
# From the output of Get-ADDomain
AD_DNS_ROOT="rancher-ad.ad.com"
AD_FOREST="rancher-ad.ad.com"
AD_NETBIOS_NAME="ad"
AD_GUID="244818ae-87ac-4fcd-92ec-e79e5252348a"
AD_SID="S-1-5-21-2126449477-2524075714-3094792973"

# Admin created
IMPERSONATION_ACCOUNT_USER="test-username"
IMPERSONATION_ACCOUNT_PASSWORD='Password@123!'
GMSAS="gmsa1,gmsa2"

AD_DOMAIN=$(cat <<EOF
{
  "DNSRoot": "$AD_DNS_ROOT",
  "Forest": "$AD_FOREST",
  "NetBIOSName": "$AD_NETBIOS_NAME",
  "ObjectGUID": "$AD_GUID",
  "SID": "$AD_SID"
}
EOF
)

helm template -n cattle-windows-gmsa-system windows-ad-setup \
  --set-json activeDirectory.domain="$AD_DOMAIN" \
  --set activeDirectory.ccg.impersonationAccount.username="$IMPERSONATION_ACCOUNT_USER" \
  --set activeDirectory.ccg.impersonationAccount.password="$IMPERSONATION_ACCOUNT_PASSWORD" \
  --set activeDirectory.gmsas="{$GMSAS}" \
  charts/windows-ad-setup/
```

### Checking if the Rancher gMSA CCG Plugin Works

1. Ensure that the init container deployed with each pod of the `rancher-gmsa-plugin-installer` chart successfully completes and does not log any errors.
2. Ensure that all pods spawned from the `rancher-gmsa-account-provider` deamon set have started properly and do not log any errors 
3. Deploy a Windows workload which leverages a gMSA account, ensure that it successfully transitions to 'Running'

## Uninstalling the Rancher gMSA CCG Plugin

After deleting the Helm Charts, all artifacts and certificates written to the host will be removed by a `Job`. Ensure this `Job` successfully completes.
