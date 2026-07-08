package internal

import (
	"os"
	"runtime"
	"time"
)

var Started = time.Now()

func SystemInfo() map[string]any {
	host, _ := os.Hostname()

	return map[string]any{
		"project":  Project,
		"version":  Version,
		"branch":   Branch,
		"commit":   Commit,
		"built_at": BuiltAt,

		"go_version": runtime.Version(),
		"goos":       runtime.GOOS,
		"goarch":     runtime.GOARCH,

		"hostname": host,
		"uptime":   time.Since(Started).String(),
	}
}
