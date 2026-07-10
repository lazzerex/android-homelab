# Android Homelab

A proof-of-concept project demonstrating that a modern Go application can be cross-compiled on Windows and executed on a legacy Android device through Termux — and that the result can be exposed to the public internet with automatic fallback when the phone is unreachable.

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
Blog frontend (Vercel) — calls PUBLIC_GO_API_BASE_URL
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

The blog frontend only ever talks to the Worker. It never knows Render or the phone exist directly — the Worker decides per request.

Setup details and manual steps: `cloudflare/worker/wrangler.toml.example`, `scripts/run-tunnel.ps1` / `run-tunnel.sh`.

## Difficulties & Workarounds

**`GOOS=android` crashes at runtime.** Discovered during Milestone 1. Cross-compiling with `GOOS=linux GOARCH=arm64` instead works perfectly on Termux, despite the device being Android. This is now the standard build target for anything deployed here.

**`cloudflared` doesn't run on Android 5.1.** The binary has runtime requirements the OS doesn't meet. Workaround: run the Cloudflare Tunnel on the laptop instead of the phone, pointed at the phone's LAN IP (`cloudflared tunnel --url http://PHONE_IP:8080`). The phone still does all the actual serving — the laptop is just a relay. Tradeoff: the phone is only reachable publicly while the laptop is also on and on the same LAN. Acceptable since the Render fallback covers the gap regardless.

**Quick Tunnel URLs are not stable.** Using Cloudflare's free Quick Tunnel (no domain/account required) means a new random `*.trycloudflare.com` URL every time the tunnel restarts (laptop reboot, phone reboot, crash). There's no code fix for this yet — the Worker's `PHONE_URL` variable has to be updated by hand in the Cloudflare dashboard after each restart. Candidate for future deployment automation.

**Don't commit real backend URLs.** `cloudflare/worker/wrangler.toml` is gitignored because it holds the real `RENDER_URL`. `wrangler.toml.example` is the committed template — copy it to `wrangler.toml` and fill in your own value before running `wrangler deploy`. This isn't hiding a credential (the URL was already client-visible in the old setup), it just avoids inviting direct traffic that bypasses the Worker and burns Render's free tier.

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

## Related

- Milestones and progress: `MILESTONES.md`
- Long-term vision and phases: `PROJECT_ROADMAP.md`
