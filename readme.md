# ViscauFS (Virtual Userspace FileSystem)

A remote filesystem optimizing container cold starts in microVMs through lazy loading and centralized storage.

## Why?

Studies show containers waste 76% of startup time downloading packages, yet only 6.4% of data is actually read. Current solutions require full image downloads before container startup, creating unnecessary latency and resource usage.

## How it Works

```
┌─────────────┐         ┌──────────────┐         ┌─────────────┐
│   MicroVM   │         │  ViscaUFS    │         │  Container  │
│  with FUSE  │◄──gRPC──┤   Server     │◄────────┤  Registry   │
│   Client    │         │              │         │             │
└─────────────┘         └──────────────┘         └─────────────┘
                              │
                        ┌─────┴──────┐
                        │   Layer    │
                        │  Indexes   │
                        └────────────┘
```

### Components

1. **FUSE Client**
    - Mounts remote filesystem locally
    - Handles file operations via gRPC
    - Transparent to applications

2. **Server**
    - Downloads container layers in background
    - Indexes filesystem image
    - Serves file requests immediately after the first layer is indexed
    - Uses Adaptive Radix Tree for fast lookups


## Smart Layer Processing

### Layer Architecture
```
Image: nginx:latest
┌────────────────────┐
│ Layer 3 (app)      │ Each layer is:
├────────────────────┤ - Independently indexed
│ Layer 2 (nginx)    │ - Cached and reusable
├────────────────────┤ - Processed in parallel
│ Layer 1 (base)     │ - Immediately searchable
└────────────────────┘
```

### Key Features

1. **Independent Layer Indexing**
    - Each layer is processed and indexed separately
    - Indexes are cached and reused across different images
    - Uses Adaptive Radix Tree for efficient lookups
    - Handles whiteouts and opaque directories

2. **Progressive Layer Merging**
   ```
   Base Layer    → Index A
   + Layer 2     → Merge(Index A, B)
   + Layer 3     → Merge(Index AB, C)
   = Final Index → Instant file lookups
   ```

3. **Intelligent Caching**
    - Layer indexes stored in SQLite
    - Shared between multiple images
    - Instant reuse for common base layers
    - Minimizes redundant processing

## Project Structure

```
├── common
│   ├── proto/           # Protocol buffers GRPC definitions
│   ├── fsindex/         # Indexing library for filesystem images
├── filesystem/          # FUSE client implementation
├── server/
│   ├── internal/
│   │   ├── service/     # Core services
│   │   └── types/       # Data models
```

## Roadmap
- Filesystem: Build a DNS SRV Content aware picker load balancer for gRPC
- Server, Filesystem: Store statistics about images X layers to allow preload in priority files that ares often used
- Server: Track Image usage
- Server: GC of non used image or orphelin layer
- Server, Filesystem: Use statistics to preload files that are frequently used
- Server, Filesystem: Create a unix socket transport for local deployment (viscaufs on the node server)
