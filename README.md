# Android Homelab

A proof-of-concept project demonstrating that a modern Go application can be cross-compiled on Windows and executed on a legacy Android device through Termux, and that the result can be exposed to the public internet with automatic fallback when the phone is unreachable.

## Device

- OPPO A37f
- Android 5.1.1
- ARM64
- Termux (legacy)

## Current Architecture

Used to host the [Lazzerex Blog API](../lazzerex-blog) as its primary backend, falling back to Render when the phone is down.

```
Reader's browser
      ↓
Blog frontend (Vercel), calls PUBLIC_GO_API_BASE_URL
      ↓
Cloudflare Worker (blog-router.*.workers.dev)
      ↓ health-checks phone on every request
      ├─ phone healthy → proxy to phone
      └─ phone down/timeout → proxy to Render
      ↓
Cloudflare Tunnel (runs on the laptop, see Difficulties below)
      ↓ LAN
Phone (Termux) running the blog's cmd/server binary
```

The blog frontend only ever talks to the Worker. It never knows Render or the phone exist directly, the Worker decides per request.

Setup details and manual steps: `cloudflare/worker/wrangler.toml.example`, `scripts/run-tunnel.ps1` / `run-tunnel.sh`.

## Full Setup Guide

Step to reproduce this, assuming you already have a Go backend you want to host.

### Prerequisites

- An old Android phone, any Android version
- [Termux](https://f-droid.org/en/packages/com.termux/) installed from F-Droid (not the Play Store version, it's outdated and unmaintained)
- A laptop with Go installed
- A free [Cloudflare](https://dash.cloudflare.com/sign-up) account
- Optional: a second always-on backend (e.g. Render free tier) to fall back to

### 1. Set up the phone

In Termux:

```bash
pkg update
pkg install golang
```

### 2. Cross-compile your Go binary on the laptop

Target `linux/arm64`, not `android/arm64`. Android as a `GOOS` target crashes at runtime on real devices, `linux` works fine under Termux.

```powershell
$env:GOOS="linux"; $env:GOARCH="arm64"
go build -o server ./cmd/server
```

(See `scripts/build-linux-arm64.ps1` for a ready-made version of this.)

### 3. Transfer the binary to the phone

Any method works: `adb push`, a cloud drive, USB file transfer, `scp` if you've set up SSH in Termux. Land it anywhere in Termux's home directory.

### 4. Run it

```bash
chmod +x server
./server
```

Confirm it works locally first:

```bash
curl http://localhost:8080/health
```

Then from the laptop, confirm LAN access using the phone's LAN IP (find it in Termux with `ifconfig | grep wlan0 -A1`):

```bash
curl http://PHONE_IP:8080/health
```

### 5. Expose the phone to the internet

`cloudflared` does not run on old Android (see Difficulties below), so this runs on the laptop instead, proxying to the phone over LAN.

Install `cloudflared` on the laptop:

```powershell
winget install --id Cloudflare.cloudflared
```

Run the tunnel, pointed at the phone's LAN IP:

```powershell
.\scripts\run-tunnel.ps1 <phone-lan-ip>
```

This prints a random `https://*.trycloudflare.com` URL. Copy it, you'll need it in step 7. It changes every time you restart the tunnel.

### 6. Deploy the Cloudflare Worker

The Worker is what the public actually talks to. It health-checks the phone on every request and falls back to a second backend if the phone doesn't respond.

```bash
npm install -g wrangler
wrangler login
cd cloudflare/worker
cp wrangler.toml.example wrangler.toml
```

Edit `wrangler.toml`, set `RENDER_URL` (or whatever your fallback backend's URL is) to a real value. Then:

```bash
wrangler deploy
```

This gives you a stable URL like `https://blog-router.<your-subdomain>.workers.dev`.

### 7. Set the phone's URL

Cloudflare dashboard, Workers & Pages, your worker, Settings, Variables, add `PHONE_URL` = the `trycloudflare.com` URL from step 5. Save.

Every time the tunnel restarts (laptop reboot, phone reboot, crash), it gets a new URL, come back here and update it.

### 8. Point your frontend at the Worker

Wherever your frontend's backend URL is configured (env var, config file), set it to the Worker URL from step 6, not the phone or the fallback backend directly. Redeploy the frontend.

### 9. Test it

```bash
curl https://blog-router.<your-subdomain>.workers.dev/health
```

Should hit the phone (check Termux logs to confirm). Then kill the phone server or the tunnel and run the same `curl` again, it should still return `200`, now from the fallback backend instead.

## Difficulties & Workarounds

**`GOOS=android` crashes at runtime.** Cross-compiling with `GOOS=linux GOARCH=arm64` instead works perfectly on Termux, despite the device being Android. This is now the standard build target for anything deployed here.

**`cloudflared` doesn't run on Android 5.1.** The binary has runtime requirements the OS doesn't meet. Workaround: run the Cloudflare Tunnel on the laptop instead of the phone, pointed at the phone's LAN IP (`cloudflared tunnel --url http://PHONE_IP:8080`). The phone still does all the actual serving, the laptop is just a relay. Tradeoff: the phone is only reachable publicly while the laptop is also on and on the same LAN. Acceptable since the Render fallback covers the gap regardless.

**Quick Tunnel URLs are not stable.** Using Cloudflare's free Quick Tunnel (no domain/account required) means a new random `*.trycloudflare.com` URL every time the tunnel restarts (laptop reboot, phone reboot, crash). There's no code fix for this yet, the Worker's `PHONE_URL` variable has to be updated by hand in the Cloudflare dashboard after each restart.

## Build

Windows

```powershell
.\scripts\build-linux-arm64.ps1
```

Linux/macOS

```bash
./scripts/build-android.sh
```

## Example Service

The endpoints below belong to this repo's own toy service (`cmd/server`), used for experimenting without touching the real blog deployment.

| Endpoint | Description |
|----------|-------------|
| / | Welcome message |
| /health | Health check |
| /info | Build & runtime information |
| /time | Current server time |

```json
{
  "project": "android-go-server-test",
  "version": "v0.1.0",
  "branch": "main",
  "commit": "8d4b7f1",
  "built_at": "2026-07-08T22:15:11+07:00",
  "go_version": "go1.22.5",
  "goos": "linux",
  "goarch": "arm64",
  "hostname": "localhost",
  "uptime": "15m"
}
```
