$Project = "android-go-server-test"
$Version = "v0.1.0"
$Commit = git rev-parse --short HEAD
$Branch = git branch --show-current
$BuiltAt = Get-Date -Format "yyyy-MM-ddTHH:mm:ssK"

$env:GOOS="linux"
$env:GOARCH="arm64"
$env:CGO_ENABLED="0"

go build `
-ldflags "
-X 'android-go-server-test/internal.Version=$Version'
-X 'android-go-server-test/internal.Commit=$Commit'
-X 'android-go-server-test/internal.Branch=$Branch'
-X 'android-go-server-test/internal.BuiltAt=$BuiltAt'
" `
-o server `
./cmd/server