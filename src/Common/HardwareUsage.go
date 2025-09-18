package common

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var (
	startTime      = time.Now()
	lastCPUTime    = time.Duration(0)
	lastSystemTime = time.Now()
)

// GetCPUUsage returns the current CPU usage percentage
func GetCPUUsage() float64 {
	switch runtime.GOOS {
	case "linux":
		return getCPUUsageLinux()
	case "windows":
		return getCPUUsageWindows()
	case "darwin":
		return getCPUUsageMacOS()
	default:
		// Fallback: return Go runtime CPU usage
		return getGoCPUUsage()
	}
}

// getCPUUsageLinux reads CPU usage from /proc/stat on Linux
func getCPUUsageLinux() float64 {
	file, err := os.Open("/proc/stat")
	if err != nil {
		return 0.0
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return 0.0
	}

	line := scanner.Text()
	fields := strings.Fields(line)
	if len(fields) < 8 || fields[0] != "cpu" {
		return 0.0
	}

	var times []uint64
	for i := 1; i <= 7; i++ {
		val, err := strconv.ParseUint(fields[i], 10, 64)
		if err != nil {
			return 0.0
		}
		times = append(times, val)
	}

	// Calculate CPU usage
	idle := times[3] + times[4] // idle + iowait
	total := uint64(0)
	for _, t := range times {
		total += t
	}

	if total == 0 {
		return 0.0
	}

	return float64(total-idle) / float64(total) * 100.0
}

func getCPUUsageWindows() float64 {
	return getGoCPUUsage()
}

func getCPUUsageMacOS() float64 {
	return getGoCPUUsage()
}

func getGoCPUUsage() float64 {
	now := time.Now()
	cpuTime := time.Duration(runtime.NumCPU()) * now.Sub(lastSystemTime)

	if lastCPUTime > 0 && cpuTime > 0 {
		usage := float64(cpuTime-lastCPUTime) / float64(cpuTime) * 100.0
		lastCPUTime = cpuTime
		lastSystemTime = now

		// Clamp between 0 and 100
		if usage < 0 {
			usage = 0
		} else if usage > 100 {
			usage = 100
		}

		return usage
	}

	lastCPUTime = cpuTime
	lastSystemTime = now
	return 0.0
}

func GetMemoryUsage() (uint64, uint64, error) {
	switch runtime.GOOS {
	case "linux":
		return getMemoryUsageLinux()
	case "windows":
		return getMemoryUsageWindows()
	case "darwin":
		return getMemoryUsageMacOS()
	default:
		return getMemoryUsageRuntime()
	}
}

func getMemoryUsageLinux() (uint64, uint64, error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	var memTotal, memFree, memBuffers, memCached uint64
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		switch fields[0] {
		case "MemTotal:":
			if val, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
				memTotal = val * 1024 // Convert from KB to bytes
			}
		case "MemFree:":
			if val, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
				memFree = val * 1024
			}
		case "Buffers:":
			if val, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
				memBuffers = val * 1024
			}
		case "Cached:":
			if val, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
				memCached = val * 1024
			}
		}
	}

	memUsed := memTotal - memFree - memBuffers - memCached
	return memUsed, memTotal, nil
}

func getMemoryUsageWindows() (uint64, uint64, error) {
	return getMemoryUsageRuntime()
}

func getMemoryUsageMacOS() (uint64, uint64, error) {
	return getMemoryUsageRuntime()
}

func getMemoryUsageRuntime() (uint64, uint64, error) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	totalMem := m.Sys * 4
	if totalMem < m.Sys {
		totalMem = m.Sys * 2
	}

	return m.Alloc, totalMem, nil
}

func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

func GetCPUCount() int {
	return runtime.NumCPU()
}

func GetGoVersion() string {
	return runtime.Version()
}

func GetOS() string {
	return runtime.GOOS
}

func GetArch() string {
	return runtime.GOARCH
}
