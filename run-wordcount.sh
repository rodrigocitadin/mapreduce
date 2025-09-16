#!/bin/bash

# Exit immediately if a command exits with a non-zero status.
set -e

# This script runs a full word-count MapReduce job.
# It starts one master and three workers, submits a job with the
# contents of the 'inputs' directory, and waits for completion.

# Function to clean up background processes on exit
cleanup() {
    echo "Cleaning up background processes..."
    # Kill any running worker processes. The master exits on its own.
    pkill -f "wordcount.go --role=worker" 2>/dev/null || true
    echo "Cleanup complete."
}

trap cleanup EXIT

echo "Starting MapReduce run for word count..."

# 1. Clean up previous run's temporary files and outputs
echo "Step 1: Cleaning up previous run..."
rm -rf tmp/ outputs/
mkdir -p tmp/ outputs/

# 2. Start the master process (logs will be shown on this terminal)
echo "Step 2: Starting master..."
go run ./examples/wordcount/wordcount.go --role=master &
MASTER_PID=$!

# Wait a moment for the master to start up and listen for connections
sleep 1

# 3. Start worker processes (logs are ignored)
WORKER_COUNT=3
echo "Step 3: Starting $WORKER_COUNT workers (logs will be discarded)..."
for i in $(seq 1 $WORKER_COUNT); do
    port=$((1235 + i))
    go run ./examples/wordcount/wordcount.go --role=worker --masterAddr="localhost:9991" --workerAddr="localhost:$port" > /dev/null 2>&1 &
done

# 4. Submit the job to the master
echo "Step 4: Submitting job with files from inputs/ directory..."
go run ./examples/wordcount/wordcount.go --role=submit --masterAddr="localhost:9991" --inputs ./inputs

# 5. Wait for the job to complete by waiting for the master process to exit
echo "Step 5: Waiting for job to complete... (monitoring master logs)"
wait $MASTER_PID

echo ""
echo "-------------------------------------"
echo "MapReduce job complete!"
echo "Final results are in the 'outputs/' directory."

