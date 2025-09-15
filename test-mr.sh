#!/bin/bash

# Exit immediately if a command exits with a non-zero status.
set -e

# A simple test script for the MapReduce 
# It builds the master and worker, runs them with a larger input,
# and checks for the final output.

# Function to clean up background processes on exit
cleanup() {
    echo "Cleaning up background processes..."
    # Kill any running worker processes. The master exits on its own.
    pkill -f worker 2>/dev/null || true
    # Remove the compiled binaries
    rm -f master worker
    echo "Cleanup complete. You can inspect the results in the 'outputs/', 'tmp/', and 'final-output.txt' files."
}

# Trap EXIT signal to ensure cleanup runs
trap cleanup EXIT

echo "Starting MapReduce test..."

# 1. Clean up previous state and create directories
echo "Step 1: Cleaning up previous run..."
rm -rf tmp/ outputs/ final-output.txt inputs/
mkdir -p inputs/ tmp/ outputs/

# 2. Generate a larger input file
echo "Step 2: Generating input file..."
# A simple text with many repeated words
cat > inputs/large-input.txt <<'EOF'
mapreduce is a programming model for processing large data sets
mapreduce is inspired by the map and reduce functions common in functional programming
the mapreduce model consists of a map procedure which performs filtering and sorting
and a reduce method which performs a summary operation
the mapreduce system orchestrates the processing by marshalling the distributed servers
running the various tasks in parallel managing all communications and data transfers
between the various parts of the system and providing for redundancy and fault tolerance
mapreduce is a framework for processing parallelizable problems across large datasets
using a large number of computers collectively referred to as a cluster
mapreduce can take terabytes of data as input and produce petabytes of data as output
a mapreduce job usually splits the input data set into independent chunks
which are processed by the map tasks in a completely parallel manner
the framework sorts the outputs of the maps which are then input to the reduce tasks
typically both the input and the output of the job are stored in a file system
the framework takes care of scheduling tasks monitoring them and re-executes the failed tasks
EOF

# 3. Build master and worker
echo "Step 3: Building master and worker..."
go build -o master cmd/master/main.go
go build -o worker cmd/worker/main.go

# 4. Start master
echo "Step 4: Starting master..."
./master &
MASTER_PID=$!

# Wait a moment for the master to start up
sleep 1

# 5. Start workers
WORKER_COUNT=3
echo "Step 5: Starting $WORKER_COUNT workers..."
for i in $(seq 1 $WORKER_COUNT); do
    port=$((1234 + i))
    # Use masterAddr flag, which is defined in worker's main.go
    ./worker --masterAddr="localhost:9991" --workerAddr="localhost:$port" &
done

# 6. Wait for the job to complete
echo "Step 6: Waiting for the MapReduce job to complete..."
# The master process will exit once the job is done.
wait $MASTER_PID
echo "Master process finished. Job should be complete."

# 7. Check the output
echo "Step 7: Checking output..."
if [ ! -d "outputs" ] || [ -z "$(ls -A outputs)" ]; then
    echo "Error: Outputs directory is empty or does not exist."
    exit 1
fi

# Aggregate and sort the results
cat outputs/mr-out-* | sort -k1 > final-output.txt

echo "Final aggregated output (in final-output.txt):"
cat final-output.txt

# A simple check for a known word count
COUNT_MAPREDUCE=$(grep "mapreduce" final-output.txt | awk '{print $2}')
if [ "$COUNT_MAPREDUCE" -eq 7 ]; then
    echo "Success: Found 'mapreduce' 7 times, as expected."
else
    echo "Failure: Expected 'mapreduce' count to be 7, but got $COUNT_MAPREDUCE."
    exit 1
fi

echo "MapReduce test completed successfully!"
# The cleanup function will be called automatically on exit.
