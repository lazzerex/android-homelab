package internal

import (
	"bufio"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// LiveMetrics reads Linux's /proc for real device stats when available
// (true on the phone under Termux), and falls back to Go runtime-only
// stats elsewhere (e.g. running this on Windows during development).
func LiveMetrics() map[string]any {
	m := map[string]any{
		"timestamp":  time.Now().Format(time.RFC3339),
		"uptime":     time.Since(Started).String(),
		"goroutines": runtime.NumGoroutine(),
	}

	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	m["go_heap_alloc_mb"] = ms.Alloc / 1024 / 1024

	if load, ok := readLoadAvg(); ok {
		m["load_avg_1m"] = load
	}

	if total, available, ok := readMemInfo(); ok {
		m["mem_total_mb"] = total / 1024
		m["mem_available_mb"] = available / 1024
		m["mem_used_percent"] = float64(total-available) / float64(total) * 100
	}

	return m
}

func readLoadAvg() (float64, bool) {
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return 0, false
	}
	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return 0, false
	}
	v, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

func readMemInfo() (total, available uint64, ok bool) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, 0, false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "MemTotal:"):
			total = parseMemInfoKB(line)
		case strings.HasPrefix(line, "MemAvailable:"):
			available = parseMemInfoKB(line)
		}
	}
	if total == 0 {
		return 0, 0, false
	}
	return total, available, true
}

func parseMemInfoKB(line string) uint64 {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return 0
	}
	v, _ := strconv.ParseUint(fields[1], 10, 64)
	return v
}
