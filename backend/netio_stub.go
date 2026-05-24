//go:build !windows

package backend

// NetIOMonitor is a no-op stub for non-Windows platforms.
// Per-process network IO data uses proportional estimation instead.
type NetIOMonitor struct{}

func NewNetIOMonitor() *NetIOMonitor {
	return &NetIOMonitor{}
}

func (nm *NetIOMonitor) Start() error {
	return nil
}

func (nm *NetIOMonitor) Stop() {}

func (nm *NetIOMonitor) GetCounters() []ProcessNetIO {
	return nil
}

func (nm *NetIOMonitor) GetBytesForPID(pid int32) (recv, sent uint64, ok bool) {
	return 0, 0, false
}
