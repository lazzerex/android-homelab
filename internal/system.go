package internal

import (
	"os"
	"runtime"
	"time"
)

var Started=time.Now()

func SystemInfo() map[string]any{
	h,_:=os.Hostname()
	return map[string]any{
		"go_version":runtime.Version(),
		"goos":runtime.GOOS,
		"goarch":runtime.GOARCH,
		"hostname":h,
		"uptime":time.Since(Started).String(),
	}
}
