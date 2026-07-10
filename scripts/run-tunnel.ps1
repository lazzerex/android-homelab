# Run on the LAPTOP, not the phone - cloudflared's binary doesn't run on
# Android 5.1. The phone still runs the actual blog server on :8080; this
# just proxies it to the internet over LAN.
#
# Usage: .\run-tunnel.ps1 <phone-lan-ip>
# Find the phone's LAN IP in Termux with: ifconfig | grep wlan0 -A1
#
# Prints a random https://*.trycloudflare.com URL - copy it into the
# Worker's PHONE_URL variable in the Cloudflare dashboard.

param(
    [Parameter(Mandatory = $true)]
    [string]$PhoneIp
)

cloudflared tunnel --url "http://${PhoneIp}:8080"
