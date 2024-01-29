# Rancher gMSA Account Provider
This helm chart will deploy the Rancher gMSA Account Provider API as a daemon set across all Windows worker nodes.

Only one instance of the Rancher gMSA Account Provider API can be deployed to a given namespace. In cases where multiple Account Provider APIs are required, individual namespaces must be created per provider. 

The Rancher gMSA Account Provider assists the `Rancher gMSA Container Credential Guard Plugin` in the non-domain-joined gMSA authorization process. The Account Provider deploys an HTTP/s API running as `hostProcess` pod on all Windows workers. As workloads utilizing gMSA accounts are created, the 'Rancher gMSA Container Credential Guard Plugin' will query the Account Provider API to obtain Active Directory Domain login information. All Domain login information is stored as secrets within the cluster, allowing for native Kubernetes management of credentials.  


## Prerequisites

+ Kubernetes v1.24+
+ ContainerD v1.7+
+ The latest version of 'cert-manager'
+ The 'Rancher gMSA CCG Plugin' chart must already be installed
+ The 'Rancher GMSA' chart must already be installed
+ One or more Windows worker nodes running one of the following OS versions: Windows Server 2019, Windows Server 2022, Windows Server Core 2019, Windows Server Core 2022
+ An Active Directory instance which can be contacted by the cluster
+ One or more Group Managed Service Accounts

