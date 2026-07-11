# Run on the LAPTOP, not the phone - cloudflared's binary doesn't run on
# Android 5.1. The phone still runs the actual blog server on :8080; this
# just proxies it to the internet over LAN.
#
# Unlike run-tunnel.ps1, this script closes the loop: it starts the Quick
# Tunnel, scrapes the random https://*.trycloudflare.com URL from its logs,
# writes it into cloudflare/worker/wrangler.toml as PHONE_URL, and runs
# `wrangler deploy` so the Worker picks it up - no manual dashboard edit.
#
# Requires: cloudflare/worker/wrangler.toml already exists (copy from
# wrangler.toml.example) and `wrangler` is already authenticated
# (wrangler login).
#
# Usage: .\run-tunnel-auto.ps1 <phone-lan-ip>
# Find the phone's LAN IP in Termux with: ifconfig | grep wlan0 -A1

param(
    [Parameter(Mandatory = $true)]
    [string]$PhoneIp
)

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$workerDir = Join-Path $repoRoot "cloudflare\worker"
$wranglerToml = Join-Path $workerDir "wrangler.toml"
$stderrLog = Join-Path $env:TEMP "cloudflared-tunnel.stderr.log"
$stdoutLog = Join-Path $env:TEMP "cloudflared-tunnel.stdout.log"

if (-not (Test-Path $wranglerToml)) {
    Write-Error "wrangler.toml not found at $wranglerToml. Copy wrangler.toml.example to wrangler.toml and fill in RENDER_URL first."
}

foreach ($f in @($stderrLog, $stdoutLog)) {
    if (Test-Path $f) { Remove-Item $f -Force }
}

Write-Host "Starting Cloudflare Quick Tunnel -> http://${PhoneIp}:8080 ..."
$proc = Start-Process -FilePath "cloudflared" `
    -ArgumentList @("tunnel", "--url", "http://${PhoneIp}:8080") `
    -RedirectStandardError $stderrLog `
    -RedirectStandardOutput $stdoutLog `
    -NoNewWindow -PassThru

$tunnelUrl = $null
$deadline = (Get-Date).AddSeconds(30)
while (-not $tunnelUrl -and (Get-Date) -lt $deadline) {
    Start-Sleep -Milliseconds 500
    if (Test-Path $stderrLog) {
        $match = Select-String -Path $stderrLog -Pattern "https://[a-zA-Z0-9-]+\.trycloudflare\.com" -ErrorAction SilentlyContinue | Select-Object -First 1
        if ($match) {
            $tunnelUrl = $match.Matches[0].Value
        }
    }
    if ($proc.HasExited) {
        Write-Error "cloudflared exited early (exit code $($proc.ExitCode)). Check $stderrLog"
    }
}

if (-not $tunnelUrl) {
    Write-Error "Timed out waiting for tunnel URL after 30s. Check $stderrLog"
}

Write-Host "Tunnel URL: $tunnelUrl"

$content = Get-Content $wranglerToml -Raw
if ($content -match '(?m)^PHONE_URL\s*=.*$') {
    $content = $content -replace '(?m)^PHONE_URL\s*=.*$', "PHONE_URL = `"$tunnelUrl`""
} else {
    $content = $content -replace '(\[vars\]\r?\n)', "`$1PHONE_URL = `"$tunnelUrl`"`r`n"
}
Set-Content -Path $wranglerToml -Value $content -NoNewline

Write-Host "Deploying Worker with updated PHONE_URL..."
Push-Location $workerDir
try {
    wrangler deploy
} finally {
    Pop-Location
}

Write-Host ""
Write-Host "Worker updated. Tunnel running (PID $($proc.Id))."
Write-Host "Press Ctrl+C to stop the tunnel."
Wait-Process -Id $proc.Id
