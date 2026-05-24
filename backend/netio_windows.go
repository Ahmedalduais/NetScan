//go:build windows

package backend

import (
	"encoding/binary"
	"fmt"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

var (
	modAdvapi32 = syscall.NewLazyDLL("advapi32.dll")
	modTdh      = syscall.NewLazyDLL("tdh.dll")

	procStartTraceW    = modAdvapi32.NewProc("StartTraceW")
	procEnableTraceEx2 = modAdvapi32.NewProc("EnableTraceEx2")
	procOpenTraceW     = modAdvapi32.NewProc("OpenTraceW")
	procProcessTrace   = modAdvapi32.NewProc("ProcessTrace")
	procCloseTrace     = modAdvapi32.NewProc("CloseTrace")
	procControlTraceW  = modAdvapi32.NewProc("ControlTraceW")
)

type etwWnodeHeader struct {
	BufferSize        uint32
	ProviderIDSize    uint32 `struct:"[unspecified]"`
	HistoricalContext uintptr
	TimeDelta         int64
}

type etwProperties struct {
	Wnode               etwWnodeHeader
	BufferSize          uint32
	MinimumBuffers      uint32
	MaximumBuffers      uint32
	MaximumFileSize     uint32
	LogFileMode         uint32
	FlushTimer          uint32
	EnableFlags         uint32
	AgeLimit            int32
	NumberOfBuffers     uint32
	FreeBuffers         uint32
	EventsLost          uint32
	BuffersWritten      uint32
	LogBuffersLost      uint32
	RealTimeBuffersLost uint32
	LoggerThreadID      uintptr
	LogFileNameOffset   uint32
	LoggerNameOffset    uint32
}

type etwLogfileW struct {
	LogFileName       *uint16
	LoggerName        *uint16
	CurrentTime       int64
	BuffersRead       uint32
	_                 uint32
	LogFileMode       uint32
	_                 uint32
	_                 uint32
	_                 uint64
	_                 [4]uint64
	EventCallback     uintptr
	BufferCallback    uintptr
	Context           uintptr
	_                 uint64
	_                 uint64
	_                 uint32
	EventsLost        uint32
	BuffersLost       uint32
	_                 uint32
	_                 uint64
	_                 uint64
}

type etwEventHeader struct {
	Size          uint16
	HeaderType    uint16
	Flags         uint16
	EventProperty uint16
	ThreadID      uint32
	ProcessID     uint32
	TimeStamp     int64
	_             [16]byte
	_             [16]byte
	_             uint32
	_             uint32
	_             uint32
	_             uint32
	ExtendedData  uintptr
	UserDataLen   uint32
	UserData      uintptr
	_             uint32
	_             uint32
}

const (
	etwRealTime        = 0x00000100
	etwModeEventRecord = 0x10000000
	etwControlStop     = 0
	etwLevelVerbose    = 0xFF
)

// NetIOMonitor tracks per-process network IO using Windows ETW.
type NetIOMonitor struct {
	mu        sync.RWMutex
	counters  map[int32]*ProcessNetIO
	session   syscall.Handle
	trace     syscall.Handle
	running   bool
	startedAt time.Time
	cb        uintptr
}

func NewNetIOMonitor() *NetIOMonitor {
	return &NetIOMonitor{
		counters: make(map[int32]*ProcessNetIO),
	}
}

// Start begins the ETW trace to collect real per-PID network byte counts.
func (nm *NetIOMonitor) Start() error {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	if nm.running {
		return nil
	}

	sessionName := fmt.Sprintf("NetScanPerf_%d", time.Now().UnixNano())
	sn := syscall.StringToUTF16(sessionName)
	nameLen := len(sn) * 2

	totalSize := int(unsafe.Sizeof(etwProperties{})) + nameLen + 2
	buf := make([]byte, totalSize)

	props := (*etwProperties)(unsafe.Pointer(&buf[0]))
	props.Wnode.BufferSize = uint32(totalSize)
	props.Wnode.ProviderIDSize = 0
	props.BufferSize = 256
	props.MinimumBuffers = 8
	props.MaximumBuffers = 64
	props.LogFileMode = etwRealTime
	props.FlushTimer = 1
	props.LoggerNameOffset = uint32(unsafe.Sizeof(etwProperties{}))

	nameDest := buf[props.LoggerNameOffset:]
	copy(nameDest, (*[1 << 20]byte)(unsafe.Pointer(&sn[0]))[:nameLen])

	var hSession syscall.Handle
	r1, _, _ := procStartTraceW.Call(
		uintptr(unsafe.Pointer(&hSession)),
		uintptr(unsafe.Pointer(&sn[0])),
		uintptr(unsafe.Pointer(props)),
	)
	if r1 != 0 && r1 != 183 { // 183 = ERROR_ALREADY_EXISTS
		// Try with a simpler session name
		sessionName = "NetScanTrace"
		sn = syscall.StringToUTF16(sessionName)
		nameLen = len(sn) * 2
		totalSize = int(unsafe.Sizeof(etwProperties{})) + nameLen + 2
		buf = make([]byte, totalSize)
		props = (*etwProperties)(unsafe.Pointer(&buf[0]))
		props.Wnode.BufferSize = uint32(totalSize)
		props.BufferSize = 256
		props.MinimumBuffers = 8
		props.MaximumBuffers = 64
		props.LogFileMode = etwRealTime
		props.FlushTimer = 1
		props.LoggerNameOffset = uint32(unsafe.Sizeof(etwProperties{}))

		nameDest = buf[props.LoggerNameOffset:]
		copy(nameDest, (*[1 << 20]byte)(unsafe.Pointer(&sn[0]))[:nameLen])

		r1, _, _ = procStartTraceW.Call(
			uintptr(unsafe.Pointer(&hSession)),
			uintptr(unsafe.Pointer(&sn[0])),
			uintptr(unsafe.Pointer(props)),
		)
		if r1 != 0 {
			return fmt.Errorf("StartTraceW failed: %d", r1)
		}
	}
	nm.session = hSession

	tcpipGUID := [16]byte{
		0xEE, 0xE2, 0x07, 0x2F, 0xDB, 0x15, 0xF1, 0x40,
		0x90, 0xEF, 0x9D, 0x7B, 0xA2, 0x82, 0x18, 0x8A,
	}

	r1, _, _ = procEnableTraceEx2.Call(
		uintptr(hSession),
		uintptr(unsafe.Pointer(&tcpipGUID[0])),
		1, // ControlCode = Enable
		etwLevelVerbose,
		0xFFFFFFFFFFFFFFFF, // AnyKeyword
		0, 0, 0,
	)
	if r1 != 0 {
		nm.cleanup()
		return fmt.Errorf("EnableTraceEx2 failed: %d", r1)
	}

	cb := syscall.NewCallback(func(eventRecord unsafe.Pointer) uintptr {
		nm.handleEvent(eventRecord)
		return 0
	})
	nm.cb = cb

	snPtr, _ := syscall.UTF16PtrFromString(sessionName)
	lf := &etwLogfileW{
		LogFileName:   nil,
		LoggerName:    snPtr,
		LogFileMode:   etwModeEventRecord,
		EventCallback: cb,
		Context:       0,
	}

	lfbuf := make([]byte, unsafe.Sizeof(*lf))
	lfptr := (*etwLogfileW)(unsafe.Pointer(&lfbuf[0]))
	*lfptr = *lf

	var hTrace syscall.Handle
	r1, _, _ = procOpenTraceW.Call(
		uintptr(unsafe.Pointer(lfptr)),
		uintptr(unsafe.Pointer(&hTrace)),
	)
	if r1 != 0 && r1 != 0x3E7 { // 0x3E7 = WNODE_FLAG_WMI_INCORRECT_LOGFILE
		nm.cleanup()
		return fmt.Errorf("OpenTraceW failed: %d", r1)
	}

	nm.trace = hTrace
	nm.running = true
	nm.startedAt = time.Now()

	go func() {
		procProcessTrace.Call(
			uintptr(hTrace),
			0,
			0,
			0,
		)
	}()

	return nil
}

func (nm *NetIOMonitor) handleEvent(eventRecord unsafe.Pointer) {
	hdr := (*etwEventHeader)(eventRecord)
	if hdr.Size < uint16(unsafe.Sizeof(etwEventHeader{})) || hdr.UserData == 0 || hdr.UserDataLen < 8 {
		return
	}

	data := (*[1 << 20]byte)(unsafe.Pointer(hdr.UserData))[:hdr.UserDataLen:hdr.UserDataLen]

	pid := int32(binary.LittleEndian.Uint32(data[0:4]))
	size := binary.LittleEndian.Uint32(data[4:8])
	if pid <= 0 || size == 0 {
		return
	}

	// Event ID is in HeaderType for TCP/IP ETW events
	// 10=TcpIp_SendIPV4, 11=TcpIp_RecvIPV4, 26=TcpIp_SendIPV6, 27=TcpIp_RecvIPV6
	isRecv := hdr.HeaderType == 11 || hdr.HeaderType == 27

	nm.mu.Lock()
	c, ok := nm.counters[pid]
	if !ok {
		c = &ProcessNetIO{PID: pid}
		nm.counters[pid] = c
	}
	if isRecv {
		c.BytesRecv += uint64(size)
	} else {
		c.BytesSent += uint64(size)
	}
	nm.mu.Unlock()
}

func (nm *NetIOMonitor) cleanup() {
	if nm.session != 0 {
		props := &etwProperties{}
		props.Wnode.BufferSize = uint32(unsafe.Sizeof(*props))
		procControlTraceW.Call(
			uintptr(nm.session),
			0,
			uintptr(unsafe.Pointer(props)),
			etwControlStop,
		)
		nm.session = 0
	}
}

func (nm *NetIOMonitor) Stop() {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	if !nm.running {
		return
	}
	nm.running = false

	if nm.trace != 0 {
		procCloseTrace.Call(uintptr(nm.trace))
		nm.trace = 0
	}
	nm.cleanup()
}

func (nm *NetIOMonitor) GetCounters() []ProcessNetIO {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	res := make([]ProcessNetIO, 0, len(nm.counters))
	for _, c := range nm.counters {
		res = append(res, *c)
	}
	return res
}

func (nm *NetIOMonitor) GetBytesForPID(pid int32) (recv, sent uint64, ok bool) {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	c, found := nm.counters[pid]
	if !found {
		return 0, 0, false
	}
	return c.BytesRecv, c.BytesSent, true
}
