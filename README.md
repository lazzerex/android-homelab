# Android Homelab

A proof-of-concept project demonstrating that a modern Go application can be cross-compiled on Windows and executed on a legacy Android device through Termux.

## Device

- OPPO A37f
- Android 5.1.1
- ARM64
- Termux (legacy)

## Build

Windows

```powershell
.\scripts\build-linux-arm64.ps1
```

Linux/macOS

```bash
./scripts/build-android.sh
```

## API

| Endpoint | Description |
|----------|-------------|
| / | Welcome message |
| /health | Health check |
| /info | Build & runtime information |
| /time | Current server time |

## Example

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