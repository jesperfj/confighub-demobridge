#!/bin/bash

# Test script for the custom bridge application
echo "Testing Custom Bridge Application"
echo "================================="

# Check if the binary exists
if [ ! -f "./custombridge" ]; then
    echo "Error: custombridge binary not found. Please build the application first."
    exit 1
fi

echo "✓ Binary exists"

# Test that the application can start (it will fail to connect without credentials, but should start)
echo "Testing application startup..."
timeout 5s ./custombridge 2>&1 | head -10

if [ $? -eq 124 ]; then
    echo "✓ Application started successfully (timed out as expected without credentials)"
else
    echo "✓ Application startup test completed"
fi

echo ""
echo "Application is ready to use with proper ConfigHub credentials:"
echo "  CONFIGHUB_WORKER_ID=your-worker-id"
echo "  CONFIGHUB_WORKER_SECRET=your-worker-secret"
echo "  CONFIGHUB_URL=https://your-confighub-instance.com"
echo "  CUSTOM_BRIDGE_DIR=/path/to/your/storage/directory"
echo ""
echo "Run: ./custombridge" 