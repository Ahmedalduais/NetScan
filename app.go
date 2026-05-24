package main

import (
	"context"

	"MyApp/backend"
)

// App struct wraps the network controller.
// Uses the Service-Controller pattern to separate concerns.
// Reference: Wails v2 Binding docs - https://wails.io/docs/howdoesitwork
type App struct {
	ctx      context.Context
	ctrl     *backend.NetworkController
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		ctrl: backend.NewNetworkController(),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.ctrl.Start(ctx)
}

// shutdown is called when the app is shutting down
func (a *App) shutdown(ctx context.Context) {
	a.ctrl.Stop()
}

// ScanNetwork delegates to the controller.
// Exposed to frontend via Wails Bind.
func (a *App) ScanNetwork(opts backend.ScanOptions) *backend.ScanResult {
	return a.ctrl.ScanNetwork(opts)
}

// QuickScan runs a scan with default options.
func (a *App) QuickScan() *backend.ScanResult {
	return a.ctrl.QuickScan()
}

// GetPlatform returns the OS name.
func (a *App) GetPlatform() string {
	return a.ctrl.GetPlatform()
}

// BlockTarget delegates to the controller's firewall.
func (a *App) BlockTarget(req backend.BlockRequest) *backend.BlockResult {
	return a.ctrl.BlockTarget(req)
}

// UnblockTarget delegates to the controller's firewall.
func (a *App) UnblockTarget(blockID string) *backend.BlockResult {
	return a.ctrl.UnblockTarget(blockID)
}

// GetBlocked delegates to the controller's firewall.
func (a *App) GetBlocked() []backend.BlockedEntry {
	return a.ctrl.GetBlocked()
}

// IsAdmin delegates to the controller's firewall.
func (a *App) IsAdmin() bool {
	return a.ctrl.IsAdmin()
}

// GetThroughput delegates to the controller's throughput monitor.
func (a *App) GetThroughput() []backend.ThroughputData {
	return a.ctrl.GetThroughput()
}

// GetInterfaceThroughput delegates to the controller's throughput monitor.
func (a *App) GetInterfaceThroughput(name string) (backend.ThroughputData, bool) {
	return a.ctrl.GetInterfaceThroughput(name)
}
