param(
  [string]$ExePath = "./elmer-syncd.exe",
  [string]$InnoPath = "C:\\Program Files (x86)\\Inno Setup 6\\ISCC.exe",
  [string]$OutDir = "dist"
)
if (-not (Test-Path $ExePath)) { Write-Error "Executable not found: $ExePath"; exit 1 }
if (-not (Test-Path $InnoPath)) { Write-Error "Inno Setup compiler not found: $InnoPath"; exit 1 }
if (-not (Test-Path $OutDir)) { New-Item -ItemType Directory -Path $OutDir | Out-Null }

$iss = (Get-Content .\ElmerSyncdInstaller.iss) -join "`n"
# write ISS to temporary file (we use ISCC defines instead of replacing)
$tmpIss = Join-Path $OutDir "ElmerSyncdInstaller_tmp.iss"
Set-Content -Path $tmpIss -Value $iss

# Resolve and quote the executable path for ISCC /D define
$exeFull = (Resolve-Path $ExePath).ProviderPath
$define = "/DMyAppExe=`"$exeFull`""

Write-Host "Running ISCC..."
& $InnoPath $tmpIss $define "/O$OutDir"
Write-Host "Installer built in $OutDir"
