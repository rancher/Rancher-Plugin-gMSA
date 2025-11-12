# Developer Test Cases

Before raising a PR, the following tests cases should be checked **_at minimum_** for each supported environment. This page provides a set of repeatable test cases which verify that the entire deployment and authorization flow of the feature functions as expected. More specific test cases relating to a particular component of the feature can be found in their respective directories.

The creation and management of test environments can be automated through the use of terraform modules included within the [`rancher/windows`](https://github.com/rancher/windows/tree/main/terraform) repo. It's advised to use those terraform modules when preparing an environment to manually verify any of the test cases listed on this page, configuring these environments manually takes significant effort and time.

# Supported Versions of RKE2

+ `v1.24.14` and up  
+ `v1.25.10` and up  
+ `v1.26.5` and up
+ `v1.27.2` and up
+ All RKE2 versions `v1.28.0` and up

# Prerequisites 

Regardless of cluster configurations, the following prerequisites must be met:

+ A cluster running Kubernetes v1.24+, with ContainerD 1.7+
    + We develop against the latest version of [RKE2](https://github.com/rancher/rke2), however this project should also work on other Kubernetes distributions which support Windows Workers and have support for `hostProcess` pods
+ Windows worker nodes running one of the following OS versions: `Windows Server 2019`, `Windows Server 2022`, `Windows Server Core 2019`, `Windows Server Core 2022`. No other versions are supported. 
+ An Active Directory Domain which can be contacted by the Windows Worker nodes and workloads
+ An Active Directory impersonation (user) account and gMSA account, both of which are in the same security group
+ The latest version of the Rancher gMSA Web-hook chart installed (v3+) 
  + During installation, configure a `GMSACredentialSpec` for your Domain and gMSA Account. Ensure that the `HostAccountConfig` field is enabled and properly configured for the Rancher CCG gMSA Plugin, and that the `PluginInput` points to a valid secret within the Account Providers namespace.
    + As a reference, the Rancher gMSA Plugin GUID is `{e4781092-f116-4b79-b55e-28eb6a224e26}`. You must ensure that the value is wrapped in curly braces, otherwise CCG will not invoke the plugin.
    + As a reference, the Rancher gMSA Plugin DLL expects a `PluginInput` format of `<ACCOUNT_PROVIDER_NAMESPACE>:<SECRET>`, where `ACCOUNT_PROVIDER_NAMESPACE` is a namespace containing an Account Provider deployment. 
  + Ensure that the GMSACredentialSpec object specifies the GUID and SID of the _domain_ and **not** the GMSA account. 
  + Ensure that a `ClusterRole` or `Role` that provides the `use` verb on the `GMSACredentialSpec` resource has been created.

# Test Cluster configuration 

It's recommended that each test run use a cluster with the following node configuration. Doing so ensures that all supported OS versions are tested at the same time. 

+ 1 Linux ETCD/CP node (running any distro supported by RKE2)
+ 4 Windows Workers
  + 1 Worker running Windows Server 2019
  + 1 Worker running Windows Server Core 2019
  + 1 Worker running Windows Server 2022
  + 1 Worker running Windows Server Core 2022

## Supported Environments and Cluster Configurations

This project supports a number of cluster environments:

+ Clusters with 1 or more Windows workers which are all **joined** to a **single** domain
+ Clusters with 1 or more Windows workers which are **joined** to a **single** domain _as well as_ 1 or more Windows workers **not joined** to any domain
+ Clusters with 1 or more Windows workers which are all **not joined** to any domain

We currently do **not** provide support for or test against the following environments, though they are technically possible

+ Clusters with 1 or more Windows workers interacting with **multiple** domains at once, either joined or un-joined
+ Clusters interacting with **multiple** Active Directory instances at once

## Configure RBAC for your service account

In order to test the Account Provider you must first install the GMSA Web-hook, which will provide the `GMSACredentialSpec` CRD and will expand the contents of the resource onto pods. The web-hook chart will deploy a purpose built `ClusterRole` which permits access to all `GMSACredentialSpec` resources. However, the default `ClusterRoleBinding` deployed by the GMSA Web-hook chart only grants access to these resources to the service account used by the GMSA web-hook. In order to create workloads which utilize a `GMSACredentialSpec`, but run under a different service account, an additional `Role` and `RoleBinding` needs to be created for the service account you intend to use for testing.

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: get-gmsa-spec
  namespace: cattle-wins-system
rules:
  - apiGroups:
      - windows.k8s.io
    resources:
      - gmsacredentialspecs
    verbs:
      - use
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: get-gmsa-spec-rb
  namespace: cattle-wins-system
roleRef:
  apiGroup: rbac.authorization.k8s.io/v1
  kind: Role
  name: get-gmsa-spec
subjects:
  - kind: ServiceAccount
    name: <YOUR_SERVICE_ACCOUNT_NAME>
```

## Sample gMSA Workload

Each test may utilize the same gMSA workload to verify the proper function of the feature. A sample workload looks like the following: 

<details hidden>

<summary>
Sample gMSA Workload
</summary>

The below yaml can be used to deploy a workload utilizing a gMSA account. Several fields must be modified in accordance with your Active Directory environment. The workload uses a `windows/servercore` base image, to leverage the Active Directory authentication apis. 

> *Note*
> The service created by this workload will not be accessible via the Rancher UI, you must directly connect to the NodeIP in your browser. For most test scenarios, using the UI is not required to determine if the pod has successfully assumed the role of a gMSA account. 

The `servercore` base image is ~4GB (!). Expect a lengthy image pull time the first time you deploy this workload; grab a drink, relax.

```yaml 
---
kind: ConfigMap
apiVersion: v1
metadata:
  labels:
    app: gmsa-demo
  name: gmsa-demo
  namespace: cattle-wins-system
data:
  run.ps1: |
    $ErrorActionPreference = "Stop"

    Write-Output "Configuring IIS with authentication."

    # Add required Windows features, since they are not installed by default.
    Install-WindowsFeature "Web-Windows-Auth", "Web-Asp-Net45"

    # Create simple ASP.NET page.
    New-Item -Force -ItemType Directory -Path 'C:\inetpub\wwwroot\app'
    Set-Content -Path 'C:\inetpub\wwwroot\app\default.aspx' -Value 'Authenticated as <B><%=User.Identity.Name%></B>, Type of Authentication: <B><%=User.Identity.AuthenticationType%></B>'

    # Configure IIS with authentication.
    Import-Module IISAdministration
    Start-IISCommitDelay
    (Get-IISConfigSection -SectionPath 'system.webServer/security/authentication/windowsAuthentication').Attributes['enabled'].value = $true
    (Get-IISConfigSection -SectionPath 'system.webServer/security/authentication/anonymousAuthentication').Attributes['enabled'].value = $false
    (Get-IISServerManager).Sites[0].Applications[0].VirtualDirectories[0].PhysicalPath = 'C:\inetpub\wwwroot\app'
    Stop-IISCommitDelay

    Write-Output "IIS with authentication is ready."

    C:\ServiceMonitor.exe w3svc
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: gmsa-demo
  name: gmsa-demo
  namespace: cattle-wins-system
spec:
  replicas: 1
  selector:
    matchLabels:
      app: gmsa-demo
  template:
    metadata:
      labels:
        app: gmsa-demo
    spec:
      serviceAccountName: gmsa
      containers:
        - name: iis
          image: mcr.microsoft.com/windows/servercore/iis:windowsservercore-ltsc2019
          imagePullPolicy: IfNotPresent
          securityContext:
            windowsOptions:
              gmsaCredentialSpecName: gmsa1-ccg
          ports:
            - containerPort: 80
          command:
            - powershell
          args:
            - -File
            - /gmsa-demo/run.ps1
          volumeMounts:
            - name: gmsa-demo
              mountPath: /gmsa-demo
      volumes:
        - configMap:
            defaultMode: 420
            name: gmsa-demo
          name: gmsa-demo
      nodeSelector:
        kubernetes.io/os: windows
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: gmsa-demo
  name: gmsa-demo
  namespace: cattle-wins-system
spec:
  ports:
    - port: 80
      targetPort: 80
  selector:
    app: gmsa-demo
  type: NodePort

```

</details>

Repeatable Environment Test Cases
---

The following procedure should be performed against all supported environments at minimum. These steps ensure that the feature works as expected on each supported Windows OS version, but does not cover specific edge cases or component specific tests. For more specific test steps refer to the `account-provider` and `plugin-installer` directories. 

0. Ensure that all the prerequisites listed in the `Prerequisites` section are met
1. Install the Rancher gMSA CCG Plugin Installer chart onto the cluster
    1. Ensure a pod is deployed onto each Windows worker
    2. Ensure that all pods become available, and that all init containers exit successfully
2. SSH / RDP into each Windows host, and ensure that the files listed in the `Expected Files For The Plugin Installer Post Install` section exist
3. Install the Rancher gMSA Account Provider chart onto the cluster
    1. Ensure a pod is deployed to all applicable windows workers
    2. Ensure that all pods become available, and there are no errors shown in the logs
4. SSH / RDP into each host, and ensure the files listed in the `Expected Files For The Account Provider Post Install` section exist
5. Create a Windows workload which leverages your gMSA account (see above sample workload for an example)
    1. Ensure the workload becomes available, and that the CCG event log does not log any errors 
6. SSH into the sample workload and run the following commands
   1. `nltest /parentdomain` should return the domain name configured within the GMSACredentialSpec 
   2. `nltest /query` should return `NERR_Success`, indicating no error was encountered contacting the domain controller
   3. `nltest /sc_query:<YOUR_DOMAIN_NAME>` should return `NERR_Success`, indicating no error was encountered contacting the domain controller
7. Connect to the service over the newly created node port, and login using an active directory user when prompted. This user should be different from the GMSA impersonation account username and password.
8. Delete the sample Workload
9. Uninstall the Rancher gMSA Account Provider Chart
10. Start to remove the CCG Plugin Installer Helm application
   1. First, modify the Helm release and change the `action` field to `uninstall` to uninstall the plugin
   2. Wait for the new ccg plugin installer container to deploy and initialize 
   3. Remove the Helm release from the cluster
11. SSH into each node, and ensure that the files listed in `Expected Files For The Plugin Installer Post Uninstall` exist 
12. SSH into each node, and Run the `cleanup.ps1` script and ensure that the files listed in `Expected Files For The Plugin Installer Post Installer` **no longer exist.** 

## Expected Files For The Plugin Installer Post Install
After initial installation, the following files should exist on each host:

`C:\Program Files\RanchergMSACredentialProvider\RanchergMSACredentialProvider.dll`

`C:\Program Files\RanchergMSACredentialProvider\RanchergMSACredentialProvider.tlb`

`C:\Program Files\RanchergMSACredentialProvider\install-plugin.ps1`

## Expected Files For The Plugin Installer Post Uninstall
After uninstallation of the chart, all of the above files should no longer exist

## Expected Files For The Account Provider Post Install

After installation, the following files should exist on the host: 

`/var/lib/rancher/gmsa/<NAMESPACE>/port.txt`

`/var/lib/rancher/gmsa/<NAMESPACE>/ssl/client/tls.pfx`
`/var/lib/rancher/gmsa/<NAMESPACE>/ssl/client/tls.crt`

`/var/lib/rancher/gmsa/<NAMESPACE>/ssl/server/tls.crt`
`/var/lib/rancher/gmsa/<NAMESPACE>/ssl/server/ca.crt`

`/var/lib/rancher/gmsa/<NAMESPACE>/ssl/ca/tls.crt`
`/var/lib/rancher/gmsa/<NAMESPACE>/ssl/ca/ca.crt`

## Expected Files For The Account Provider Post Uninstall
All the previously listed files should no longer exist once the Account Provider chart is uninstalled from the cluster. The uninstallation process is handled via a [Helm Hook](https://helm.sh/docs/topics/charts_hooks/). If these resources remain on a Windows node post uninstall of the chart, then there is an issue with the Helm Hook logic which must be addressed.


## Additional Assistance

If you need additional debugging assistance, you can refer to this documentation https://learn.microsoft.com/en-us/virtualization/windowscontainers/manage-containers/gmsa-troubleshooting