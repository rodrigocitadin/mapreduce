# MapReduce: A Distributed Word Count Implementation

This project is a functional implementation of the MapReduce programming model, written entirely in Go. It provides a core library for distributed data processing and includes a classic word count example to demonstrate its capabilities. The system processes large datasets in parallel across multiple worker nodes, coordinated by a central master.

## What is MapReduce?

MapReduce is a programming model for processing and generating large data sets with a parallel, distributed algorithm on a cluster. A MapReduce job is composed of two main phases:

1.  **Map Phase**: The master node takes the input, partitions it into smaller sub-problems, and distributes them to worker nodes. Each worker applies a `map` function to its sub-problem and writes the intermediate results.
2.  **Reduce Phase**: The master node signals the workers to start the reduce phase. Each worker collects the intermediate results relevant to its partition, applies a `reduce` function to aggregate the data, and produces the final output.

The framework handles the complexities of parallelization, fault tolerance, data distribution, and load balancing.

## How This Project Implements MapReduce

This implementation is structured as a library (`pkg/`) and a standalone example (`examples/wordcount`):

*   **Master (`pkg/master`)**: The core logic for the central coordinator. It is responsible for:
    *   Accepting job submissions via RPC.
    *   Scheduling Map and Reduce tasks to available Worker nodes.
    *   Tracking the state of each task (Idle, InProgress, Completed).
    *   Transitioning from the Map phase to the Reduce phase.
    *   Signaling when the entire job is finished.

*   **Worker (`pkg/worker`)**: The core logic for a worker process. A single worker can perform both Map and Reduce tasks as assigned by the Master. The worker is generic and receives the `mapF` and `reduceF` functions from the specific application (e.g., word count) that uses it.

*   **Word Count Example (`examples/wordcount`)**: A self-contained application that demonstrates the library's usage. It can be run in one of three roles:
    1.  **Master**: Starts the master server.
    2.  **Worker**: Starts a worker process, providing it with the word count `map` and `reduce` functions.
    3.  **Submit**: Connects to the master and submits a job with a directory of input files.

## Key Features

*   **Library-Based Design**: The core MapReduce logic is decoupled from the specific application, making the framework reusable for other problems.
*   **RPC-Based Communication**: Robust and type-safe communication between the master and workers using Go's built-in `net/rpc` package.
*   **Centralized Task Scheduling**: A thread-safe `TaskQueue` implemented with Go channels ensures that workers receive tasks efficiently and concurrently.
*   **Automated Testing**: The project includes unit tests for core library logic (`make test`) and a full end-to-end execution script (`make script`).

## Project Structure

```
/
├── examples/
│   └── wordcount/       # Self-contained word count application
├── inputs/              # Directory for input text files
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
└── run-wordcount.sh     # End-to-end execution script
```

## How to Run

### Prerequisites
*   Go (version 1.21 or later)
*   `make`

### Running the Full Example (Recommended)

This is the easiest way to see the entire system in action. The script automates starting the master and workers, submitting the job, and waiting for completion.

```sh
make script
```

This command will:
1.  Clean up artifacts from any previous runs.
2.  Start one master process in the background (with logs visible in your terminal).
3.  Start three worker processes in the background (with logs discarded).
4.  Submit a job using the files in the `inputs/` directory.
5.  Wait for the master to signal that the job is complete.

After execution, you can inspect the final results in the `outputs/` directory.

### Running Unit Tests

To run the unit tests for the core library packages:

```sh
make test
```

### Running Manually (Step-by-Step)

For debugging or a more hands-on demonstration, you can run the components in separate terminals:

1.  **Terminal 1: Start the Master**
    ```sh
    make master
    ```

2.  **Terminal 2 (and 3, 4...): Start one or more Workers**
    ```sh
    make wordcount-worker
    ```

3.  **Terminal 5: Submit the Job**
    ```sh
    make wordcount-submit
    ```
