package backend

import (
	"context"
	"sync"
	"time"
)

// NetworkController serves as the Wails binding layer (Controller pattern).
// It wraps ScannerService, FirewallService, and ThroughputMonitor, exposing
// thread-safe methods to the frontend.
// Reference: Service-Controller pattern as described in Go standard project layout guidelines.
type NetworkController struct {
	scanner   *ScannerService
	firewall  *FirewallService
	throughput *ThroughputMonitor
	scanMu    sync.Mutex
}

// NewNetworkController creates a new controller with the singleton scanner and firewall.
func NewNetworkController() *NetworkController {
	nc := &NetworkController{
		scanner:   NewScannerService(),
		firewall:  NewFirewallService(),
		throughput: NewThroughputMonitor(),
	}
	nc.throughput.Start()
	return nc
}

// Start initializes the scanner lifecycle.
func (nc *NetworkController) Start(ctx context.Context) {
	nc.scanner.Start(ctx)
}

// Stop cleanly shuts down the scanner and throughput monitor.
func (nc *NetworkController) Stop() {
	nc.scanner.Stop()
	nc.throughput.Stop()
}

// ScanNetwork is the main binding method called from the frontend.
// It returns a ScanResult with all interfaces, addresses, and connections.
// Thread-safe: uses a mutex to prevent concurrent scans.
func (nc *NetworkController) ScanNetwork(opts ScanOptions) *ScanResult {
	nc.scanMu.Lock()
	defer nc.scanMu.Unlock()

	ctx := context.Background()
	if opts.Timeout <= 0 || opts.Timeout > 60 {
		opts.Timeout = 15
	}

	result, err := nc.scanner.Scan(ctx, opts)
	if err != nil {
		return &ScanResult{
			Timestamp:  time.Now().UnixMilli(),
			Error:      err.Error(),
			DurationMs: 0,
		}
	}

	return result
}

// QuickScan is a convenience binding that scans with default options.
func (nc *NetworkController) QuickScan() *ScanResult {
	return nc.ScanNetwork(ScanOptions{
		IncludeLoopback: false,
		Timeout:         15,
	})
}

// GetPlatform returns the current operating system name for UI adaptation.
func (nc *NetworkController) GetPlatform() string {
	return "windows"
}

// GetThroughput returns real-time transfer rates for all interfaces.
func (nc *NetworkController) GetThroughput() []ThroughputData {
	return nc.throughput.GetThroughput()
}

// GetInterfaceThroughput returns throughput for a single named interface.
func (nc *NetworkController) GetInterfaceThroughput(name string) (ThroughputData, bool) {
	return nc.throughput.GetInterfaceThroughput(name)
}

// Firewall - Blocking operations

// BlockTarget blocks an IP address or PID based on the request.
func (nc *NetworkController) BlockTarget(req BlockRequest) *BlockResult {
	admin := nc.firewall.IsAdmin()
	if !admin {
		return &BlockResult{
			Success: false,
			Message: "Administrator privileges required. Please run as Administrator.",
			Action:  "block",
			IsAdmin: false,
		}
	}

	switch req.Type {
	case "ip":
		entry, err := nc.firewall.BlockIP(req.Target, req.Interface)
		if err != nil {
			return &BlockResult{
				Success: false,
				Message: err.Error(),
				Action:  "block",
				IsAdmin: true,
			}
		}
		return &BlockResult{
			Success: true,
			Message: "IP " + req.Target + " blocked successfully",
			Action:  "block",
			Entry:   entry,
			IsAdmin: true,
		}

	case "pid":
		entries, err := nc.firewall.BlockPID(req.PID, nil)
		if err != nil {
			return &BlockResult{
				Success: false,
				Message: err.Error(),
				Action:  "block",
				IsAdmin: true,
			}
		}
		return &BlockResult{
			Success: true,
			Message: "PID " + string(rune(req.PID)) + " blocked (" + string(rune(len(entries))) + " rules)",
			Action:  "block",
			IsAdmin: true,
		}

	default:
		return &BlockResult{
			Success: false,
			Message: "Unknown block type: " + req.Type,
			Action:  "block",
			IsAdmin: admin,
		}
	}
}

// UnblockTarget removes a block rule by its ID.
func (nc *NetworkController) UnblockTarget(blockID string) *BlockResult {
	admin := nc.firewall.IsAdmin()
	if !admin {
		return &BlockResult{
			Success: false,
			Message: "Administrator privileges required.",
			Action:  "unblock",
			IsAdmin: false,
		}
	}

	// Parse the block ID to determine type
	if len(blockID) > 3 && blockID[:3] == "ip_" {
		ip := blockID[3:]
		err := nc.firewall.UnblockIP(ip)
		if err != nil {
			return &BlockResult{
				Success: false, Message: err.Error(), Action: "unblock", IsAdmin: true,
			}
		}
		return &BlockResult{
			Success: true, Message: "Unblocked IP " + ip, Action: "unblock", IsAdmin: true,
		}
	}

	return &BlockResult{
		Success: false, Message: "Unknown block ID: " + blockID, Action: "unblock", IsAdmin: admin,
	}
}

// GetBlocked returns a list of all currently blocked entries.
func (nc *NetworkController) GetBlocked() []BlockedEntry {
	return nc.firewall.GetBlocked()
}

// IsAdmin checks if the app has administrator privileges.
func (nc *NetworkController) IsAdmin() bool {
	return nc.firewall.IsAdmin()
}
