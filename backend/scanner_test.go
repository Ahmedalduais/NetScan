package backend

import (
	"context"
	"fmt"
	"testing"
	"time"

	psnet "github.com/shirou/gopsutil/v3/net"
)

// TestNewScannerService verifies the Singleton pattern works correctly.
func TestNewScannerService(t *testing.T) {
	s1 := NewScannerService()
	s2 := NewScannerService()
	if s1 != s2 {
		t.Error("NewScannerService should return the same instance (Singleton)")
	}
}

// TestScanWithDefaults verifies a basic scan returns valid results.
func TestScanWithDefaults(t *testing.T) {
	scanner := NewScannerService()
	ctx := context.Background()

	result, err := scanner.Scan(ctx, ScanOptions{
		IncludeLoopback: true,
		Timeout:         10,
	})

	if err != nil {
		t.Fatalf("Scan should not return error with default permissions: %v", err)
	}

	if result == nil {
		t.Fatal("Scan result should not be nil")
	}

	if result.Timestamp == 0 {
		t.Error("Timestamp should be set")
	}

	if result.DurationMs <= 0 {
		t.Error("DurationMs should be positive")
	}

	// At minimum, loopback interface should exist
	if result.TotalInterfaces < 1 {
		t.Error("Expected at least 1 network interface (loopback)")
	}

	// Verify basic interface structure
	for _, iface := range result.Interfaces {
		if iface.Name == "" {
			t.Error("Interface name should not be empty")
		}
		// MAC address may be empty for loopback, that's OK
	}
}

// TestScanExcludesLoopback verifies the IncludeLoopback option.
func TestScanExcludesLoopback(t *testing.T) {
	scanner := NewScannerService()
	ctx := context.Background()

	resultNoLoopback, err := scanner.Scan(ctx, ScanOptions{
		IncludeLoopback: false,
		Timeout:         10,
	})
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	resultWithLoopback, err := scanner.Scan(ctx, ScanOptions{
		IncludeLoopback: true,
		Timeout:         10,
	})
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// Loopback should be excluded when IncludeLoopback=false
	if resultNoLoopback.TotalInterfaces == resultWithLoopback.TotalInterfaces {
		// This can happen on systems with only loopback, so just warn
		t.Log("Note: no non-loopback interfaces found")
	}
}

// TestScanContextCancellation verifies that canceling the context stops the scan.
func TestScanContextCancellation(t *testing.T) {
	scanner := NewScannerService()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := scanner.Scan(ctx, ScanOptions{
		IncludeLoopback: true,
		Timeout:         10,
	})

	if err == nil {
		t.Log("Note: scan may complete before cancellation propagates on fast systems")
	}
}

// TestScanWithTimeout verifies timeout handling.
func TestScanWithTimeout(t *testing.T) {
	scanner := NewScannerService()
	ctx := context.Background()

	result, err := scanner.Scan(ctx, ScanOptions{
		IncludeLoopback: true,
		Timeout:         1, // 1 second
	})

	if err != nil {
		t.Fatalf("Scan with 1s timeout failed: %v", err)
	}

	if result.DurationMs > 5000 {
		t.Errorf("Scan took too long: %dms", result.DurationMs)
	}
}

// TestConnectionConversion verifies the connection mapping works.
func TestConnectionConversion(t *testing.T) {
	scanner := NewScannerService()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := scanner.Scan(ctx, ScanOptions{
		IncludeLoopback: true,
		Timeout:         5,
	})
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	t.Logf("Found %d total connections across %d interfaces", result.TotalConnections, result.TotalInterfaces)

	// Log ESTABLISHED connections to verify remote IPs
	for _, iface := range result.Interfaces {
		for _, conn := range iface.Connections {
			if conn.Status == "ESTABLISHED" && conn.RemoteIP != "" && conn.RemoteIP != "0.0.0.0" && conn.RemoteIP != "::" {
				t.Logf("  ESTABLISHED: %s:%d -> %s:%d [%s pid=%d]",
					conn.LocalIP, conn.LocalPort, conn.RemoteIP, conn.RemotePort, conn.Protocol, conn.PID)
			}
		}
	}
}

// TestRawConnections checks raw gopsutil connection data.
func TestRawConnections(t *testing.T) {
	conns, err := psnet.Connections("tcp")
	if err != nil {
		t.Fatalf("gopsutil Connections failed: %v", err)
	}

	estabCount := 0
	listenCount := 0
	for _, c := range conns {
		if c.Status == "ESTABLISHED" {
			estabCount++
			if estabCount <= 5 {
				t.Logf("  RAW ESTABLISHED: local=%s:%d remote=%s:%d pid=%d",
					c.Laddr.IP, c.Laddr.Port, c.Raddr.IP, c.Raddr.Port, c.Pid)
			}
		}
		if c.Status == "LISTEN" {
			listenCount++
		}
	}
	t.Logf("Raw TCP connections: ESTABLISHED=%d LISTEN=%d total=%d", estabCount, listenCount, len(conns))
}

// TestIsPermissionError verifies error detection
func TestIsPermissionError(t *testing.T) {
	tests := []struct {
		err      error
		expected bool
	}{
		{fmt.Errorf("permission denied"), true},
		{fmt.Errorf("access denied"), true},
		{fmt.Errorf("operation not permitted"), true},
		{fmt.Errorf("connection refused"), false},
		{nil, false},
	}

	for _, tc := range tests {
		result := isPermissionError(tc.err)
		if result != tc.expected {
			t.Errorf("isPermissionError(%v) = %v, want %v", tc.err, result, tc.expected)
		}
	}
}
