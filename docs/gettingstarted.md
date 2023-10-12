# Getting Started

The Rancher CCG gMSA Plugin feature comprises two components, the plugin installer (which installs the CCG plugin), and the account provider (which the plugin uses to obtain login credentials for a domain). Installing both Helm charts is required in order for a cluster to support domainless GMSA.  

## Prerequisites 

+ Kubernetes v1.24+, containerD 1.7+
+ Nodes running `Windows Server 2019`, `Windows Server 2022`, `Windows Server Core 2019`, or `Windows Server Core 2022 
`+ Upstream gMSA Web-hook chart installed ([Official Repo](https://github.com/kubernetes-sigs/windows-gmsa), [Rancher Specific Chart](https://github.com/rancher/charts/tree/release-v2.7/charts/rancher-windows-gmsa))\
  + **Note:** Currently, the gMSA Web-Hook chart offered by Rancher **does not contain the required changes to support non-domain joined nodes**. This chart will be updated soon, in the meantime you should use the following repository `harrisonwaffel/charts` and test against the `update-gmsa` branch (this can be done by configuring a new app repository in the Rancher UI).
+ An Active Directory instance which is reachable by worker nodes
+ One or more Active Directory gMSA accounts 

## Simple Installation

### In Rancher (via Apps & Marketplace)

1. Install the latest version of `cert-manager` onto your cluster
2. Navigate to `Apps & Marketplace -> Repositories` in your target downstream cluster and create a Repository that points to a `Git repository containing Helm chart or cluster template definitions` where the `Git Repo URL` is `https://github.com/rancher/Rancher-Plugin-gMSA` and the `Git Branch` is `main`
3. Navigate to `Apps & Marketplace -> Repositories` in your target downstream cluster and create a Repository that points to a `Git repository containing Helm chart or cluster template definitions` where the `Git Repo URL` is `https://github.com/harrisonwaffel/charts` and the `Git Branch` is `update-gmsa`
4. Navigate to `Apps & Marketplace -> Charts`; you should see three charts under the two new Repositories you created:  `Windows GMSA`, `Rancher gMSA CCG Plugin`, and `Rancher gMSA Account Provider`.
5. Install `Windows GMSA` and configure a `GMSACredentialSpec`
6. Install `Rancher gMSA CCG Plugin` 
7. Install `Rancher gMSA Account Provider` 

### In a normal Kubernetes cluster (via running Helm 3 locally)

1. Install `rancher-gmsa-plugin-installer` onto your cluster via Helm to install the CCG Plugin DLL onto your Windows worker nodes

```
helm install -n cattle-gmsa-system Rancher-Plugin-gMSA charts/rancher-gmsa-plugin-installer
```

2. Install `rancher-gmsa-account-provider` chart to deploy the account provider API onto your Windows worker nodes.  

```bash
helm install -n cattle-helm-system Rancher-Plugin-gMSA-account-provider charts/rancher-gmsa-account-provider
```

### Account Configuration

After installation, two resources need to be properly configured to leverage a gMSA account on a non-domain joined nodes.

1. A `GMSACredentialSpec` for _each_ gMSA account within a Domain
2. A Kubernetes secret for _each_ Active Directory Domain

The Kubernetes secret will contain the credentials required to authorize with Active Directory on behalf of the container. The user specified within this secret must have permissions within Active Directory to retrieve the gMSA account password. Once a user is properly configured, a simple secret can be created to outline the username and password of the account, as well as the domain which the account exists in. This secret must be created in the same namespace as an instance of the Account Provider API.  

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

### Checking if the Rancher gMSA CCG Plugin Works

1. Ensure that the init container deployed with each pod of the `rancher-gmsa-plugin-installer` chart successfully completes and does not log any errors.
2. Ensure that all pods spawned from the `rancher-gmsa-account-provider` deamon set have started properly and do not log any errors 
3. Deploy a Windows workload which leverages a gMSA account, ensure that it successfully transitions to 'Running'

## Uninstalling the Rancher gMSA CCG Plugin

After deleting the Helm Charts, you may want to remove artifacts and certificates written to the host. To do this, SSH into each host and run `C:\Program Files\RanchergMSACredentialProvider\cleanup.ps1`
