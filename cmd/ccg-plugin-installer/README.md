`rancher-ccg-plugin-installer`
===

The CCG Plugin Manager is responsible for managing the [CCG Plugin](../../src/rancher-ccg-plugin) as a [DLL (Dynamic Link Library)](https://learn.microsoft.com/en-us/troubleshoot/windows-client/deployment/dynamic-link-library) on Windows hosts. The application embeds a version of the Rancher gMSA Plugin DLL, as well as an installation, uninstallation, and cleanup scripts. This application can only be run on Windows systems, execution on any other platform will result in a no-op. 

The application currently supports two commands:

- On `install`, it registers the DLL onto the host using `regsvc.exe` and modifies the CCG registry entry to include a reference to the Rancher gMSA CCG Plugin DLL.

- On `uninstall`, it will deregister the DLL from the host using `regsvc.exe`, and generate a cleanup script which can be used to remove all data and references to the Rancher gMSA Credential Plugin feature. 

