package backend

import (
	"context"
	"fmt"
	"sync"
	"time"

	psnet "github.com/shirou/gopsutil/v3/net"
)

// ThroughputData holds the real-time transfer rates for one interface.
type ThroughputData struct {
	Interface string  `json:"interface"`
	RXBytes   uint64  `json:"rx_bytes"`
	TXBytes   uint64  `json:"tx_bytes"`
	RXPkts    uint64  `json:"rx_pkts"`
	TXPkts    uint64  `json:"tx_pkts"`
	RXBps     float64 `json:"rx_bps"`
	TXBps     float64 `json:"tx_bps"`
	RXBpsStr  string  `json:"rx_bps_str"`
	TXBpsStr  string  `json:"tx_bps_str"`
}

// ThroughputMonitor tracks network I/O rates by periodically polling gopsutil.
// Reference: gopsutil net.IOCounters - reads /proc/net/dev on Linux, 
// GetIfEntry2 on Windows (RFC 2863, IF-MIB).
type ThroughputMonitor struct {
	mu         sync.Mutex
	previous   map[string]psnet.IOCountersStat
	latest     map[string]ThroughputData
	cancel     context.CancelFunc
	running    bool
	pollPeriod time.Duration
}

// NewThroughputMonitor creates a new throughput monitor.
func NewThroughputMonitor() *ThroughputMonitor {
	return &ThroughputMonitor{
		pollPeriod: 1 * time.Second,
	}
}

// Start begins polling interface counters at the configured interval.
func (tm *ThroughputMonitor) Start() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tm.running {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	tm.cancel = cancel
	tm.running = true

	// Initial read to establish baseline
	counters, err := psnet.IOCounters(true)
	if err == nil {
		tm.previous = make(map[string]psnet.IOCountersStat)
		for _, c := range counters {
			tm.previous[c.Name] = c
		}
	}

	go tm.pollLoop(ctx)
}

func (tm *ThroughputMonitor) pollLoop(ctx context.Context) {
	ticker := time.NewTicker(tm.pollPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			tm.sample()
		}
	}
}

func (tm *ThroughputMonitor) sample() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	counters, err := psnet.IOCounters(true)
	if err != nil {
		return
	}

	current := make(map[string]psnet.IOCountersStat)
	for _, c := range counters {
		current[c.Name] = c
	}

	now := time.Now()
	if tm.latest == nil {
		tm.latest = make(map[string]ThroughputData)
	}

	for name, cur := range current {
		td := ThroughputData{
			Interface: name,
			RXBytes:   cur.BytesRecv,
			TXBytes:   cur.BytesSent,
			RXPkts:    cur.PacketsRecv,
			TXPkts:    cur.PacketsSent,
		}

		if prev, ok := tm.previous[name]; ok {
			elapsed := tm.pollPeriod.Seconds()
			if elapsed > 0 {
				rxDelta := float64(cur.BytesRecv - prev.BytesRecv)
				txDelta := float64(cur.BytesSent - prev.BytesSent)
				td.RXBps = rxDelta / elapsed * 8 // bits per second
				td.TXBps = txDelta / elapsed * 8
				td.RXBpsStr = formatBits(td.RXBps)
				td.TXBpsStr = formatBits(td.TXBps)
			}
		} else {
			td.RXBpsStr = "—"
			td.TXBpsStr = "—"
		}

		tm.latest[name] = td
		_ = now
	}

	tm.previous = current
}

// Stop halts the polling goroutine.
func (tm *ThroughputMonitor) Stop() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tm.cancel != nil {
		tm.cancel()
	}
	tm.running = false
}

// GetThroughput returns the latest throughput data for all interfaces.
func (tm *ThroughputMonitor) GetThroughput() []ThroughputData {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	var result []ThroughputData
	for _, td := range tm.latest {
		result = append(result, td)
	}
	return result
}

// GetInterfaceThroughput returns throughput for a specific interface.
func (tm *ThroughputMonitor) GetInterfaceThroughput(name string) (ThroughputData, bool) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	td, ok := tm.latest[name]
	return td, ok
}

// formatBits converts a bit rate to a human-readable string (e.g., "1.5 Mbps").
func formatBits(bps float64) string {
	switch {
	case bps >= 1_000_000_000:
		return fmt.Sprintf("%.2f Gbps", bps/1_000_000_000)
	case bps >= 1_000_000:
		return fmt.Sprintf("%.2f Mbps", bps/1_000_000)
	case bps >= 1_000:
		return fmt.Sprintf("%.2f Kbps", bps/1_000)
	default:
		return fmt.Sprintf("%.0f bps", bps)
	}
}
