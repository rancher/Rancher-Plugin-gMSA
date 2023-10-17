# Rancher gMSA CCG Plugin Installer

Rancher GMSA Plugin Installer is a **utility chart** that can be deployed onto a Windows cluster to install the Container Credential Guard Rancher Kubernetes Cluster Plugin (CCGRKC Plugin) onto your Windows hosts.

This plugin is invoked by the Windows Container Credential Guard during the [non-domain-joined gMSA authorization process](https://learn.microsoft.com/en-us/virtualization/windowscontainers/manage-containers/manage-serviceaccounts#gmsa-architecture-and-improvements).

Once installed, the plugin also requires you to deploy the Rancher gMSA Account Provider onto the cluster.

## Prerequisites

+ Kubernetes v1.24+
+ ContainerD v1.7+
+ One or more Windows worker nodes running one of the following OS versions: Windows Server 2019, Windows Server 2022, Windows Server Core 2019, Windows Server Core 2022

