Param(
  [string]$Root = "C:\dt\code\dev26\ElmerPad\syncfiles",
  [int]$Port = 12345,
  [string]$Token = "mytoken",
  [switch]$AllowLan,
  [switch]$UseGoRun,
  [switch]$Background,
  [string]$Exe = ".\elmer-syncd.exe",
  [string]$Log = ".\elmer-syncd.log"
)

# Prepare
if (-not (Test-Path $Root)) { New-Item -ItemType Directory -Path $Root -Force | Out-Null }
$bind = if ($AllowLan) { "0.0.0.0" } else { "127.0.0.1" }
$commonArgs = @("--root", $Root, "--port", $Port.ToString(), "--token", $Token)
if ($AllowLan) { $commonArgs += "--allow-lan" }

# Choose runner
if ($UseGoRun -or -not (Test-Path $Exe)) {
  $file = "go"
  $args = @("run","./cmd/elmer-syncd") + $commonArgs
} else {
  $file = (Resolve-Path $Exe).ProviderPath
  $args = $commonArgs
}

# Start foreground or background
if ($Background) {
  $startInfo = @{
    FilePath = $file
    ArgumentList = $args
    RedirectStandardOutput = $true
    RedirectStandardError  = $true
    NoNewWindow = $true
    PassThru = $true
  }
  $proc = Start-Process @startInfo
  # redirect output asynchronously to log
  Start-Sleep -Milliseconds 200
  "Started (PID=$($proc.Id)). Logging to $Log"
  # stream output to log file
  $proc | Out-Null
  while (-not $proc.HasExited) {
    try {
      $out = $proc.StandardOutput.ReadToEnd()
      $err = $proc.StandardError.ReadToEnd()
      if ($out) {Add-Content -Path $Log -Value $out}
      if ($err) {Add-Content -Path $Log -Value $err}
    } catch { Start-Sleep -Milliseconds 200 }
    Start-Sleep -Seconds 1
  }
  "Process exited with code $($proc.ExitCode)" | Out-File -FilePath $Log -Append
} else {
  # foreground — run and show output
  & $file @args
}