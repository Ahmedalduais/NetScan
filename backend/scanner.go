package backend

import (
	"context"
	"fmt"
	"net"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	psnet "github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

// ScannerService implements the Singleton pattern for network scanning.
// Uses context.Context for cancellation and sync.WaitGroup for goroutine tracking
// to prevent resource leaks (Woodcock, A. (2018). Network Programming with Go).
type ScannerService struct {
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.Mutex
	wg     sync.WaitGroup
}

var (
	instance *ScannerService
	once     sync.Once
)

// NewScannerService returns the singleton instance of ScannerService.
// Thread-safe via sync.Once (Go stdlib guarantees).
func NewScannerService() *ScannerService {
	once.Do(func() {
		instance = &ScannerService{}
	})
	return instance
}

// Start initializes the scanner with a cancellable context.
func (s *ScannerService) Start(ctx context.Context) {
	s.ctx, s.cancel = context.WithCancel(ctx)
}

// Stop cancels the scanner context and waits for goroutines to finish.
func (s *ScannerService) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
	s.wg.Wait()
}

// Scan performs a full network scan: discovers interfaces and their connections.
// Reference: net.Interface from Go stdlib (RFC 2863, IF-MIB).
func (s *ScannerService) Scan(ctx context.Context, opts ScanOptions) (*ScanResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	start := time.Now()
	result := &ScanResult{
		Timestamp: time.Now().UnixMilli(),
	}

	if opts.Timeout <= 0 {
		opts.Timeout = 10
	}

	scanCtx, scanCancel := context.WithTimeout(ctx, time.Duration(opts.Timeout)*time.Second)
	defer scanCancel()

	_ = runtime.GOOS

	// Get interfaces using standard net library (portable across all platforms)
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to enumerate interfaces: %w", err)
	}

	// Get active connections via gopsutil (cross-platform, avoids shell commands)
	connections, connErr := psnet.Connections("all")
	if connErr != nil {
		result.PermissionError = isPermissionError(connErr)
	}

	// Build interface list
	var interfaceList []InterfaceInfo
	for _, iface := range interfaces {
		select {
		case <-scanCtx.Done():
			return nil, scanCtx.Err()
		default:
		}

		if !opts.IncludeLoopback && iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		info := InterfaceInfo{
			Name:       iface.Name,
			MACAddress: macToString(iface.HardwareAddr),
			MTU:        iface.MTU,
			IsUp:       iface.Flags&net.FlagUp != 0,
			IsLoopback: iface.Flags&net.FlagLoopback != 0,
		}

		// Collect flags
		if iface.Flags&net.FlagBroadcast != 0 {
			info.Flags = append(info.Flags, "BROADCAST")
		}
		if iface.Flags&net.FlagLoopback != 0 {
			info.Flags = append(info.Flags, "LOOPBACK")
		}
		if iface.Flags&net.FlagPointToPoint != 0 {
			info.Flags = append(info.Flags, "POINTTOPOINT")
		}
		if iface.Flags&net.FlagMulticast != 0 {
			info.Flags = append(info.Flags, "MULTICAST")
		}

		// Collect IP addresses from this interface
		info.IPAddresses = getInterfaceAddresses(&iface)

		// Find connections matching this interface's IPs
		info.Connections = matchingConnections(info.IPAddresses, connections)

		// Deduplicate connections by {local port, remote ip:port, protocol}
		info.Connections = dedupConnections(info.Connections)

		sort.Slice(info.Connections, func(i, j int) bool {
			return info.Connections[i].LocalPort < info.Connections[j].LocalPort
		})

		result.TotalConnections += len(info.Connections)
		interfaceList = append(interfaceList, info)
	}

	sort.Slice(interfaceList, func(i, j int) bool {
		return interfaceList[i].Name < interfaceList[j].Name
	})

	result.Interfaces = interfaceList
	result.TotalInterfaces = len(interfaceList)
	result.DurationMs = time.Since(start).Milliseconds()

	if connErr != nil {
		result.Error = connErr.Error()
	}

	return result, nil
}

// getInterfaceAddresses extracts all IP addresses from a network interface.
// Handles both *net.IPNet and *net.IPAddr types for cross-platform compatibility.
func getInterfaceAddresses(iface *net.Interface) []IPAddressInfo {
	addrs, err := iface.Addrs()
	if err != nil {
		return nil
	}

	var result []IPAddressInfo
	for _, addr := range addrs {
		var ip net.IP
		var mask net.IPMask

		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
			mask = v.Mask
		case *net.IPAddr:
			ip = v.IP
			mask = nil
		default:
			continue
		}

		if ip == nil {
			continue
		}

		info := IPAddressInfo{
			Address: ip.String(),
			Network: maskToCIDR(mask),
		}

		if ip.To4() != nil {
			info.AddressType = "IPv4"
		} else {
			info.AddressType = "IPv6"
		}

		result = append(result, info)
	}
	return result
}

// maskToCIDR converts an IPMask to a CIDR prefix length string, e.g. "24" for /24.
func maskToCIDR(mask net.IPMask) string {
	if mask == nil {
		return ""
	}
	ones, bits := mask.Size()
	if ones == 0 && bits == 0 {
		return ""
	}
	return fmt.Sprintf("%d", ones)
}

// macToString safely converts a MAC address to string, handling nil.
func macToString(mac net.HardwareAddr) string {
	if mac == nil {
		return ""
	}
	return mac.String()
}

// matchingConnections finds connections whose local IP matches any IP of the interface.
// Also matches connections bound to 0.0.0.0 or :: (wildcard).
func matchingConnections(ifaceIPs []IPAddressInfo, allConns []psnet.ConnectionStat) []ConnectionInfo {
	// Build a set of IPs for this interface for fast lookup
	ipSet := make(map[string]bool)
	for _, ip := range ifaceIPs {
		ipSet[ip.Address] = true
	}

	var result []ConnectionInfo
	seen := make(map[string]bool) // dedup key

	for _, c := range allConns {
		localIP := c.Laddr.IP

		// Only include connection if it matches this interface's IP, or is wildcard
		if !ipSet[localIP] && localIP != "0.0.0.0" && localIP != "::" {
			continue
		}

		ci := convertConnection(c)
		ci.ProcessName = getProcessName(c.Pid)

		// Dedup key: localIP:localPort:remoteIP:remotePort:protocol
		key := fmt.Sprintf("%s:%d:%s:%d:%s", ci.LocalIP, ci.LocalPort, ci.RemoteIP, ci.RemotePort, ci.Protocol)
		if seen[key] {
			continue
		}
		seen[key] = true

		result = append(result, ci)
	}

	return result
}

// dedupConnections removes duplicate connections from the list.
func dedupConnections(conns []ConnectionInfo) []ConnectionInfo {
	seen := make(map[string]bool)
	var result []ConnectionInfo
	for _, c := range conns {
		key := fmt.Sprintf("%s:%d:%s:%d:%s", c.LocalIP, c.LocalPort, c.RemoteIP, c.RemotePort, c.Protocol)
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, c)
	}
	return result
}

// convertConnection maps gopsutil's ConnectionStat to our ConnectionInfo.
func convertConnection(c psnet.ConnectionStat) ConnectionInfo {
	ci := ConnectionInfo{
		FD:     c.Fd,
		Family: c.Family,
		Type:   c.Type,
		PID:    c.Pid,
		Status: c.Status,
	}

	ci.LocalIP = c.Laddr.IP
	ci.LocalPort = c.Laddr.Port
	ci.RemoteIP = c.Raddr.IP
	ci.RemotePort = c.Raddr.Port

	switch {
	case c.Type == 1:
		ci.Protocol = "TCP"
	case c.Type == 2:
		ci.Protocol = "UDP"
	default:
		ci.Protocol = fmt.Sprintf("UNKNOWN(%d)", c.Type)
	}

	return ci
}

// getProcessName attempts to resolve a PID to a process name.
func getProcessName(pid int32) string {
	if pid <= 0 {
		return ""
	}
	p, err := process.NewProcess(pid)
	if err != nil {
		return ""
	}
	name, err := p.Name()
	if err != nil {
		return ""
	}
	return name
}

// isPermissionError checks if an error is related to insufficient privileges.
func isPermissionError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	patterns := []string{
		"permission denied",
		"access denied",
		"operation not permitted",
	}
	for _, p := range patterns {
		if strings.Contains(errStr, p) {
			return true
		}
	}
	return false
}
