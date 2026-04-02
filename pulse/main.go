package main

import (
	"bufio"
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

//go:embed dashboard.html
var dashboardHTML embed.FS

// ── Structs ────────────────────────────────────────────────────────────────

type MemStats struct {
	TotalMB     int     `json:"total_mb"`
	UsedMB      int     `json:"used_mb"`
	FreeMB      int     `json:"free_mb"`
	UsedPercent float64 `json:"used_percent"`
}

type ContainerStat struct {
	Name    string `json:"name"`
	CPU     string `json:"cpu"`
	MemUsed string `json:"mem_used"`
	MemPerc string `json:"mem_perc"`
	Status  string `json:"status"`
}

type Metrics struct {
	Timestamp  string          `json:"timestamp"`
	CPUPercent float64         `json:"cpu_percent"`
	Memory     MemStats        `json:"memory"`
	Containers []ContainerStat `json:"containers"`
	Uptime     string          `json:"uptime"`
}

// ── Cache ──────────────────────────────────────────────────────────────────

var (
	cachedMetrics Metrics
	cacheMu       sync.RWMutex
)

func startCollector(interval time.Duration) {
	collect := func() {
		m := Metrics{
			Timestamp:  time.Now().Format("2006-01-02 15:04:05"),
			CPUPercent: getCPUPercent(),
			Memory:     getMemStats(),
			Containers: getContainers(),
			Uptime:     getUptime(),
		}
		cacheMu.Lock()
		cachedMetrics = m
		cacheMu.Unlock()
	}

	collect() // primera lectura inmediata al arrancar
	go func() {
		for range time.Tick(interval) {
			collect()
		}
	}()
}

// ── CPU ────────────────────────────────────────────────────────────────────

type cpuTimes struct {
	user, nice, system, idle, iowait, irq, softirq, steal uint64
}

func readCPUTimes() (cpuTimes, error) {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return cpuTimes{}, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "cpu ") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 8 {
			break
		}
		var t cpuTimes
		vals := []*uint64{&t.user, &t.nice, &t.system, &t.idle, &t.iowait, &t.irq, &t.softirq, &t.steal}
		for i, v := range vals {
			*v, _ = strconv.ParseUint(fields[i+1], 10, 64)
		}
		return t, nil
	}
	return cpuTimes{}, fmt.Errorf("cpu line not found")
}

func getCPUPercent() float64 {
	t1, err := readCPUTimes()
	if err != nil {
		return 0
	}
	time.Sleep(200 * time.Millisecond)
	t2, err := readCPUTimes()
	if err != nil {
		return 0
	}

	idle1 := t1.idle + t1.iowait
	idle2 := t2.idle + t2.iowait
	total1 := t1.user + t1.nice + t1.system + idle1 + t1.irq + t1.softirq + t1.steal
	total2 := t2.user + t2.nice + t2.system + idle2 + t2.irq + t2.softirq + t2.steal

	totalDiff := float64(total2 - total1)
	idleDiff := float64(idle2 - idle1)

	if totalDiff == 0 {
		return 0
	}
	return (totalDiff - idleDiff) / totalDiff * 100
}

// ── Memory ─────────────────────────────────────────────────────────────────

func getMemStats() MemStats {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return MemStats{}
	}
	defer f.Close()

	data := map[string]uint64{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 2 {
			key := strings.TrimSuffix(fields[0], ":")
			val, _ := strconv.ParseUint(fields[1], 10, 64)
			data[key] = val
		}
	}

	totalKB := data["MemTotal"]
	availKB := data["MemAvailable"]
	usedKB := totalKB - availKB

	totalMB := int(totalKB / 1024)
	usedMB := int(usedKB / 1024)
	freeMB := int(availKB / 1024)

	var pct float64
	if totalMB > 0 {
		pct = float64(usedMB) / float64(totalMB) * 100
	}

	return MemStats{
		TotalMB:     totalMB,
		UsedMB:      usedMB,
		FreeMB:      freeMB,
		UsedPercent: pct,
	}
}

// ── Docker ─────────────────────────────────────────────────────────────────

func getContainers() []ContainerStat {
	statsOut, err := exec.Command(
		"docker", "stats", "--no-stream", "--format",
		"{{.Name}}|{{.CPUPerc}}|{{.MemUsage}}|{{.MemPerc}}",
	).Output()

	containers := []ContainerStat{}
	if err != nil {
		return containers
	}

	lines := strings.Split(strings.TrimSpace(string(statsOut)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) < 4 {
			continue
		}
		memParts := strings.Split(parts[2], " / ")
		memUsed := parts[2]
		if len(memParts) >= 1 {
			memUsed = strings.TrimSpace(memParts[0])
		}
		containers = append(containers, ContainerStat{
			Name:    parts[0],
			CPU:     parts[1],
			MemUsed: memUsed,
			MemPerc: parts[3],
			Status:  "running",
		})
	}
	return containers
}

// ── Uptime ─────────────────────────────────────────────────────────────────

func getUptime() string {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return "unknown"
	}
	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return "unknown"
	}
	secs, _ := strconv.ParseFloat(fields[0], 64)
	days := int(secs) / 86400
	hours := (int(secs) % 86400) / 3600
	mins := (int(secs) % 3600) / 60
	return fmt.Sprintf("%dd %dh %dm", days, hours, mins)
}

// ── Handlers ───────────────────────────────────────────────────────────────

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	cacheMu.RLock()
	m := cachedMetrics
	cacheMu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(m)
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	data, err := dashboardHTML.ReadFile("dashboard.html")
	if err != nil {
		http.Error(w, "dashboard not found", 500)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(data)
}

// ── Main ───────────────────────────────────────────────────────────────────

func main() {
	port := "2019"
	if p := os.Getenv("PULSE_PORT"); p != "" {
		port = p
	}

	intervalSecs := 3
	if s := os.Getenv("PULSE_INTERVAL"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			intervalSecs = n
		}
	}

	fmt.Printf("Nebula Pulse arrancando...\n")
	fmt.Printf("  Intervalo de recoleccion: %ds\n", intervalSecs)

	startCollector(time.Duration(intervalSecs) * time.Second)

	http.HandleFunc("/", dashboardHandler)
	http.HandleFunc("/metrics", metricsHandler)

	fmt.Printf("  Dashboard -> http://localhost:%s\n", port)
	fmt.Printf("  Metrics   -> http://localhost:%s/metrics\n\n", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
