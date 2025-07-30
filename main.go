// Copyright (C) ConfigHub, Inc.
// SPDX-License-Identifier: MIT

// This application implements a custom bridge worker for ConfigHub that wraps the standard
// Kubernetes bridge and adds file persistence functionality.
package main

import (
	"log"
	"os"

	function "github.com/confighub/sdk/function"
	"github.com/confighub/sdk/worker"
)

func main() {
	// Note: The SDK currently uses controller-runtime logging internally.
	// Until the SDK provides a way to configure logging, you may see warnings about
	// uninitialized loggers. These can be safely ignored for this example.

	// For your own logging, you can use standard log package as shown in this example
	log.Printf("[INFO] Starting custom bridge worker...")

	// Create bridge dispatcher
	bridgeDispatcher := worker.NewBridgeDispatcher()

	// Create custom bridge that wraps the standard Kubernetes bridge
	baseDir := os.Getenv("CUSTOM_BRIDGE_DIR")
	if baseDir == "" {
		baseDir = "/tmp/confighub-custom-bridge"
	}
	log.Printf("[INFO] Using base directory: %s", baseDir)

	customBridge, err := NewCustomKubernetesBridge("custom-kubernetes-bridge", baseDir)
	if err != nil {
		log.Fatalf("Failed to create custom bridge: %v", err)
	}
	bridgeDispatcher.RegisterBridge(customBridge)

	// Create function executor with all standard functions
	executor := function.NewStandardExecutor()

	// The connector is the "engine" of the worker. You register bridges and functions with it.
	// Then you start it and it connects to ConfigHub and offers its local capabilities to your ConfigHub org.
	connector, err := worker.NewConnector(worker.ConnectorOptions{
		WorkerID:         os.Getenv("CONFIGHUB_WORKER_ID"),
		WorkerSecret:     os.Getenv("CONFIGHUB_WORKER_SECRET"),
		ConfigHubURL:     os.Getenv("CONFIGHUB_URL"),
		BridgeDispatcher: &bridgeDispatcher,
		FunctionExecutor: executor,
	})

	if err != nil {
		log.Fatalf("Failed to create connector: %v", err)
	}

	log.Printf("[INFO] Starting connector...")
	err = connector.Start()
	if err != nil {
		log.Fatalf("Failed to start connector: %v", err)
	}
}
