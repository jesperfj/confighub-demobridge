// Copyright (C) ConfigHub, Inc.
// SPDX-License-Identifier: MIT

// This application implements a custom bridge worker for ConfigHub that wraps the standard
// Kubernetes bridge and adds file persistence functionality.
package main

import (
	"context"
	"log"
	"os"

	"github.com/confighub/sdk/bridge-worker/api"
	"github.com/confighub/sdk/bridge-worker/lib"
	"github.com/confighub/sdk/function"
	funcApi "github.com/confighub/sdk/function/api"
	"go.opentelemetry.io/otel/metric/noop"
)

func main() {
	// Note: The SDK currently uses controller-runtime logging internally.
	// Until the SDK provides a way to configure logging, you may see warnings about
	// uninitialized loggers. These can be safely ignored for this example.

	// For your own logging, you can use standard log package as shown in this example
	log.Printf("[INFO] Starting custom bridge worker...")

	// Create custom bridge that wraps the standard Kubernetes bridge
	baseDir := os.Getenv("CUSTOM_BRIDGE_DIR")
	if baseDir == "" {
		baseDir = "/tmp/confighub-custom-bridge"
	}
	saveOnly := os.Getenv("SAVE_ONLY") != ""
	log.Printf("[INFO] Using base directory: %s (save-only: %v)", baseDir, saveOnly)

	customBridge, err := NewCustomKubernetesBridge("custom-kubernetes-bridge", baseDir, saveOnly)
	if err != nil {
		log.Fatalf("Failed to create custom bridge: %v", err)
	}

	// Create function executor with all standard functions
	executor := function.NewStandardExecutor()

	// Use lib.Worker directly with a no-op meter to work around
	// https://github.com/confighubai/confighub/issues/3669
	meter := noop.NewMeterProvider().Meter("")
	worker := lib.New(os.Getenv("CONFIGHUB_URL"), os.Getenv("CONFIGHUB_WORKER_ID"), os.Getenv("CONFIGHUB_WORKER_SECRET")).
		WithBridgeWorker(customBridge).
		WithFunctionWorker(&functionWorkerAdapter{executor: executor}).
		WithMetricsMeter(meter)

	log.Printf("[INFO] Starting worker...")
	if err := worker.Start(context.Background()); err != nil {
		log.Fatalf("Failed to start worker: %v", err)
	}
}

// functionWorkerAdapter wraps FunctionExecutor to satisfy api.FunctionWorker
type functionWorkerAdapter struct {
	executor *function.FunctionExecutor
}

func (a *functionWorkerAdapter) Info() api.FunctionWorkerInfo {
	return api.FunctionWorkerInfo{
		SupportedFunctions: a.executor.RegisteredFunctions(),
	}
}

func (a *functionWorkerAdapter) Invoke(ctx api.FunctionWorkerContext, req funcApi.FunctionInvocationRequest) (funcApi.FunctionInvocationResponse, error) {
	resp, err := a.executor.Invoke(ctx.Context(), &req)
	if err != nil {
		return funcApi.FunctionInvocationResponse{}, err
	}
	if resp == nil {
		return funcApi.FunctionInvocationResponse{}, nil
	}
	return *resp, nil
}
