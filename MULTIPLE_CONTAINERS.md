# Multiple Container Support - Implementation Guide

## Overview

The Redway project has been enhanced to support multiple LXC containers running independently. The key architectural change separates **LXC system preparation** from **individual container initialization**.

## Architecture Changes

### 1. LXC System Preparation (One-Time Setup)

**New Component: `LXCPreparer`**

The `LXCPreparer` handles all system-level LXC setup that only needs to be done once:

- Checking kernel modules (binder support)
- Checking LXC tools availability
- Setting up LXC networking (lxcbr0 bridge)
- Configuring NAT and IP forwarding
- Adjusting OCI template
- Checking required tools (skopeo, umoci, jq)

**Usage:**

```bash
redway prepare-lxc
```

This command sets `LXCReady: true` in the config, preventing redundant setup.

### 2. Container Initialization

**Enhanced Component: `Initializer`**

The `Initializer` now handles individual container setup:

- Creates the LXC container from OCI image
- Fixes container filesystem
- Creates data directory
- Adjusts container configuration
- Applies Redroid-specific networking workarounds

**Usage:**

```bash
redway init [container-name] [image-url]
```

## Configuration Structure

### Old Structure

```json
{
  "container_name": "redroid",
  "image_url": "...",
  "data_path": "...",
  "log_file": "...",
  "gpu_mode": "...",
  "initialized": true
}
```

### New Structure

```json
{
  "lxc_ready": true,
  "containers": {
    "redroid": {
      "name": "redroid",
      "image_url": "...",
      "data_path": "...",
      "log_file": "...",
      "gpu_mode": "...",
      "initialized": true
    },
    "android1": {
      "name": "android1",
      "image_url": "...",
      "data_path": "...",
      "log_file": "...",
      "gpu_mode": "...",
      "initialized": true
    }
  }
}
```

## Usage Examples

### First-Time Setup

```bash
# 1. Prepare LXC system (one-time)
redway prepare-lxc

# 2. Create default container
redway init

# 3. Start container
redway start
```

### Multiple Containers

```bash
# Create multiple containers with different images
redway init android1 docker://redroid/redroid:15.0.0_64only-latest
redway init android2 docker://redroid/redroid:16.0.0_64only-latest

# Start them independently
redway start android1
redway start android2

# List all containers
redway list

# Manage each container separately
redway status android1
redway adb-connect android2
redway shell android1
redway stop android2
redway restart android1
```

### Container Operations

All commands now support optional container name parameter:

```bash
# Default container (redroid)
redway start
redway stop
redway status
redway shell
redway adb-connect
redway log

# Specific container
redway start android1
redway stop android2
redway status android1
redway shell android2
redway adb-connect android1
redway log android2
```

## File Structure

### Modified Files

1. **pkg/config/config.go**
   - New `Container` struct for individual container configuration
   - Updated `Config` struct with `Containers` map and `LXCReady` flag
   - Helper methods: `GetContainer()`, `AddContainer()`, `RemoveContainer()`, `ListContainers()`

2. **pkg/container/initializer.go**
   - New `LXCPreparer` struct for system-level setup
   - Enhanced `Initializer` struct for container-specific setup
   - Separated methods for LXC preparation vs container initialization

3. **pkg/container/manager.go**
   - Updated `Manager` to support multiple containers via `containerName` field
   - New `NewManagerForContainer()` function
   - Enhanced `Lister` with `ListRedwayContainers()` method

4. **pkg/utils/status.go**
   - Updated to accept `containerName` parameter
   - Shows status for specific container

5. **pkg/utils/shell.go**
   - Updated to accept `containerName` parameter
   - Enters shell of specific container

6. **pkg/utils/adb.go**
   - Updated to accept `containerName` parameter
   - Shows ADB connection for specific container

7. **pkg/utils/log.go**
   - Updated to accept `containerName` parameter
   - Shows logs for specific container

8. **cmd/command.go**
   - New `prepare-lxc` command
   - Updated all commands to parse container name from arguments
   - Default to `DefaultContainerName` if not specified

9. **cmd/usage.go**
   - Updated with new command documentation
   - Added examples for multiple container usage

## Key Features

### 1. Independent Container Execution

Each container has its own:

- Data directory
- Log file
- Configuration
- Network namespace
- Filesystem

### 2. Backward Compatibility

- Default container name is "redroid"
- Commands work without specifying container name (uses default)
- Existing workflows continue to work

### 3. Efficient System Setup

- LXC system preparation runs only once
- Subsequent containers skip system setup
- Reduces initialization time for multiple containers

### 4. Container Management

- List all managed containers
- View status of each container
- Start/stop containers independently
- Remove containers individually

## Implementation Details

### Container Isolation

Each container is isolated through:

- Separate LXC container directories (`/var/lib/lxc/<name>`)
- Separate data directories (`~/data-<name>`)
- Separate log files (`<name>.log`)
- Independent LXC configuration

### Configuration Persistence

The config file stores:

- `LXCReady`: Whether system is prepared
- `Containers`: Map of all managed containers

This allows:

- Skipping LXC preparation on subsequent runs
- Tracking multiple containers
- Persisting container state

### Error Handling

- Validates container exists before operations
- Provides clear error messages
- Checks if LXC system is prepared before initialization

## Migration from Single to Multiple Containers

If you have an existing single-container setup:

1. The system automatically migrates on first run
2. Existing container becomes the default "redroid"
3. New containers can be added without affecting existing one
4. All existing commands continue to work

## Performance Considerations

- **LXC Preparation**: ~30-60 seconds (one-time)
- **Container Initialization**: ~2-5 minutes per container
- **Container Startup**: ~10-30 seconds per container
- **Multiple Containers**: Can run simultaneously with proper resource allocation

## Future Enhancements

Potential improvements:

- Container templates/profiles
- Resource limits per container
- Container networking between instances
- Automated backup/restore
- Container cloning
- Batch operations
