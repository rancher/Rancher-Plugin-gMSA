# Rancher gMSA CCG Plugin Installer

This helm chart will install the Rancher Container Credential Guard Plugin onto all Windows nodes. This plugin is invoked by the Windows Container Credential Guard during the non-domain-joined gMSA authorization process. This plugin requires that the Rancher gMSA Account Provider is also installed onto the cluster. 

Only one instance of this chart needs to be installed per cluster in order to install the Plugin onto all Windows workers. 

## Prerequisites

+ Kubernetes v1.24+
+ ContainerD v1.7+
+ One or more Windows worker nodes running one of the following OS versions: Windows Server 2019, Windows Server 2022, Windows Server Core 2019, Windows Server Core 2022

