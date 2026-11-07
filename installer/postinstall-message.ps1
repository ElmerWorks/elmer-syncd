param(
  [int]$Port = 12345,
  [string]$InstallDir = $null
)

# Determine script directory (install dir)
# If an explicit InstallDir was provided use it. Otherwise prefer $PSScriptRoot (PowerShell v3+). Fallback to MyInvocation if not set.
if ($InstallDir) {
  $scriptDir = $InstallDir
} elseif ($PSVersionTable -and $PSVersionTable.PSVersion) {
  if ($PSScriptRoot) {
    $scriptDir = $PSScriptRoot
  } else {
    $scriptPath = $MyInvocation.MyCommand.Definition
    $scriptDir = Split-Path -Parent $scriptPath
  }
} else {
  $scriptPath = $MyInvocation.MyCommand.Definition
  $scriptDir = Split-Path -Parent $scriptPath
}

function Get-LocalIPv4 {
  $ips = @()
  try {
    $host = [System.Net.Dns]::GetHostEntry([System.Net.Dns]::GetHostName())
    foreach ($a in $host.AddressList) {
      if ($a.AddressFamily -eq 'InterNetwork') {
        $s = $a.ToString()
        if (-not $s.StartsWith('127.') -and -not $s.StartsWith('169.')) { $ips += $s }
      }
    }
  } catch {
    # fallback
  }
  if ($ips.Count -eq 0) { $ips = @('127.0.0.1') }
  return $ips
}

$ips = Get-LocalIPv4
$lines = @()
$lines += "ElmerSyncd installed. Connection examples (port $Port):"
foreach ($ip in $ips) {
  $lines += "  WebSocket: ws://$($ip):$Port/ws"
  $lines += "  REST list: http://$($ip):$Port/api/list?path="
}
$lines += ""
$lines += "If you configured a token, include it in header X-API-Token when connecting."

# write a connection-info.txt file next to the script
try {
  $outFile = Join-Path $scriptDir 'connection-info.txt'
  # Ensure target directory exists
  if (-not (Test-Path -Path $scriptDir -PathType Container)) {
    New-Item -ItemType Directory -Path $scriptDir -Force | Out-Null
  }
  Set-Content -Path $outFile -Value $lines -Encoding UTF8 -Force
  Write-Host "Wrote connection info to $outFile"
} catch {
  # If writing to the install dir fails, write a copy to the user's TEMP and emit an error message
  $err = $_.Exception.Message
  $tempFile = Join-Path $env:TEMP 'elmer-connection-info.txt'
  try {
    Set-Content -Path $tempFile -Value ("ERROR writing to $($outFile)`n`nError: $($err)`n`nFallback content:`n" + ($lines -join "`n")) -Encoding UTF8 -Force
    Write-Host "Failed to write to $($outFile): $($err). Wrote fallback to $($tempFile)"
  } catch {
    # Last resort: write error to host
    Write-Host "Failed to write connection info to both $($outFile) and $($tempFile): $($err)"
  }
}

