# Getting Started

The Rancher CCG gMSA Plugin feature comprises two components, the plugin installer (which installs the CCG plugin), and the account provider (which the ccg plugin uses to obtain domain login credentials). Installing both Helm charts is required in order for a cluster to support non-domain-joined nodes. 

## Prerequisites 

+ Kubernetes v1.24+, containerD 1.7+
+ Nodes running `Windows Server 2019`, `Windows Server 2022`, `Windows Server Core 2019`, or `Windows Server Core 2022 
`+ The gMSA Web-hook chart installed ([Official Repo](https://github.com/kubernetes-sigs/windows-gmsa), [Rancher Specific Chart](https://github.com/rancher/charts/tree/release-v2.7/charts/rancher-windows-gmsa))\
  + **Note:** Currently, the gMSA Web-Hook chart offered by Rancher **does not contain the required changes to support non-domain joined nodes**. This chart will be updated soon, in the meantime you should use the following repository `harrisonwaffel/charts` and test against the `update-gmsa` branch (this can be done by configuring a new app repository in the Rancher UI).
+ An Active Directory instance which is reachable by worker nodes
+ One or more Active Directory gMSA accounts 

## Simple Installation

### In Rancher (via Apps & Marketplace)

1. Install the latest version of `cert-manager` onto your cluster
2. Navigate to `Apps & Marketplace -> Repositories` in your target downstream cluster and create a Repository that points to a `Git repository containing Helm chart or cluster template definitions` where the `Git Repo URL` is `https://github.com/rancher/Rancher-Plugin-gMSA` and the `Git Branch` is `main`
3. Navigate to `Apps & Marketplace -> Repositories` in your target downstream cluster and create a Repository that points to a `Git repository containing Helm chart or cluster template definitions` where the `Git Repo URL` is `https://github.com/harrisonwaffel/charts` and the `Git Branch` is `update-gmsa`
4. Navigate to `Apps & Marketplace -> Charts`; you should see three charts under the two new Repositories you created:  `Windows GMSA`, `Rancher gMSA CCG Plugin`, and `Rancher gMSA Account Provider`.
5. Install `Windows GMSA` first, and configure a `GMSACredentialSpec`
6. Install `Rancher gMSA CCG Plugin` second
7. Install `Rancher gMSA Account Provider` third

### In a normal Kubernetes cluster (via running Helm 3 locally)

1. Ensure that the Rancher `Windows GMSA` chart is already installed

2. Install `rancher-gmsa-plugin-installer` onto your cluster via Helm to install the CCG Plugin DLL onto your Windows worker nodes

```
helm install -n cattle-gmsa-system Rancher-Plugin-gMSA charts/rancher-gmsa-plugin-installer
```

3. Install `rancher-gmsa-account-provider` chart to deploy the account provider API onto your Windows worker nodes.  

```bash
helm install -n cattle-helm-system Rancher-Plugin-gMSA-account-provider charts/rancher-gmsa-account-provider
```

### Checking if the Rancher gMSA CCG Plugin Works

1. Ensure that the init container deployed with each pod of the `rancher-gmsa-plugin-installer` chart successfully completes and does not log any errors.
2. Ensure that all pods spawned from the `rancher-gmsa-account-provider` deamon set have started properly and do not log any errors
3. Configure a `GMSACredentialSpec` resource to specify an existing gMSA account, and ensure that the `HostAccountConfig` field is present and configured to use the Rancher CCG gMSA Plugin. 
4. Deploy a Windows workload which leverages a gMSA account, ensure that it successfully transitions to 'Running'

## Uninstalling the Rancher gMSA CCG Plugin

After deleting the Helm Charts, you may want to remove artifacts and certificates written to the host. To do this, SSH into each host and run `C:\Program Files\RanchergMSACredentialProvider\cleanup.ps1`
