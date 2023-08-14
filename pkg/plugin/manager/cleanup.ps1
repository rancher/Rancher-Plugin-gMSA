# MUST remove all dynamic directories
Write-Host "Running Rancher Plugin Cleanup Script"

Write-Host "Removing /var/lib/rancher/gmsa"
$Destination = /var/lib/rancher/gmsa

# remove-item -recurse does not work properly, will not delete the root folder on its own
Get-ChildItem -Path $Destination -Recurse | Remove-Item -force -recurse
Remove-Item $Destination -Force

# SHOULD remove certificates added to store
# todo