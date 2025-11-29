package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/starbased-co/shine/pkg/config"
	"github.com/starbased-co/shine/pkg/rpc"
)

// Handlers handles RPC method calls
type Handlers struct {
	pm       *PanelManager
	state    *StateManager
	cfgPath  string
}

// handlePanelList returns a list of all active panels
func (h *Handlers) handlePanelList(ctx context.Context) (*rpc.PanelListResult, error) {
	panels := h.pm.ListPanels()

	result := &rpc.PanelListResult{
		Panels: make([]rpc.PanelInfo, len(panels)),
	}

	for i, panel := range panels {
		healthy := h.pm.CheckHealth(panel)
		result.Panels[i] = rpc.PanelInfo{
			Instance: panel.Instance,
			Name:     panel.Name,
			PID:      panel.PID,
			Socket:   panel.SocketPath,
			Healthy:  healthy,
		}
	}

	return result, nil
}

// handlePanelSpawn spawns a new panel
func (h *Handlers) handlePanelSpawn(ctx context.Context, req *rpc.PanelSpawnRequest) (*rpc.PanelSpawnResult, error) {
	// Parse config map into PrismEntry via JSON round-trip
	configJSON, err := json.Marshal(req.Config)
	if err != nil {
		return nil, rpc.ErrInvalidParams(fmt.Sprintf("failed to serialize config: %v", err))
	}

	var prismConfig config.PrismConfig
	if err := json.Unmarshal(configJSON, &prismConfig); err != nil {
		return nil, rpc.ErrInvalidParams(fmt.Sprintf("failed to parse prism config: %v", err))
	}

	// Validate required fields
	if prismConfig.Name == "" {
		return nil, rpc.ErrInvalidParams("prism name is required")
	}

	// Parse restart policy fields if present
	entry := &PrismEntry{
		PrismConfig: &prismConfig,
	}

	if restart, ok := req.Config["restart"].(string); ok {
		entry.Restart = restart
	}
	if restartDelay, ok := req.Config["restart_delay"].(string); ok {
		entry.RestartDelay = restartDelay
	}
	if maxRestarts, ok := req.Config["max_restarts"].(float64); ok {
		entry.MaxRestarts = int(maxRestarts)
	}

	// Validate restart policy if specified
	if err := entry.ValidateRestartPolicy(); err != nil {
		return nil, rpc.ErrConfig(err.Error())
	}

	// Generate instance name (use prism name if not specified)
	instanceName := prismConfig.Name
	if instance, ok := req.Config["instance"].(string); ok && instance != "" {
		instanceName = instance
	}

	// Check if panel already exists
	if _, exists := h.pm.GetPanel(instanceName); exists {
		return nil, rpc.ErrResourceBusy(fmt.Sprintf("panel instance %s already exists", instanceName))
	}

	log.Printf("panel/spawn: spawning panel %s (prism: %s)", instanceName, prismConfig.Name)

	// Spawn the panel
	panel, err := h.pm.SpawnPanel(entry, instanceName)
	if err != nil {
		return nil, rpc.ErrOperationFailed("spawn panel", err)
	}

	// Update state
	h.state.OnPanelSpawned(panel.Instance, panel.Name, panel.PID, true)

	log.Printf("panel/spawn: successfully spawned panel %s at %s", instanceName, panel.SocketPath)

	return &rpc.PanelSpawnResult{
		Instance: panel.Instance,
		Socket:   panel.SocketPath,
	}, nil
}

// handlePanelKill kills a panel
func (h *Handlers) handlePanelKill(ctx context.Context, req *rpc.PanelKillRequest) (*rpc.PanelKillResult, error) {
	if req.Instance == "" {
		return nil, rpc.ErrInvalidParams("instance name required")
	}

	err := h.pm.KillPanel(req.Instance)
	if err != nil {
		return &rpc.PanelKillResult{Killed: false}, err
	}

	// Update state
	h.state.OnPanelKilled(req.Instance)

	return &rpc.PanelKillResult{Killed: true}, nil
}

// handleServiceStatus returns aggregated service status
func (h *Handlers) handleServiceStatus(ctx context.Context) (*rpc.ServiceStatusResult, error) {
	panels := h.pm.ListPanels()

	result := &rpc.ServiceStatusResult{
		Panels:  make([]rpc.PanelInfo, len(panels)),
		Uptime:  h.state.Uptime().Milliseconds(),
		Version: version,
	}

	for i, panel := range panels {
		healthy := h.pm.CheckHealth(panel)
		result.Panels[i] = rpc.PanelInfo{
			Instance: panel.Instance,
			Name:     panel.Name,
			PID:      panel.PID,
			Socket:   panel.SocketPath,
			Healthy:  healthy,
		}
	}

	return result, nil
}

// handleConfigReload reloads the configuration
func (h *Handlers) handleConfigReload(ctx context.Context) (*rpc.ConfigReloadResult, error) {
	log.Println("config/reload via RPC")

	// Reload configuration using existing reloadConfig function
	err := reloadConfig(h.pm, h.cfgPath)
	if err != nil {
		return &rpc.ConfigReloadResult{
			Reloaded: false,
			Errors:   []string{err.Error()},
		}, nil
	}

	return &rpc.ConfigReloadResult{Reloaded: true}, nil
}
