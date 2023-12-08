# Rancher gMSA Plugin Installer

This directory contains the source code for the Rancher gMSA CCG Plugin Installer.

This component is responsible for installing, uninstalling, and upgrading the Rancher CCG Plugin DLL on Windows nodes.

> **Note**: Installing the Rancher CCG Plugin DLL is a **prerequisite / requirement** for installing the gMSA Account Provider API on your cluster.

## Install / Upgrade

On installing / upgrading this DLL, the CCG DLL Installer will take the following steps:

1. Create a directory at `C:\Program Files\RanchergMSACredentialProvider`

2. Install the DLL embedded in the CCG Plugin Installer binary at `C:\Program Files\RanchergMSACredentialProvider\RanchergMSACredentialProvider.dll`

> **Note**: If there is already a DLL installed, the CCG Plugin Installer will move the existing DLL to a temporary file and delete it once the upgrade is complete.
>
> If the upgrade fails, it will attempt to move the old DLL back into place.

3. Write and execute a Powershell-based install script at `C:\Program Files\RanchergMSACredentialProvider\install-plugin.ps1`.

> **Note**: This small Powershell script will register the CCG Plugin DLL onto the system using [`regsvc.exe`](https://learn.microsoft.com/en-us/dotnet/framework/tools/regsvcs-exe-net-services-installation-tool).

## Uninstall

On uninstalling this DLL, the CCG DLL Installer will take the following steps:

1. Write and execute a Powershell-based uninstall script at `C:\Program Files\RanchergMSACredentialProvider\uninstall-plugin.ps1`.

> **Note**: This small Powershell script will deregister the CCG Plugin DLL from the system using [`regsvc.exe`](https://learn.microsoft.com/en-us/dotnet/framework/tools/regsvcs-exe-net-services-installation-tool).

2. Delete the directory at `C:\Program Files\RanchergMSACredentialProvider` along with all its contents (i.e. the DLL file and any scripts).

## How do I use the plugin? 

Please see the [Getting Started](../../../docs/gettingstarted.md) docs!