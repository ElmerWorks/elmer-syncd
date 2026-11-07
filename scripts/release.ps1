param(
  [Parameter(Mandatory=$true)] [string]$Version
)
if (-not (Get-Command goreleaser -ErrorAction SilentlyContinue)) {
  Write-Error "goreleaser not found in PATH"
  exit 1
}
Write-Host "Releasing $Version"
goreleaser release --rm-dist --tag $Version
