# `ccg-plugin-installer`

The CCG Plugin Manager is responsible for installing, upgrading, and uninstalling the [CCG Plugin](../../src/rancher-ccg-plugin) as a [DLL (Dynamic Link Library)](https://learn.microsoft.com/en-us/troubleshoot/windows-client/deployment/dynamic-link-library) on a Windows host.

On a non-Windows host, the CCG Plugin Manager will emit a log on each operation; no further changes will be made.

## Supported Commands

The following commands are supported by the `ccg-plugin-installer`.

### `./bin/ccg-plugin-installer install`

The Plugin Manager will register the DLL onto the host using [`regsvc.exe`](https://learn.microsoft.com/en-us/dotnet/framework/tools/regsvcs-exe-net-services-installation-tool) and modify the CCG registry entry to include a reference to the Rancher gMSA CCG Plugin DLL it installed.

### `./bin/ccg-plugin-installer upgrade`

The Plugin Manager will re-register the DLL onto the host using [`regsvc.exe`](https://learn.microsoft.com/en-us/dotnet/framework/tools/regsvcs-exe-net-services-installation-tool).

### `./bin/ccg-plugin-installer uninstall`

The Plugin Manager will deregister the DLL from the host using `regsvc.exe` and execute a cleanup script which can be used to remove all data and references to the Rancher gMSA Credential Plugin feature. 
