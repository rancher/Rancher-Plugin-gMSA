function Execute-Scripts {
    param (
        [Parameter(Mandatory=$true)]
        [string[]] $Scripts
    )

    foreach ($script in $Scripts) {
        Write-Host "---"
        Write-Host "RUN: $script.ps1..." -ForegroundColor Green
        & $PSScriptRoot\$script.ps1
        $exitCode = $LASTEXITCODE
        if ($exitCode -eq 0) {
            continue
        }
        Write-Host "FAILED: $script.ps1 (exit code $exitCode)" -ForegroundColor Red
        exit $exitCode
    }
}

function Get-BuildData {
    if ((git status --porcelain --untracked-files=no | Measure-Object).Count -gt 0) {
        $dirty="-dirty"
    }
    $commit = git rev-parse --short HEAD
    $gitTag = git tag -l --contains HEAD | Select-Object -Last 1
    if ((-not $dirty) -and ($gitTag)) {
        $version = $gitTag
    } else {
        $version = "$commit$dirty"
    }
    $os = go env GOHOSTOS
    $arch = go env GOHOSTARCH
    $ltsc = Get-LTSCVersion
    $repo = "rancher"
    if ($env:REPO) {
        $repo = $env:REPO
    }
    $tag = "$version-$arch"
    if ($env:TAG) {
        $repo = $env:TAG
    }
    return [PSCustomObject]@{
        VERSION = $version

        OS = $os
        ARCH = $arch
        LTSC = $ltsc

        REPO = $repo
        TAG = $tag
    }
}

function Clone-Directory {
    param (
        [Parameter(Mandatory=$true)]
        [string] $From,
        [Parameter(Mandatory=$true)]
        [string] $To
    )

    Remove-Item "$To" -Recurse -Force -ErrorAction SilentlyContinue
    $null = New-Item -ItemType Directory -Force -Path "$To"
    Copy-Item -Path "$From/*" -Destination "$To" -Recurse -Force
}

function Go-Build {
    param (
        [Parameter(Mandatory=$true)]
        [string[]] $Apps,
        [Parameter(Mandatory=$false)]
        [string[]] $OSArchs
    )

    if (-not $OSArchs) {
        $buildData = Get-BuildData
        $os = $buildData.OS
        $arch = $buildData.ARCH
        $OSArchs = @("$os $arch")
    }

    foreach ($app in $apps) {
        Write-Host "Building binaries for $app..." -ForegroundColor Yellow
        $LINKFLAGS="-X github.com/rancher/Rancher-Plugin-gMSA/pkg/version.Version=$VERSION"
        $LINKFLAGS="-X github.com/rancher/Rancher-Plugin-gMSA/pkg/version.GitCommit=$COMMIT $LINKFLAGS"

        foreach ($osArch in $osArchs) {
            $os = ($osArch -split ' ')[0]
            $arch = ($osArch -split ' ')[1]
            $suffix = "$os-$arch"
            if ($os -eq "windows") {
                $suffix="$suffix.exe"
            }
            $env:GOOS=$os
            $env:GOARCH=$arch
            $env:CGO_ENABLED=0
            go build -ldflags $LINKFLAGS -o "bin/$app-$suffix" ./cmd/$app
        }
    }
}

function Docker-Build {
    param (
        [Parameter(Mandatory=$true)]
        [string[]] $Apps
    )

    $buildData = Get-BuildData
    $os = $buildData.OS
    $arch = $buildData.ARCH
    $repo = $buildData.REPO
    $tag = $buildData.TAG
    $ltsc = $buildData.LTSC

    foreach ($app in $Apps) {
        $dockerfile = "package/$App/Dockerfile.windows"
        $image = "$repo/$app`:$tag-$ltsc"

        Write-Host "Building $image..." -ForegroundColor Yellow
        docker build --build-arg "ARCH=$arch" --build-arg "OS=$os" --build-arg="NANOSERVER_VERSION=$ltsc" -f $dockerfile -t $image .
    }

    Write-Host "To push these images to DockerHub:" -ForegroundColor Blue

    foreach ($app in $Apps) {
        $image = "$repo/$app`:$tag-$ltsc"
        Write-Host "docker push $image" -ForegroundColor Blue
    }
}

function Get-LTSCVersion {
    $windowsVersion = (Get-WmiObject -Class Win32_OperatingSystem).Version
    $buildNumber = $windowsVersion.Split('.')[2]
    if ($buildNumber -ge 17763 -and $buildNumber -le 19044) {
        $ltsc = "ltsc2019"
    } elseif ($buildNumber -gt 19044) {
        $ltsc = "ltsc2022"
    } elseif ($buildNumber -lt 17763) {
        throw "Invalid build version $windowsVersion. Only ltsc2019 or ltsc2022 are supported."
    }
    return $ltsc
}

function Check-DirectoryAndFiles {
    param (
        [Parameter(Mandatory=$true)]
        [string] $Directory,
        [Parameter(Mandatory=$false)]
        [string[]] $Files
    )

    # Check if the directory exists
    if (Test-Path $Directory) {
        Write-Output "Directory $Directory exists"

        # Check if each file exists
        foreach ($fileName in $Files) {
            $filePath = Join-Path -Path $directory -ChildPath $fileName
            if (Test-Path $filePath) {
                Write-Output "File $filePath exists"
            } else {
                throw "File $filePath does not exist"
            }
        }
    } else {
        throw "Directory $Directory does not exist"
    }
}

function Check-DLLComClassRegistered {
    param (
        [Parameter(Mandatory=$true)]
        [string] $DLLGuid
    )

    $DLLGuid = $DLLGuid.ToLower()

    $registryKey = "HKLM:\SYSTEM\CurrentControlSet\Control\CCG\COMClasses\{$DLLGuid}"

    # Check if the directory exists
    if (Test-Path $registryKey) {
        Write-Output "Registry key $registryKey exists"
    } else {
        throw "Registry key $registryKey does not exist"
    }
}

function Check-DllRegistered {
    param (
        [Parameter(Mandatory=$true)]
        [string] $DLLGuid
    )

    $DLLGuid = "{${DLLGuid}}"

    $keys = Get-ChildItem "HKLM:\Software\Classes\CLSID" -Recurse -ErrorAction SilentlyContinue |
            Get-ItemProperty -ErrorAction SilentlyContinue |
            Where-Object { $_.PSChildName -eq $DLLGuid }

    if ($keys) {
        Write-Output "DLL with GUID $DLLGuid is registered"
        $keys | Format-List
        return
    }

    throw "DLL with GUID $DLLGuid is not registered"
}

function Check-DLLInstalled {
    Check-DirectoryAndFiles -Directory "C:\Program Files\RanchergMSACredentialProvider" -Files @("install-plugin.ps1", "RanchergMSACredentialProvider.dll", "RanchergMSACredentialProvider.tlb")
    Check-DLLComClassRegistered -DLLGuid "E4781092-F116-4B79-B55E-28EB6A224E26"
    Check-DllRegistered -DLLGuid "E4781092-F116-4B79-B55E-28EB6A224E26"
}

function Check-DLLUninstalled {
    $foundError = $false

    try {
        Check-DirectoryAndFiles -Directory "C:\Program Files\RanchergMSACredentialProvider"
        $foundError = $true
    } catch {
        Write-Output $_.Exception.Message
    }

    if ($foundError) {
        throw "Expected C:\Program Files\RanchergMSACredentialProvider to not exist"
    }

    try {
        Check-DLLComClassRegistered -DLLGuid "E4781092-F116-4B79-B55E-28EB6A224E26"
        $foundError = $true
    } catch {
        Write-Output $_.Exception.Message
    }

    if ($foundError) {
        throw "Expected COM class to not be registered"
    }

    try {
        Check-DllRegistered -DLLGuid "E4781092-F116-4B79-B55E-28EB6A224E26"
        $foundError = $true
    } catch {
        Write-Output $_.Exception.Message
    }

    if ($foundError) {
        throw "Expected DLL to not be registered"
    }
}
