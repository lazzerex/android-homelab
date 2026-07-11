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

Setup details and manual steps: `cloudflare/worker/wrangler.toml.example`, `scripts/run-tunnel-auto.ps1` / `run-tunnel-auto.sh` (recommended — also updates the Worker), or `scripts/run-tunnel.ps1` / `run-tunnel.sh` (tunnel only, manual Worker update).

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

### 5. Deploy the Cloudflare Worker

The Worker is what the public actually talks to. It health-checks the phone on every request and falls back to a second backend if the phone doesn't respond.

```bash
npm install -g wrangler
wrangler login
cd cloudflare/worker
cp wrangler.toml.example wrangler.toml
```

Edit `wrangler.toml`, set `RENDER_URL` (or whatever your fallback backend's URL is) to a real value. Leave `PHONE_URL` unset for now, step 6 fills it in. Then:

```bash
wrangler deploy
```

This gives you a stable URL like `https://blog-router.<your-subdomain>.workers.dev`.

### 6. Expose the phone to the internet

`cloudflared` does not run on old Android (see Difficulties below), so this runs on the laptop instead, proxying to the phone over LAN.

Install `cloudflared` on the laptop:

```powershell
winget install --id Cloudflare.cloudflared
```

Run the tunnel with the auto-update script, pointed at the phone's LAN IP:

```powershell
.\scripts\run-tunnel-auto.ps1 <phone-lan-ip>
```

```bash
./scripts/run-tunnel-auto.sh <phone-lan-ip>
```

This starts the tunnel, scrapes the random `https://*.trycloudflare.com` URL from its own logs, writes it into `wrangler.toml` as `PHONE_URL`, and runs `wrangler deploy` for you. No dashboard step needed. It changes every time you restart the tunnel — just rerun the script, no manual URL copying.

(`scripts/run-tunnel.ps1` / `run-tunnel.sh` still exist if you want the tunnel without the auto-update, e.g. for debugging — then you'd set `PHONE_URL` by hand in the Cloudflare dashboard, Workers & Pages → your worker → Settings → Variables.)

### 7. Point your frontend at the Worker

Wherever your frontend's backend URL is configured (env var, config file), set it to the Worker URL from step 5, not the phone or the fallback backend directly. Redeploy the frontend.

### 8. Test it

```bash
curl https://blog-router.<your-subdomain>.workers.dev/health
```

Should hit the phone (check Termux logs to confirm). Then kill the phone server or the tunnel and run the same `curl` again, it should still return `200`, now from the fallback backend instead.

## Difficulties & Workarounds

**`GOOS=android` crashes at runtime.** Cross-compiling with `GOOS=linux GOARCH=arm64` instead works perfectly on Termux, despite the device being Android. This is now the standard build target for anything deployed here.

**`cloudflared` doesn't run on Android 5.1.** The binary has runtime requirements the OS doesn't meet. Workaround: run the Cloudflare Tunnel on the laptop instead of the phone, pointed at the phone's LAN IP (`cloudflared tunnel --url http://PHONE_IP:8080`). The phone still does all the actual serving, the laptop is just a relay. Tradeoff: the phone is only reachable publicly while the laptop is also on and on the same LAN. Acceptable since the Render fallback covers the gap regardless.

**Quick Tunnel URLs are not stable.** Using Cloudflare's free Quick Tunnel (no domain/account required) means a new random `*.trycloudflare.com` URL every time the tunnel restarts (laptop reboot, phone reboot, crash). `scripts/run-tunnel-auto.ps1` / `.sh` scrapes the new URL from the tunnel's own log output and redeploys the Worker automatically, so this no longer needs a manual dashboard edit. A named tunnel with a fixed hostname would remove the redeploy-on-restart step entirely, but that needs a domain added to the Cloudflare account, which this project doesn't have yet.

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
| /health | Health check. Also fires a Termux notification + vibrate if `termux-api` is installed (no-op otherwise) |
| /info | Build & runtime information |
| /time | Current server time |
| /stream | Server-Sent Events: live stats (load avg, memory, goroutines) once a second |
| /dashboard | Browser page that renders /stream live |
| /checksum | POST a body, get its FNV-1a checksum back. Computed via C through cgo when built natively with clang, pure Go otherwise - response's `backend` field says which |
| GET /guestbook | List guestbook entries, persisted in a local SQLite file (`GUESTBOOK_DB_PATH`, default `guestbook.db`) |
| POST /guestbook | Add an entry - JSON body `{"name": "...", "message": "..."}` |
| POST /pastebin | Create a paste from the request body, returns an id + url |
| GET /pastebin/{id} | Fetch a paste by id (in-memory, lost on restart - unlike the guestbook) |

All routes are rate-limited per client IP (5 req/s, burst 10) - hit it faster and you'll get `429 Too Many Requests`.

For the notification on `/health`: install the [Termux:API](https://f-droid.org/en/packages/com.termux.api/) app plus `pkg install termux-api` in Termux. Without it, `/health` just skips the notification silently.

## Experiments

Polyglot learning exercises living alongside the main Go server. All optional - the normal build/deploy flow above doesn't touch or need any of this.

- `experiments/c-http-server/` - HTTP server from raw POSIX sockets, no libraries. Shows what `net/http` hides: the accept loop, manual request parsing, hand-written response framing.
- `experiments/rust-metrics/` - standalone binary that reads `/proc/loadavg` and `/proc/meminfo`, prints one line of JSON, exits.
- `internal/cbridge/` - FNV-1a checksum, implemented twice: once via cgo calling real C (`bridge_cgo.go`, needs `CGO_ENABLED=1` + a C compiler), once in pure Go (`bridge_fallback.go`, what the normal Windows cross-compile build uses, no setup needed). Backs the `/checksum` endpoint.

### Running them

Everything below runs **on the phone, in Termux** - native compilers, no cross-toolchain setup.

**Get the source onto the phone**, one of:

```bash
pkg install git
git clone <your-repo-url>
```

or, if you'd rather transfer files the same way you move built binaries (Downloads folder):

```bash
termux-setup-storage        # one-time, grants Termux access to shared storage
mv ~/storage/downloads/android-homelab ~/android-homelab   # after copying the folder into Downloads
```

**1. C raw-socket server**

```bash
pkg install clang
cd android-homelab/experiments/c-http-server
clang -O2 -o c-http-server server.c
./c-http-server &            # listens on :8081, separate from the Go server's :8080
curl http://localhost:8081/health
```

**2. cgo bridge** (`/checksum`'s `backend` field flips `go` → `cgo`)

```bash
cd android-homelab
export CGO_ENABLED=1
export GOOS=linux GOARCH=arm64
go build -o server-cgo ./cmd/server
./server-cgo &
curl -X POST -d "hello" http://localhost:8080/checksum
```

**3. Rust subprocess helper** (`/stream`'s `metrics_source` flips `go-native` → `helper`)

```bash
pkg install rust
cd android-homelab/experiments/rust-metrics
cargo build --release
export METRICS_HELPER_BIN=$(pwd)/target/release/rust-metrics
cd ../..
./server &            # or ./server-cgo
```

Then open `http://PHONE_IP:8080/dashboard` from the laptop - cards now include `helper: load_avg_1m`, `helper: mem_total_mb`, etc.

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
