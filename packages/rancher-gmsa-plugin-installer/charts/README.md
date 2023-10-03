rancher-gmsa-plugin-installer
========

Rancher gMSA Plugin Installer is a **utility chart** that can be deployed onto a Windows cluster to install the Container Credential Guard Rancher Kubernetes Cluster Plugin (CCGRKC Plugin) onto your Windows hosts.

This plugin is invoked by the Windows Container Credential Guard during the [non-domain-joined gMSA authorization process](https://learn.microsoft.com/en-us/virtualization/windowscontainers/manage-containers/manage-serviceaccounts#gmsa-architecture-and-improvements). 

Once installed, the plugin also requires you to deploy the Rancher gMSA Account Provider onto the cluster.

> **Note**: This chart is **optional**. It is not required for the CCGRKC Plugin to work.
>
> See manual installation steps below if you would like to manually install / uninstall the plugin onto your hosts.

> **Note**: Only one instance of this chart needs to be installed per cluster in order to install the Plugin onto all Windows workers.


## Who needs the Rancher gMSA Plugin Installer?

Anyone who would like to declaratively manage the installation of the CCGRKC Plugin onto Windows hosts in a Kubernetes cluster.

## Components

### Install / Upgrade

On install / upgrade, a `DaemonSet` of `HostProcess` containers will be scheduled onto every Windows host in your cluster.

These `HostProcess` containers will run `ccg-plugin-installer.exe install` as an `initContainer` to install the CCGRKC Plugin DLL onto the Windows host.

After the CCGRKC Plugin DLL is successfully installed, the main container of each `Pod` will run no more operations (i.e. pause).

> **Note**: To check if the install or uninstall was successful, you should make sure you look at the `initContainer` logs since the main `container` logs will have no logs generated.

### Uninstall

On an uninstall, two workloads will be deployed as Helm [`post-delete` hooks](https://helm.sh/docs/topics/charts_hooks/): a `DaemonSet` of `HostProcess` containers and a `kubectl` Job.

The `HostProcess` containers will run `ccg-plugin-installer.exe uninstall` as an `initContainer` to uninstall the CCGRKC Plugin DLL from the Windows host.

After the CCGRKC Plugin DLL is successfully uninstalled, the main container of each `Pod` will run no more operations (i.e. pause).

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

### Manual Installation

If you would like to manually install this plugin onto your hosts instead of using this chart, you can find the `ccg-plugin-installer.exe` corresponding to the version of the CCGRKC Plugin you would like to install from the GitHub Releases page of this repository.

Once the binary is copied onto your host, simply run `ccg-plugin-installer.exe install` to install the plugin and `ccg-plugin-installer.exe uninstall` to uninstall the plugin from your host.

## Debugging

### How do I identify whether the plugin was successfully installed onto or uninstalled from the host?

TODO: TBD

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
