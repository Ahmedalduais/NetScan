package backend

import (
	"os"
	"runtime"
	"sort"

	"github.com/shirou/gopsutil/v3/process"
)

// GetProcessDetails returns all unique processes with their connection counts
// and estimated bandwidth usage. Throughput data is used to proportionally
// allocate per-interface bandwidth to each PID based on connection count.
func GetProcessDetails(interfaces []InterfaceInfo, throughput []ThroughputData) []ProcessDetail {
	type pidIface struct {
		pid   int32
		iface string
	}
	pidIfaceCounts := make(map[pidIface]int)
	pidMap := make(map[int32]*ProcessDetail)

	// Count connections per (PID, interface)
	for _, iface := range interfaces {
		for _, conn := range iface.Connections {
			if conn.PID <= 0 {
				continue
			}
			key := pidIface{pid: conn.PID, iface: iface.Name}
			pidIfaceCounts[key]++

			if _, ok := pidMap[conn.PID]; !ok {
				exePath := getProcessExe(conn.PID)
				pidMap[conn.PID] = &ProcessDetail{
					PID:     conn.PID,
					Name:    conn.ProcessName,
					ExePath: exePath,
				}
			}
			pidMap[conn.PID].ConnectionCount++
		}
	}

	// Build interface -> total connections map
	ifaceTotalConns := make(map[string]int)
	for key, count := range pidIfaceCounts {
		ifaceTotalConns[key.iface] += count
	}

	// Allocate bandwidth proportionally
	tpMap := make(map[string]ThroughputData)
	for _, td := range throughput {
		tpMap[td.Interface] = td
	}

	for key, count := range pidIfaceCounts {
		totalConns := ifaceTotalConns[key.iface]
		if totalConns == 0 {
			continue
		}
		share := float64(count) / float64(totalConns)

		if td, ok := tpMap[key.iface]; ok {
			pd := pidMap[key.pid]
			pd.RXBytes += uint64(float64(td.RXBytes) * share)
			pd.TXBytes += uint64(float64(td.TXBytes) * share)
			pd.RXBps += td.RXBps * share
			pd.TXBps += td.TXBps * share
			pd.Estimated = true
		}
	}

	// Format bit rates
	for _, pd := range pidMap {
		pd.RXBpsStr = formatBits(pd.RXBps)
		pd.TXBpsStr = formatBits(pd.TXBps)
	}

	// Sort by PID
	var result []ProcessDetail
	for _, pd := range pidMap {
		result = append(result, *pd)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].PID < result[j].PID
	})

	return result
}

// getProcessExe returns the executable path for a given PID.
func getProcessExe(pid int32) string {
	p, err := process.NewProcess(pid)
	if err != nil {
		return ""
	}
	exe, err := p.Exe()
	if err != nil {
		return ""
	}
	return exe
}

// GetProcessIcon returns a base64-encoded PNG icon for a process.
// On Windows it extracts the icon from the executable.
// On other platforms it returns empty string.
func GetProcessIcon(pid int32) string {
	if runtime.GOOS != "windows" {
		return ""
	}

	exePath := getProcessExe(pid)
	if exePath == "" {
		return ""
	}

	if _, err := os.Stat(exePath); err != nil {
		return ""
	}

	return extractIconWindows(exePath)
}
