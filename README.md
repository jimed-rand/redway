# Redway - LXC Container Manager for Redroid

Redway is a lightweight Go-based **LXC container manager** for
[redroid](https://github.com/remote-android/redroid-doc). It provides a simple
CLI interface for managing redroid Android containers using native LXC with OCI
image support.

## Key Features

- **Pure LXC Implementation** - Uses native LXC containers, not Docker
- **OCI Image Support** - Fetches images from OCI registries via skopeo/umoci
- **Cross-Distro Compatible** - Works on Ubuntu, Debian, Fedora, Arch, openSUSE,
  Gentoo, and more
- **Smart Detection** - Auto-detects binder support (module, built-in, or
  binderfs)
- **Simple CLI Interface** - Easy-to-use command-line tool
- **ADB Integration** - Built-in ADB connection management
- **Persistent Storage** - Android data persists across container restarts
- **Lightweight Single Binary** - No runtime dependencies, just system tools

## What is Redroid?

Redroid (Remote Android) is a GPU accelerated AIC (Android In Container)
solution that allows you to run Android in containers with near-native
performance.

## Architecture

Redway uses a pure LXC-based approach:

```
┌─────────────────────────────────────────────────┐
│                   Redway CLI                    │
├─────────────────────────────────────────────────┤
│                                                 │
│  ┌─────────┐    ┌─────────┐    ┌─────────────┐ │
│  │ skopeo  │───►│  umoci  │───►│ lxc-create  │ │
│  │(fetch)  │    │(unpack) │    │  (OCI tpl)  │ │
│  └─────────┘    └─────────┘    └─────────────┘ │
│                                      │          │
│                              ┌───────▼────────┐ │
│                              │  LXC Container │ │
│                              │   (redroid)    │ │
│                              └────────────────┘ │
└─────────────────────────────────────────────────┘
```

**How it works:**

1. `skopeo` fetches OCI images from container registries
2. `umoci` unpacks the OCI image into a rootfs
3. `lxc-create` creates an LXC container using the OCI template
4. LXC manages the container lifecycle (start/stop/restart)

## Prerequisites

### System Requirements

- Linux kernel with binder support (ashmem not required for modern redroid)
- LXC tools (lxc-utils)
- Root/sudo privileges

### Binder Support

Redroid requires **binder** kernel support. Different distributions provide this
in different ways:

| Distribution   | Binder Support             |
| -------------- | -------------------------- |
| Ubuntu/Debian  | Module (`binder_linux`)    |
| Fedora         | Module (`binder_linux`)    |
| Arch Linux     | Built-in or module         |
| Custom kernels | Often built-in or binderfs |

Redway will check for binder support during initialization and provide guidance
if not detected.

#### Checking Binder Support

```bash
# Check for binder (any of these indicates support)
ls -la /dev/binder* /dev/binderfs 2>/dev/null
lsmod | grep binder
```

#### Loading Modules (if needed)

##### Ubuntu/Debian

```bash
sudo apt install linux-modules-extra-$(uname -r)
sudo modprobe binder_linux devices="binder,hwbinder,vndbinder"
```

##### Fedora

```bash
sudo dnf install kernel-modules-extra
sudo modprobe binder_linux devices="binder,hwbinder,vndbinder"
```

##### Arch Linux

```bash
# Only if modules aren't built-in
sudo modprobe binder_linux devices="binder,hwbinder,vndbinder"
```

#### Persistent Module Loading (if using modules)

```bash
# Create module configuration
sudo tee /etc/modules-load.d/redway.conf <<EOF
binder_linux
EOF

sudo tee /etc/modprobe.d/redway.conf <<EOF
options binder_linux devices="binder,hwbinder,vndbinder"
EOF
```

> **Note:** If your kernel has built-in binder support (common on Arch and
> custom kernels), you don't need to load modules. Redway will detect this
> automatically.

### Required Packages

Install these packages for your distribution:

#### Ubuntu/Debian

```bash
sudo apt update
sudo apt install -y lxc-utils skopeo umoci jq
```

#### Arch Linux

```bash
sudo pacman -S lxc skopeo umoci jq
```

#### Fedora

```bash
sudo dnf install -y lxc lxc-templates skopeo umoci jq
```

#### openSUSE

```bash
sudo zypper install lxc skopeo umoci jq
```

## Installation

### From Source

```bash
cd redway
make build
sudo make install
```

### Manual Compilation

```bash
go build -ldflags="-s -w" -o redway .
sudo cp redway /usr/local/bin/
sudo chmod +x /usr/local/bin/redway
```

### Static Binary

```bash
make static
sudo cp redway /usr/local/bin/
```

## Quick Start

### 1. Load Kernel Modules

**Before using Redway, you must load the required kernel modules:**

```bash
sudo modprobe binder_linux devices="binder,hwbinder,vndbinder"
sudo modprobe ashmem_linux
```

Verify:

```bash
lsmod | grep binder
```

### 2. Initialize LXC Container

Using default image (Android 13):

```bash
sudo redway init
```

Using specific image:

```bash
sudo redway init docker://redroid/redroid:12.0.0_64only-latest
```

Available redroid images:

- `redroid/redroid:13.0.0-latest` - Android 13 (arm64)
- `redroid/redroid:13.0.0_64only-latest` - Android 13 (x86_64)
- `redroid/redroid:12.0.0-latest` - Android 12 (arm64)
- `redroid/redroid:12.0.0_64only-latest` - Android 12 (x86_64)
- `redroid/redroid:11.0.0-latest` - Android 11

> **Note:** The `docker://` prefix is used by skopeo to identify OCI registry
> sources. This does NOT mean Docker is required - Redway uses native LXC
> containers.

### 3. Start Container

```bash
sudo redway start
```

### 4. Connect via ADB

Get connection information:

```bash
redway adb-connect
```

Connect:

```bash
adb connect <IP>:5555
```

### 5. Install Apps

```bash
adb install app.apk
```

## Usage

### Container Management

```bash
sudo redway init [image]     # Initialize new LXC container
sudo redway start            # Start container
sudo redway stop             # Stop container
sudo redway restart          # Restart container
sudo redway remove           # Remove container
```

### Information & Debugging

```bash
redway status                # Show container status
redway list                  # List all LXC containers
redway log                   # Show container logs
redway adb-connect           # Display ADB connection info
```

### Direct Access

```bash
sudo redway shell            # Enter container via nsenter
```

## Configuration

Configuration is stored in `~/.config/redway/config.json`:

```json
{
  "container_name": "redroid",
  "image_url": "docker://redroid/redroid:13.0.0_64only-latest",
  "data_path": "/home/user/data-redroid",
  "log_file": "redroid.log",
  "gpu_mode": "guest",
  "initialized": true
}
```

### Configuration Options

| Option           | Description                   | Default                                         |
| ---------------- | ----------------------------- | ----------------------------------------------- |
| `container_name` | LXC container name            | `redroid`                                       |
| `image_url`      | OCI image reference           | `docker://redroid/redroid:13.0.0_64only-latest` |
| `data_path`      | Android data persistence path | `~/data-redroid`                                |
| `log_file`       | Log file location             | `redroid.log`                                   |
| `gpu_mode`       | GPU rendering mode            | `guest`                                         |

### GPU Modes

- `guest` - Software rendering (default, most compatible)
- `host` - GPU passthrough (requires compatible GPU)

## Project Structure

```
redway/
├── main.go                  # Entry point
├── go.mod                   # Go module definition
├── Makefile                 # Build system
├── cmd/                     # CLI commands
│   ├── command.go          # Command dispatcher
│   └── usage.go            # Help text
└── pkg/                     # Core packages
    ├── config/             # Configuration management
    │   └── config.go
    ├── container/          # LXC container operations
    │   ├── initializer.go  # Container setup
    │   └── manager.go      # Lifecycle management
    └── utils/              # Utility functions
        ├── adb.go          # ADB integration
        ├── shell.go        # Shell access
        ├── status.go       # Status reporting
        └── log.go          # Log viewing
```

## Directory Structure

```
~/.config/redway/
└── config.json              # Redway configuration

~/data-redroid/              # Android persistent data
├── data/                    # App data
└── system/                  # System data

/var/lib/lxc/redroid/        # LXC container
├── config                   # Container configuration
└── rootfs/                  # Android filesystem
```

## LXC Integration Details

Redway leverages LXC's OCI template to create containers from OCI images:

### Container Creation Flow

1. **Fetch Image** - `skopeo` downloads the OCI image from registry
2. **Unpack Image** - `umoci` extracts the image to a rootfs
3. **Create Container** - `lxc-create -t oci` creates the LXC container
4. **Configure** - Redway adjusts LXC config for Android compatibility
5. **Apply Workarounds** - Network and other Android-specific fixes

### LXC Configuration

Redway configures LXC containers with:

```
lxc.init.cmd = /init androidboot.hardware=redroid androidboot.redroid_gpu_mode=guest
lxc.apparmor.profile = unconfined
lxc.autodev = 1
lxc.autodev.tmpfs.size = 25000000
lxc.mount.entry = /home/user/data-redroid data none bind 0 0
```

## Troubleshooting

### Container won't start

1. Check kernel modules:

```bash
lsmod | grep binder
lsmod | grep ashmem
```

2. Load modules if missing:

```bash
sudo modprobe binder_linux devices="binder,hwbinder,vndbinder"
sudo modprobe ashmem_linux
```

3. Check logs:

```bash
redway log
```

### Module not found

Install kernel modules for your distribution:

```bash
# Ubuntu/Debian
sudo apt install linux-modules-extra-$(uname -r)

# Arch
sudo pacman -S linux-headers

# Fedora
sudo dnf install kernel-modules-extra
```

### No IP address

Check LXC networking:

```bash
sudo lxc-info redroid
```

### ADB connection fails

```bash
adb kill-server
adb start-server
adb connect <IP>:5555
```

### Permission issues

Always use sudo for container operations:

```bash
sudo redway start
```

## Advanced Usage

### Custom Data Directory

Edit `~/.config/redway/config.json`:

```json
{
  "data_path": "/mnt/storage/android-data"
}
```

### Multiple Containers

Edit container name in config before initializing:

```json
{
  "container_name": "redroid-test"
}
```

### Direct LXC Access

```bash
# Using nsenter
PID=$(sudo lxc-info redroid -p | awk '{print $2}')
sudo nsenter -t $PID -a sh

# Using lxc-attach
sudo lxc-attach -n redroid

# List LXC containers
sudo lxc-ls -f
```

## Distribution-Specific Notes

### Ubuntu/Debian

- Works out of the box with `linux-modules-extra`
- LXC networking usually pre-configured

### Arch Linux

- May need to enable `lxc.service`
- Install `linux-headers` for current kernel

### Fedora

- SELinux may need configuration
- Use `kernel-modules-extra` package

### openSUSE

- May require manual bridge setup
- Check firewall rules for LXC

## Known Limitations

- Requires kernel with binder/ashmem support
- Root privileges required for operations
- Kernel modules must be loaded manually
- GPU passthrough is hardware-dependent

## Why LXC Instead of Docker?

Redway uses LXC for several reasons:

1. **Lightweight** - No daemon required, direct kernel integration
2. **System Containers** - LXC is designed for full system containers like
   Android
3. **Fine-grained Control** - Direct access to container configuration
4. **No Extra Services** - No Docker daemon overhead
5. **Native Linux** - Uses standard Linux container primitives

## Contributing

Contributions welcome! Please ensure:

- Code follows Go standards
- Cross-distro compatibility maintained
- Documentation updated
- No hardcoded kernel module loading

## License

GPL-2.0 License

## Credits

- [redroid](https://github.com/remote-android/redroid-doc) - Remote-Android
  project
- [LXC](https://linuxcontainers.org/) - Linux Containers
- [skopeo](https://github.com/containers/skopeo) - OCI image operations
- [umoci](https://github.com/opencontainers/umoci) - OCI image unpacker

## Links

- Redroid Documentation: https://github.com/remote-android/redroid-doc
- LXC Documentation: https://linuxcontainers.org/lxc/documentation/
- OCI Image Spec: https://github.com/opencontainers/image-spec
