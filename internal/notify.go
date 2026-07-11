package internal

import (
	"context"
	"log"
	"os/exec"
	"time"
)

// no-op if termux-api isn't installed (also true on non-Termux dev machines)
func Notify(title, message string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if path, err := exec.LookPath("termux-notification"); err == nil {
			if err := exec.CommandContext(ctx, path, "-t", title, "-c", message).Run(); err != nil {
				log.Printf("termux-notification failed: %v", err)
			}
		}

		if path, err := exec.LookPath("termux-vibrate"); err == nil {
			if err := exec.CommandContext(ctx, path, "-d", "200").Run(); err != nil {
				log.Printf("termux-vibrate failed: %v", err)
			}
		}
	}()
}
