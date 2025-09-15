# MapReduce: A Distributed Word Count Implementation

This project is a functional implementation of the MapReduce programming model, written entirely in Go. It demonstrates the core concepts of distributed data processing through the classic word count example, showcasing how large datasets can be processed in parallel across multiple worker nodes, coordinated by a central master.

## What is MapReduce?

MapReduce is a programming model and an associated implementation for processing and generating large data sets with a parallel, distributed algorithm on a cluster. It was originally developed by Google. The model is inspired by the `map` and `reduce` functions common in functional programming.

A MapReduce job is composed of two main phases:

1.  **Map Phase**: The master node takes the input, partitions it up into smaller sub-problems, and distributes them to worker nodes. A worker node processes its sub-problem and passes the result back to the master.
2.  **Reduce Phase**: The master node collects the answers to all the sub-problems, combines them in some way, and produces the final output.

The framework handles all the complexities of parallelization, fault tolerance, data distribution, and load balancing.

## How This Project Implements MapReduce

This implementation adheres to the core principles of the MapReduce architecture:

*   **Master Process (`cmd/master`)**: The central coordinator. It is responsible for:
    *   Reading input files and creating a set of Map tasks.
    *   Scheduling tasks to available Worker nodes.
    *   Tracking the state of each task (Idle, InProgress, Completed).
    *   Transitioning from the Map phase to the Reduce phase once all Map tasks are complete.
    *   Signaling when the entire job is finished.

*   **Worker Processes (`cmd/worker`)**: The workhorses. A single worker binary can perform both Map and Reduce tasks as assigned by the Master.
    *   **Map Task**: A worker reads an input file, applies the `mapF` function to generate `(word, "1")` key-value pairs, and partitions these pairs into intermediate files using a hash function.
    *   **Reduce Task**: A worker reads all intermediate files corresponding to its partition, sorts the key-value pairs, and applies the `reduceF` function to aggregate the counts for each word, writing the final result to an output file.

*   **Communication**: The Master and Workers communicate via Go's built-in `net/rpc` package. Workers periodically send heartbeats and request tasks, while the Master assigns work and tracks progress.

## Key Features

*   **RPC-Based Communication**: Robust and type-safe communication between the master and workers.
*   **Centralized Task Scheduling**: A thread-safe `TaskQueue` implemented with Go channels ensures that workers receive tasks efficiently and concurrently.
*   **Phase Orchestration**: The master correctly manages the transition from the Map to the Reduce phase, only starting the latter after all Map tasks have successfully completed.
*   **Automated Testing**: The project includes both unit tests for core logic (`go test`) and a full end-to-end integration test script (`test-mr.sh`) that simulates a real-world scenario with multiple workers.

## Project Structure

```
/
├── cmd/
│   ├── master/          # Main application for the master process
│   └── worker/          # Main application for the worker process
├── inputs/              # Directory for input text files (gitignored)
├── outputs/             # Directory for final reduce output (gitignored)
├── tmp/                 # Directory for intermediate files (gitignored)
├── pkg/
│   ├── master/          # Core logic for the master
│   ├── scheduler/       # Task queue implementation
│   ├── types/           # Shared types for tasks and RPC
│   └── worker/          # Core logic for the worker
├── .gitignore
├── go.mod
├── Makefile             # For easy execution of common tasks
└── test-mr.sh           # End-to-end test script
```

## How to Run

### Prerequisites
*   Go (version 1.21 or later)
*   `make`

### Running the Full Test Script

This is the easiest way to see the entire system in action. The script automates building, running the master and workers, and verifying the final output.

```sh
make script
```

This command will:
1.  Clean up any artifacts from previous runs.
2.  Generate a large input file.
3.  Build the `master` and `worker` binaries.
4.  Start the master process in the background.
5.  Start three worker processes in the background.
6.  Wait for the master to signal that the job is complete.
7.  Aggregate the results and verify that the word "mapreduce" is counted correctly.

You should see a `Success: Found 'mapreduce' 7 times, as expected.` message at the end.

### Running Unit Tests

To run the unit tests for the scheduler, master, and worker logic:

```sh
make test
```

This will execute `go test` on all packages within the `pkg/` directory and report a pass or fail status.
