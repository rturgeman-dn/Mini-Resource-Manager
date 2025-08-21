# Mini Resource Manager

A high-performance, in-memory resource manager for integer pools built in Go. This project implements a concurrent allocator using goroutines and channels, exposing a REST API for managing integer resource pools with deterministic allocation policies.

## 🎯 Project Overview

The Mini Resource Manager is designed to efficiently manage integer resources within defined ranges. It uses a "lowest free first" allocation policy and provides a robust REST API for creating templates, pools, and allocating/releasing integer values.

## ✨ Features

- **Concurrent Resource Allocation**: Multi-worker goroutine pool for handling allocation requests
- **Deterministic Allocation Policy**: "Lowest free first" strategy ensures predictable resource assignment
- **RESTful API**: Clean HTTP endpoints for all operations
- **In-Memory Storage**: Fast, thread-safe storage using Go maps with `sync.RWMutex`
- **Graceful Shutdown**: Proper cleanup on SIGINT/SIGTERM signals
- **Request Timeouts**: Built-in timeout handling for long-running operations
- **Metrics Tracking**: Monitor allocations, releases, and timeouts
- **Error Handling**: Comprehensive error responses with appropriate HTTP status codes

## 🏗️ Architecture

### Core Components

```
Mini-Resource-Manager/
├── cmd/server/          # Application entry point
├── internal/
│   ├── core/           # Core types, store, and error definitions
│   ├── api/            # HTTP handlers and routing
│   └── alloc/          # Concurrent allocator implementation
├── go.mod              # Go module definition
└── test.sh             # Integration test script
```

### Key Design Patterns

- **Dependency Injection**: Store and allocator are injected into handlers
- **Channel-based Communication**: HTTP handlers communicate with workers via channels
- **Worker Pool Pattern**: Multiple goroutines handle allocation requests concurrently
- **Mutex-protected Resources**: Thread-safe access to shared data structures

## 🚀 Installation

### Prerequisites

- Go 1.25.0 or newer
- Git

### Setup

1. Clone the repository:
```bash
git clone https://github.com/rturgeman-dn/Mini-Resource-Manager.git
cd Mini-Resource-Manager
```

2. Build the application:
```bash
go build -o server cmd/server/main.go
```

3. Run the server:
```bash
 go run ./cmd/server
```

The server will start on port 8080.

## 📡 API Reference

### Base URL
```
http://localhost:8080
```

### Endpoints

#### 1. Create Template
**POST** `/templates`

Creates a new template defining an integer range.

**Request Body:**
```json
{
  "name": "vlan",
  "min": 100,
  "max": 105
}
```

**Validation Rules:**
- `name` must be non-empty and unique
- `min` must be less than or equal to `max`

**Responses:**
- `201 Created`: Template created successfully
- `400 Bad Request`: Invalid input (min > max)
- `409 Conflict`: Template name already exists

**Flow Diagram:**
```
Client → HTTP POST /templates → main.go (Register Routes) → http.go (mux.HandleFunc) 
    ↓
handlers.go (HandleCreateTemplate):
├── decode JSON from request body
├── validate template (name, min <= max)
├── check if template name already exists
├── store.CreateTemplate() [Lock store.mutex]
├── save to in-memory templates map
└── return 201 Created with template data
```

#### 2. Create Pool
**POST** `/pools`

Creates a new resource pool based on an existing template.

**Request Body:**
```json
{
  "name": "vlan-pool",
  "template": "vlan"
}
```

**Responses:**
- `201 Created`: Pool created successfully
- `404 Not Found`: Template does not exist
- `409 Conflict`: Pool name already exists

**Flow Diagram:**
```
Client → HTTP POST /pools → main.go (Register Routes) → http.go (mux.HandleFunc)
    ↓
handlers.go (HandleCreatePool):
├── decode JSON from request body
├── validate pool name
├── store.TemplateExists() [RLock store.mutex]
├── check if template exists
├── store.PoolExists() [RLock store.mutex]
├── check if pool name already exists
├── create new Pool struct with template min/max
├── initialize pool.InUse map and pool.Next pointer
├── store.CreatePool() [Lock store.mutex]
├── save to in-memory pools map
└── return 201 Created with pool data
```

#### 3. Allocate Resource
**POST** `/allocate`

Allocates a single integer value from a pool using "lowest free first" policy.

**Request Body:**
```json
{
  "pool": "vlan-pool"
}
```

**Responses:**
- `200 OK`: Allocation successful
  ```json
  {
    "value": 100
  }
  ```
- `404 Not Found`: Pool does not exist
- `409 Conflict`: No free items available
- `408 Request Timeout`: Operation timed out

**Flow Diagram:**
```
Client → HTTP POST /allocate → main.go (Register Routes) → http.go (mux.HandleFunc)
    ↓
handlers.go (HandleAllocate):
├── decode JSON from request body
├── validate pool name
├── store.PoolExists() [RLock store.mutex]
├── check if pool exists
├── create context with 10s timeout
├── create replyCh (chan AllocateResponse)
├── build AllocateRequest struct
├── send request to allocator.AllocCh
    ↓
allocator.go (Worker Goroutine):
├── receive AllocateRequest from AllocCh
├── store.PoolExists() [RLock store.mutex]
├── get pool reference
├── pool.Mutex.Lock() [Lock pool for allocation]
├── scan from pool.Next to pool.Max
├── find lowest free value
├── mark value as allocated (pool.InUse[value] = true)
├── update pool.Next pointer
├── increment allocation metrics
├── send AllocateResponse to replyCh
└── pool.Mutex.Unlock()
    ↓
handlers.go (HandleAllocate):
├── receive response from replyCh
├── check for errors (no_free_items, timeout)
└── return 200 OK with allocated value
```

#### 4. Release Resource
**POST** `/release`

Releases a previously allocated integer value back to the pool.

**Request Body:**
```json
{
  "pool": "vlan-pool",
  "value": 100
}
```

**Responses:**
- `200 OK`: Release successful
- `404 Not Found`: Pool does not exist
- `400 Bad Request`: Value out of range
- `409 Conflict`: Value not currently allocated

**Flow Diagram:**
```
Client → HTTP POST /release → main.go (Register Routes) → http.go (mux.HandleFunc)
    ↓
handlers.go (HandleRelease):
├── decode JSON from request body
├── validate pool name and value
├── store.PoolExists() [RLock store.mutex]
├── check if pool exists
├── create context with 10s timeout
├── create replyCh (chan error)
├── build ReleaseRequest struct
├── send request to allocator.ReleaseCh
    ↓
allocator.go (Worker Goroutine):
├── receive ReleaseRequest from ReleaseCh
├── store.PoolExists() [RLock store.mutex]
├── get pool reference
├── pool.Mutex.Lock() [Lock pool for release]
├── validate value is within pool range (min <= value <= max)
├── check if value is currently allocated
├── mark value as free (pool.InUse[value] = false)
├── update pool.Next if value < pool.Next
├── increment release metrics
├── send nil error to replyCh
└── pool.Mutex.Unlock()
    ↓
handlers.go (HandleRelease):
├── receive error from replyCh
├── check for errors (value_out_of_range, not_allocated)
└── return 200 OK with success message
```

#### 5. Get Pool Status
**GET** `/pools/{name}`

Retrieves the current status of a pool.

**Responses:**
- `200 OK`: Pool information
- `404 Not Found`: Pool does not exist

**Flow Diagram:**
```
Client → HTTP GET /pools/{name} → main.go (Register Routes) → http.go (mux.HandleFunc)
    ↓
handlers.go (HandleGetPools):
├── extract pool name from URL path
├── store.PoolExists() [RLock store.mutex]
├── check if pool exists
├── get pool reference from in-memory map
└── return 200 OK with pool data (JSON)
```

## 🧪 Testing

### Running Tests

```bash
# Start the server first
 go run ./cmd/server

# In another terminal, run the test script
chmod +x test.sh
./test.sh
```

The test script covers:
- Template and pool creation
- Happy path allocations and releases
- Error handling scenarios
- Concurrent allocation testing
- Timeout and backpressure testing

### Test Scenarios

The integration tests verify:
- ✅ Template creation with validation
- ✅ Pool creation from templates
- ✅ Sequential allocation and release
- ✅ Concurrent allocation handling
- ✅ Error conditions (invalid inputs, missing resources)
- ✅ Timeout handling
- ✅ Resource exhaustion scenarios

## 🔧 Configuration

### Server Configuration

The server runs on port 8080 by default. Key configuration parameters:

- **Worker Count**: 3 concurrent workers (configurable in `main.go`)
- **Channel Buffer Size**: 64 requests per channel
- **Request Timeout**: 10 seconds
- **Shutdown Timeout**: 30 seconds


## 📊 Metrics

The application tracks the following metrics:
- **Allocations**: Total number of successful allocations
- **Releases**: Total number of successful releases
- **Timeouts**: Number of allocation timeouts

Metrics are displayed on graceful shutdown.

## 🛡️ Error Handling

The application provides comprehensive error handling:

- **Input Validation**: All requests are validated before processing
- **Resource Management**: Proper handling of resource exhaustion
- **Concurrency Safety**: Thread-safe operations with proper locking
- **Timeout Management**: Graceful handling of long-running operations
- **Graceful Degradation**: System continues operating under load

## 🔄 Concurrency Model

### Worker Pool Architecture

```
HTTP Handlers → Request Channels → Worker Goroutines → Response Channels
```

- **3 Worker Goroutines**: Handle allocation and release requests
- **Bounded Channels**: Prevent memory exhaustion under load
- **Context Timeouts**: Ensure requests don't hang indefinitely
- **Mutex Protection**: Thread-safe access to shared resources

### Allocation Algorithm

The "lowest free first" algorithm:
1. Start from the `Next` pointer in the pool
2. Scan forward until finding an unallocated value
3. Mark the value as allocated
4. Update the `Next` pointer for future allocations
5. Return the allocated value

## 🚀 Performance Characteristics

- **High Throughput**: Concurrent processing with worker pools
- **Low Latency**: In-memory operations with minimal overhead
- **Scalable**: Configurable worker count and buffer sizes
- **Deterministic**: Predictable allocation patterns