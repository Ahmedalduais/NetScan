package backend

// InterfaceInfo represents a network interface with its addresses and connections.
// Uses net.Interface and net.Addr from the standard library (RFC 1122, Section 2).
type InterfaceInfo struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	MACAddress  string           `json:"mac_address"`
	IsUp        bool             `json:"is_up"`
	IsLoopback  bool             `json:"is_loopback"`
	MTU         int              `json:"mtu"`
	Flags       []string         `json:"flags"`
	IPAddresses []IPAddressInfo  `json:"ip_addresses"`
	Connections []ConnectionInfo `json:"connections"`
}

// IPAddressInfo holds a single IP address and its subnet mask/prefix length.
type IPAddressInfo struct {
	Address     string `json:"address"`
	Network     string `json:"network"`
	AddressType string `json:"address_type"`
}

// ConnectionInfo represents an active or listening socket connection.
// Modeled after net.ConnectionStat from gopsutil (Summers, M. (2021). Go in Action).
type ConnectionInfo struct {
	FD          uint32 `json:"fd"`
	Family      uint32 `json:"family"`
	Type        uint32 `json:"type"`
	LocalIP     string `json:"local_ip"`
	LocalPort   uint32 `json:"local_port"`
	RemoteIP    string `json:"remote_ip"`
	RemotePort  uint32 `json:"remote_port"`
	Status      string `json:"status"`
	PID         int32  `json:"pid"`
	ProcessName string `json:"process_name"`
	Protocol    string `json:"protocol"`
}

// ScanResult is the top-level result returned to the frontend.
type ScanResult struct {
	Interfaces       []InterfaceInfo `json:"interfaces"`
	TotalInterfaces  int             `json:"total_interfaces"`
	TotalConnections int             `json:"total_connections"`
	Timestamp        int64           `json:"timestamp"`
	DurationMs       int64           `json:"duration_ms"`
	Error            string          `json:"error,omitempty"`
	PermissionError  bool            `json:"permission_error"`
}

// ScanOptions allows the frontend to control scan behavior.
type ScanOptions struct {
	IncludeLoopback bool `json:"include_loopback"`
	Timeout         int  `json:"timeout"`
}

// BlockRequest is sent from the frontend to block an IP or PID.
type BlockRequest struct {
	Type      string `json:"type"` // "ip" or "pid"
	Target    string `json:"target"` // IP address or PID
	Interface string `json:"interface,omitempty"`
	PID       int32  `json:"pid,omitempty"`
}

// BlockResult is returned from block/unblock operations.
type BlockResult struct {
	Success  bool        `json:"success"`
	Message  string      `json:"message"`
	Action   string      `json:"action"` // "block" or "unblock"
	Entry    BlockedEntry `json:"entry,omitempty"`
	IsAdmin  bool        `json:"is_admin"`
}

// IsAdminResult reports admin status.
type IsAdminResult struct {
	IsAdmin bool `json:"is_admin"`
}
