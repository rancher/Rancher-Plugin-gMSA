# Rancher gMSA CCG Plugin Installer Testing & Troubleshooting Documentation 

The plugin installer is a simple `hostProcess` container that creates two files which are written to disk during the installation process.

The two files written to disk are:

1. `RanchergMSACredentialProvider.dll` is the CCG Plugin DLL developed by Rancher
2. `install-plugin.ps1` is a PowerShell script responsible for registering the CCG Plugin onto the host, as well as modifying a registry key such that CCG can invoke the plugin. This script is run by the installer automatically after being created.

During uninstallation, all files that were created during installation are expected to be deleted.

## Detailed Host Modifications

> **Note:** Registry keys in the below sections are abbreviated. `HKLM` is shown as `HK_LOCAL_MACHINE` in the registry editor.

### Install

During execution of the `install` command the host will be modified in three ways:

1. A new directory will be created (`C:\Program Files\RanchergMSACredentialProvider`) which will contain the CCG Plugin (`RanchergMSACredentialProvider.dll`) and the installation script used to register the plugin (`install-plugin.ps1`)
2. The installation script will be executed, which will invoke `regsvc.exe` to register the plugin DLL onto the host. This will result in a new registry key being created (`HKLM:\Software\Classes\CLSID\{E4781092-F116-4B79-B55E-28EB6A224E26}`). Additionally, `regsvc.exe` will generate the required type library for the plugin (`RanchergMSACredentialProvider.tlb`) within the DLL Directory.
3. The installation script will then create a new registry key to expose the plugin to CCG (`HKLM:SYSTEM\CurrentControlSet\Control\CCG\COMClasses\{e4781092-f116-4b79-b55e-28eb6a224e26}`)

### Uninstall

During execution of the `uninstall` command the host will be modified in two ways:

1. The plugin DLL file and type library file will be removed, and the installation script will be replaced with the uninstallation script (`uninstall-plugin.ps1`)
2. The uninstallation script will be executed, which will then deregister the DLL using `regsvc.exe \u`, subsequently removing the CLSID registry entry

## Debugging

### Install

The plugin installer contains logic to verify if the plugin is already installed on a host, as well as if the installation completed successfully. However, if you wish to verify a proper installation manually, the following should be done:

1. Ensure that the `C:\Program Files\RanchergMSACredentialProvider` directory exists using File Explorer or PowerShell
2. Ensure that the `RanchergMSACredentialProvider.dll` file exists in the directory 
3. Ensure that the `RanchergMSACredentialProvider.tlb` file exists in the directory
   1. If this file is missing, then it's likely that the plugin install script did not complete successfully
4. Using `regedit.exe`, ensure that the following registry keys exist 
   1. `HKLM:\Software\Classes\CLSID\{E4781092-F116-4B79-B55E-28EB6A224E26}`
      1. If this key is missing then `regsvc.exe` did not properly register the DLL onto the host
   2. `HKLM:SYSTEM\CurrentControlSet\Control\CCG\COMClasses\{e4781092-f116-4b79-b55e-28eb6a224e26}`
      1. If this key is missing then `install-plugin.ps1` likely failed to acquire permissions to modify the registry.

### Uninstall

1. Ensure that the `C:\Program Files\RanchergMSACredentialProvider` directory exists using File Explorer or PowerShell
2. Ensure that the `RanchergMSACredentialProvider.dll` file no longer exists in the directory
3. Ensure that the `RanchergMSACredentialProvider.tlb` file no longer exists in the directory
4. Ensure that the `cleanup.ps1` script exists in the directory
5. Using `regedit.exe`, ensure that the following registry key does not exist
    1. `HKLM:\Software\Classes\CLSID\{E4781092-F116-4B79-B55E-28EB6A224E26}`
6. (Optional, only applicable to setups which also have the Account provider installed) execute `clean.ps1` and ensure all subdirectories of `/var/lib/rancher/gmsa` have been removed properly


## Enabling / Viewing Debug Logs

The CCG Plugin, once installed, can optionally log to the Windows Event log. This functionality is disabled by default. To enable these logs, simply create a `enable-logs.txt` file in the `/var/lib/rancher/gmsa/<NAMESPACE>` directory, where `NAMESPACE` is the namespace containing an Account Provider. The file does not need to contain any content. 

Once enabled, you may use the Windows Event log to view the plugin logs. All logs are written to the global `Application` log source, and three log ids are used - `101` for `INFO` logs, `201` for `DEBUG` logs, and `301` for `ERROR` logs. 

