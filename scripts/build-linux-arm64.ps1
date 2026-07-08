$Project = "android-homelab"
$Version = "v0.1.0"
$Commit = git rev-parse --short HEAD
$Branch = git branch --show-current
$BuiltAt = Get-Date -Format "yyyy-MM-ddTHH:mm:ssK"

$env:GOOS="linux"
$env:GOARCH="arm64"
$env:CGO_ENABLED="0"

go build `
-ldflags "
-X 'android-homelab/internal.Version=$Version'
-X 'android-homelab/internal.Commit=$Commit'
-X 'android-homelab/internal.Branch=$Branch'
-X 'android-homelab/internal.BuiltAt=$BuiltAt'
" `
-o server `
./cmd/server