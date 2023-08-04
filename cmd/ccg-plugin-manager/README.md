`rancher-ccg-plugin-manager`
===

The CCG Plugin Manager is responsible for managing the [CCG Plugin](../../src/rancher-ccg-plugin) as a [DLL (Dynamic Link Library)](https://learn.microsoft.com/en-us/troubleshoot/windows-client/deployment/dynamic-link-library) on a Windows host.

It supports three commands:

- On `install`, it registers the DLL onto the host.

- On `upgrade`, it will reregister the DLL that it is packaged with onto the host.

- On `uninstall`, it will unregister the DLL from the host.

