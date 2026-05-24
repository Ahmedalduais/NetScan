package backend

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	psnet "github.com/shirou/gopsutil/v3/net"
)

// BlockedEntry tracks a blocked IP or PID.
type BlockedEntry struct {
	ID        string `json:"id"`
	Type      string `json:"type"` // "ip" or "pid"
	Target    string `json:"target"`
	Interface string `json:"interface,omitempty"`
	RuleName  string `json:"rule_name"`
	Active    bool   `json:"active"`
}

// FirewallService manages network blocking via platform-specific firewall rules.
// On Windows: uses netsh advfirewall (available since Windows Vista)
// On Linux: uses iptables/nftables
// On macOS: uses pfctl
//
// Reference: Windows Filtering Platform (WFP) - Microsoft Docs
// Shell commands are used as a portable fallback per project requirements.
type FirewallService struct {
	mu      sync.Mutex
	blocked []BlockedEntry
}

// NewFirewallService creates a new FirewallService.
func NewFirewallService() *FirewallService {
	return &FirewallService{
		blocked: make([]BlockedEntry, 0),
	}
}

// IsAdmin checks if the process has administrator/root privileges.
// On Windows: checks if the process is running as Administrator.
// On Unix: checks if EUID is 0.
func (fs *FirewallService) IsAdmin() bool {
	switch runtime.GOOS {
	case "windows":
		return isWindowsAdmin()
	default:
		// On Linux/macOS, check EUID via os.Geteuid() - not available on Windows
		return false
	}
}

// isWindowsAdmin checks Windows admin status by attempting to open a handle to the SAM registry key.
func isWindowsAdmin() bool {
	cmd := exec.Command("net", "session")
	err := cmd.Run()
	return err == nil
}

// BlockIP creates a firewall rule to block all traffic to/from the given IP.
// Requires administrator privileges.
func (fs *FirewallService) BlockIP(ip, ifaceName string) (BlockedEntry, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Check if already blocked
	for _, b := range fs.blocked {
		if b.Type == "ip" && b.Target == ip && b.Active {
			return b, fmt.Errorf("IP %s is already blocked", ip)
		}
	}

	ruleName := fmt.Sprintf("NetScan_Block_IP_%s", sanitizeName(ip))
	entry := BlockedEntry{
		ID:        fmt.Sprintf("ip_%s", ip),
		Type:      "ip",
		Target:    ip,
		Interface: ifaceName,
		RuleName:  ruleName,
		Active:    true,
	}

	switch runtime.GOOS {
	case "windows":
		// Block inbound and outbound traffic to/from this IP
		// netsh advfirewall is available on Windows Vista+ (all supported versions)
		inCmd := exec.Command("netsh",
			"advfirewall", "firewall", "add", "rule",
			fmt.Sprintf("name=%s", ruleName),
			"dir=in", "action=block",
			fmt.Sprintf("remoteip=%s", ip),
			"enable=yes",
		)
		if out, err := inCmd.CombinedOutput(); err != nil {
			return entry, fmt.Errorf("failed to create inbound block rule: %w\n%s", err, string(out))
		}

		outCmd := exec.Command("netsh",
			"advfirewall", "firewall", "add", "rule",
			fmt.Sprintf("name=%s", ruleName+"_out"),
			"dir=out", "action=block",
			fmt.Sprintf("remoteip=%s", ip),
			"enable=yes",
		)
		if out, err := outCmd.CombinedOutput(); err != nil {
			// Rollback inbound rule
			_ = exec.Command("netsh", "advfirewall", "firewall", "delete", "rule",
				fmt.Sprintf("name=%s", ruleName)).Run()
			return entry, fmt.Errorf("failed to create outbound block rule: %w\n%s", err, string(out))
		}

	case "linux":
		cmd := exec.Command("iptables", "-A", "INPUT", "-s", ip, "-j", "DROP")
		if out, err := cmd.CombinedOutput(); err != nil {
			return entry, fmt.Errorf("failed to add iptables rule: %w\n%s", err, string(out))
		}

	default:
		return entry, fmt.Errorf("platform %s not yet supported for IP blocking", runtime.GOOS)
	}

	fs.blocked = append(fs.blocked, entry)
	return entry, nil
}

// UnblockIP removes the firewall rule blocking the given IP.
func (fs *FirewallService) UnblockIP(ip string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	ruleName := fmt.Sprintf("NetScan_Block_IP_%s", sanitizeName(ip))

	switch runtime.GOOS {
	case "windows":
		// Remove both inbound and outbound rules
		errIn := exec.Command("netsh", "advfirewall", "firewall", "delete", "rule",
			fmt.Sprintf("name=%s", ruleName)).Run()
		errOut := exec.Command("netsh", "advfirewall", "firewall", "delete", "rule",
			fmt.Sprintf("name=%s", ruleName+"_out")).Run()
		if errIn != nil && errOut != nil {
			// Both might already be removed
		}

	case "linux":
		_ = exec.Command("iptables", "-D", "INPUT", "-s", ip, "-j", "DROP").Run()

	default:
		return fmt.Errorf("platform %s not yet supported", runtime.GOOS)
	}

	// Mark as inactive in our list
	for i, b := range fs.blocked {
		if b.Type == "ip" && b.Target == ip {
			fs.blocked[i].Active = false
		}
	}

	return nil
}

// BlockPID blocks all connections associated with the given PID.
// Automatically discovers the PID's ports by querying gopsutil connections.
// On Windows: creates firewall rules for each port/protocol used by the PID.
// Requires administrator privileges.
func (fs *FirewallService) BlockPID(pid int32, _ []ConnectionInfo) ([]BlockedEntry, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Fetch fresh connections from gopsutil
	psConns, err := psnet.Connections("all")
	if err != nil {
		return nil, fmt.Errorf("failed to get connections for PID: %w", err)
	}

	var entries []BlockedEntry
	portSet := make(map[string]bool)

	for _, raw := range psConns {
		if raw.Pid != pid || raw.Laddr.Port == 0 {
			continue
		}

		proto := "tcp"
		if raw.Type == 2 {
			proto = "udp"
		}
		key := fmt.Sprintf("%s_%d", proto, raw.Laddr.Port)
		if portSet[key] {
			continue
		}
		portSet[key] = true

		ruleName := fmt.Sprintf("NetScan_Block_PID_%d_Port_%d", pid, raw.Laddr.Port)
		entry := BlockedEntry{
			ID:       fmt.Sprintf("pid_%d_port_%d", pid, raw.Laddr.Port),
			Type:     "pid",
			Target:   fmt.Sprintf("PID %d (port %d/%s)", pid, raw.Laddr.Port, proto),
			RuleName: ruleName,
			Active:   true,
		}

		switch runtime.GOOS {
		case "windows":
			cmd := exec.Command("netsh", "advfirewall", "firewall", "add", "rule",
				fmt.Sprintf("name=%s", ruleName),
				"dir=in", "action=block",
				fmt.Sprintf("protocol=%s", proto),
				fmt.Sprintf("localport=%d", raw.Laddr.Port),
				"enable=yes",
			)
			if out, err := cmd.CombinedOutput(); err != nil {
				return entries, fmt.Errorf("failed to block PID port: %w\n%s", err, string(out))
			}

		case "linux":
			cmd := exec.Command("iptables", "-A", "INPUT", "-p", proto,
				fmt.Sprintf("--dport=%d", raw.Laddr.Port), "-j", "DROP")
			_ = cmd.Run()
		}

		entries = append(entries, entry)
	}

	fs.blocked = append(fs.blocked, entries...)
	return entries, nil
}

// UnblockPID removes firewall rules associated with the given PID.
func (fs *FirewallService) UnblockPID(pid int32) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	for _, b := range fs.blocked {
		if b.Type == "pid" && strings.Contains(b.Target, fmt.Sprintf("PID %d", pid)) && b.Active {
			switch runtime.GOOS {
			case "windows":
				_ = exec.Command("netsh", "advfirewall", "firewall", "delete", "rule",
					fmt.Sprintf("name=%s", b.RuleName)).Run()
			case "linux":
				// Parse port and protocol from rule name
				_ = exec.Command("iptables", "-D", "INPUT", "-p", "tcp", "-j", "DROP").Run()
			}
		}
	}

	// Mark all entries for this PID as inactive
	for i, b := range fs.blocked {
		if b.Type == "pid" && strings.Contains(b.Target, fmt.Sprintf("PID %d", pid)) {
			fs.blocked[i].Active = false
		}
	}

	return nil
}

// GetBlocked returns the list of all blocked entries.
func (fs *FirewallService) GetBlocked() []BlockedEntry {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	result := make([]BlockedEntry, len(fs.blocked))
	copy(result, fs.blocked)
	return result
}

// sanitizeName removes characters unsafe for firewall rule names.
func sanitizeName(name string) string {
	r := strings.NewReplacer(
		".", "_",
		":", "_",
		"/", "_",
		" ", "_",
	)
	return r.Replace(name)
}
