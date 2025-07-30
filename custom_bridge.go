// Copyright (C) ConfigHub, Inc.
// SPDX-License-Identifier: MIT

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/confighub/sdk/bridge-worker/api"
	"github.com/confighub/sdk/bridge-worker/impl"
)

// CustomKubernetesBridge wraps the standard Kubernetes bridge and adds file persistence
type CustomKubernetesBridge struct {
	name             string
	baseDir          string
	kubernetesBridge *impl.KubernetesBridgeWorker
}

// NewCustomKubernetesBridge creates a new CustomKubernetesBridge instance
func NewCustomKubernetesBridge(name, baseDir string) (*CustomKubernetesBridge, error) {
	if baseDir == "" {
		baseDir = "/tmp/confighub-custom-bridge"
	}
	err := os.MkdirAll(baseDir, 0755)
	if err != nil {
		return nil, err
	}

	// Create the underlying Kubernetes bridge
	kubernetesBridge := &impl.KubernetesBridgeWorker{}

	return &CustomKubernetesBridge{
		name:             name,
		baseDir:          baseDir,
		kubernetesBridge: kubernetesBridge,
	}, nil
}

// Info returns information about the bridge's capabilities
func (cb *CustomKubernetesBridge) Info(opts api.InfoOptions) api.BridgeInfo {
	// Delegate to the underlying Kubernetes bridge
	return cb.kubernetesBridge.Info(opts)
}

// Apply handles the apply operation by delegating to the Kubernetes bridge and then saving files
func (cb *CustomKubernetesBridge) Apply(ctx api.BridgeContext, payload api.BridgePayload) error {
	startTime := time.Now()

	// Send initial status
	if err := ctx.SendStatus(&api.ActionResult{
		UnitID:            payload.UnitID,
		SpaceID:           payload.SpaceID,
		QueuedOperationID: payload.QueuedOperationID,
		ActionResultBaseMeta: api.ActionResultBaseMeta{
			Action:    api.ActionApply,
			Result:    api.ActionResultNone,
			Status:    api.ActionStatusProgressing,
			Message:   fmt.Sprintf("Starting apply operation for %s", cb.name),
			StartedAt: startTime,
		},
	}); err != nil {
		return err
	}

	// Delegate to the underlying Kubernetes bridge
	err := cb.kubernetesBridge.Apply(ctx, payload)
	if err != nil {
		// Send error status
		endedAt := time.Now()
		ctx.SendStatus(&api.ActionResult{
			UnitID:            payload.UnitID,
			SpaceID:           payload.SpaceID,
			QueuedOperationID: payload.QueuedOperationID,
			ActionResultBaseMeta: api.ActionResultBaseMeta{
				Action:       api.ActionApply,
				Result:       api.ActionResultApplyFailed,
				Status:       api.ActionStatusFailed,
				Message:      fmt.Sprintf("Apply failed: %v", err),
				StartedAt:    startTime,
				TerminatedAt: &endedAt,
			},
		})
		return err
	}

	// Save config unit data and metadata to files
	err = cb.saveConfigUnit(payload)
	if err != nil {
		// Send error status
		endedAt := time.Now()
		ctx.SendStatus(&api.ActionResult{
			UnitID:            payload.UnitID,
			SpaceID:           payload.SpaceID,
			QueuedOperationID: payload.QueuedOperationID,
			ActionResultBaseMeta: api.ActionResultBaseMeta{
				Action:       api.ActionApply,
				Result:       api.ActionResultApplyFailed,
				Status:       api.ActionStatusFailed,
				Message:      fmt.Sprintf("Failed to save config unit: %v", err),
				StartedAt:    startTime,
				TerminatedAt: &endedAt,
			},
		})
		return err
	}

	// Send success status
	endedAt := time.Now()
	return ctx.SendStatus(&api.ActionResult{
		UnitID:            payload.UnitID,
		SpaceID:           payload.SpaceID,
		QueuedOperationID: payload.QueuedOperationID,
		ActionResultBaseMeta: api.ActionResultBaseMeta{
			Action:       api.ActionApply,
			Result:       api.ActionResultApplyCompleted,
			Status:       api.ActionStatusCompleted,
			Message:      fmt.Sprintf("Apply completed successfully for %s", cb.name),
			StartedAt:    startTime,
			TerminatedAt: &endedAt,
		},
	})
}

// Refresh handles the refresh operation by delegating to the Kubernetes bridge
func (cb *CustomKubernetesBridge) Refresh(ctx api.BridgeContext, payload api.BridgePayload) error {
	return cb.kubernetesBridge.Refresh(ctx, payload)
}

// Import handles the import operation by delegating to the Kubernetes bridge
func (cb *CustomKubernetesBridge) Import(ctx api.BridgeContext, payload api.BridgePayload) error {
	return cb.kubernetesBridge.Import(ctx, payload)
}

// Destroy handles the destroy operation by delegating to the Kubernetes bridge and then deleting files
func (cb *CustomKubernetesBridge) Destroy(ctx api.BridgeContext, payload api.BridgePayload) error {
	startTime := time.Now()

	// Send initial status
	if err := ctx.SendStatus(&api.ActionResult{
		UnitID:            payload.UnitID,
		SpaceID:           payload.SpaceID,
		QueuedOperationID: payload.QueuedOperationID,
		ActionResultBaseMeta: api.ActionResultMeta{
			Action:    api.ActionDestroy,
			Result:    api.ActionResultNone,
			Status:    api.ActionStatusProgressing,
			Message:   fmt.Sprintf("Starting destroy operation for %s", cb.name),
			StartedAt: startTime,
		},
	}); err != nil {
		return err
	}

	// Delegate to the underlying Kubernetes bridge
	err := cb.kubernetesBridge.Destroy(ctx, payload)
	if err != nil {
		// Send error status
		endedAt := time.Now()
		ctx.SendStatus(&api.ActionResult{
			UnitID:            payload.UnitID,
			SpaceID:           payload.SpaceID,
			QueuedOperationID: payload.QueuedOperationID,
			ActionResultBaseMeta: api.ActionResultBaseMeta{
				Action:       api.ActionDestroy,
				Result:       api.ActionResultDestroyFailed,
				Status:       api.ActionStatusFailed,
				Message:      fmt.Sprintf("Destroy failed: %v", err),
				StartedAt:    startTime,
				TerminatedAt: &endedAt,
			},
		})
		return err
	}

	// Delete config unit files
	err = cb.deleteConfigUnit(payload)
	if err != nil {
		// Send error status
		endedAt := time.Now()
		ctx.SendStatus(&api.ActionResult{
			UnitID:            payload.UnitID,
			SpaceID:           payload.SpaceID,
			QueuedOperationID: payload.QueuedOperationID,
			ActionResultBaseMeta: api.ActionResultBaseMeta{
				Action:       api.ActionDestroy,
				Result:       api.ActionResultDestroyFailed,
				Status:       api.ActionStatusFailed,
				Message:      fmt.Sprintf("Failed to delete config unit files: %v", err),
				StartedAt:    startTime,
				TerminatedAt: &endedAt,
			},
		})
		return err
	}

	// Send success status
	endedAt := time.Now()
	return ctx.SendStatus(&api.ActionResult{
		UnitID:            payload.UnitID,
		SpaceID:           payload.SpaceID,
		QueuedOperationID: payload.QueuedOperationID,
		ActionResultBaseMeta: api.ActionResultBaseMeta{
			Action:       api.ActionDestroy,
			Result:       api.ActionResultDestroyCompleted,
			Status:       api.ActionStatusCompleted,
			Message:      fmt.Sprintf("Destroy completed successfully for %s", cb.name),
			StartedAt:    startTime,
			TerminatedAt: &endedAt,
		},
	})
}

// Finalize handles the finalize operation by delegating to the Kubernetes bridge
func (cb *CustomKubernetesBridge) Finalize(ctx api.BridgeContext, payload api.BridgePayload) error {
	return cb.kubernetesBridge.Finalize(ctx, payload)
}

// saveConfigUnit saves the config unit data and metadata to files
func (cb *CustomKubernetesBridge) saveConfigUnit(payload api.BridgePayload) error {
	// Create directory structure: space-id/unit-slug
	unitDir := filepath.Join(cb.baseDir, payload.SpaceID.String(), payload.UnitSlug)
	err := os.MkdirAll(unitDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create unit directory: %v", err)
	}

	// Save YAML file with decoded Data field
	if len(payload.Data) > 0 {
		// Data is already in bytes, no need to decode from base64
		yamlFile := filepath.Join(unitDir, "data.yaml")
		err = os.WriteFile(yamlFile, payload.Data, 0644)
		if err != nil {
			return fmt.Errorf("failed to write YAML file: %v", err)
		}
	}

	// Save JSON file with metadata (excluding Data field)
	metadata := map[string]interface{}{
		"UnitID":            payload.UnitID,
		"SpaceID":           payload.SpaceID,
		"UnitSlug":          payload.UnitSlug,
		"ToolchainType":     payload.ToolchainType,
		"ProviderType":      payload.ProviderType,
		"QueuedOperationID": payload.QueuedOperationID,
		"TargetParams":      payload.TargetParams,
		"ExtraParams":       payload.ExtraParams,
		"RevisionNum":       payload.RevisionNum,
	}

	jsonData, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata to JSON: %v", err)
	}

	jsonFile := filepath.Join(unitDir, "metadata.json")
	err = os.WriteFile(jsonFile, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write JSON file: %v", err)
	}

	return nil
}

// deleteConfigUnit deletes the config unit files
func (cb *CustomKubernetesBridge) deleteConfigUnit(payload api.BridgePayload) error {
	unitDir := filepath.Join(cb.baseDir, payload.SpaceID.String(), payload.UnitSlug)

	// Remove the entire unit directory
	err := os.RemoveAll(unitDir)
	if err != nil {
		return fmt.Errorf("failed to delete unit directory: %v", err)
	}

	return nil
}
