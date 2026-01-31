# Architecture Overview: Multiple Container Support

## System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Redway CLI                              │
│                      (cmd/command.go)                           │
└────────────────────────────┬────────────────────────────────────┘
                             │
                ┌────────────┼────────────┐
                │            │            │
         ┌──────▼──────┐ ┌──▼──────┐ ┌──▼──────────┐
         │  prepare-lxc│ │  init   │ │  start/stop │
         │  (one-time) │ │ (per    │ │  (per       │
         │             │ │container)│ │ container) │
         └──────┬──────┘ └──┬──────┘ └──┬──────────┘
                │           │           │
         ┌──────▼───────────▼───────────▼──────┐
         │    Container Management Layer       │
         │  (pkg/container/initializer.go)     │
         │  (pkg/container/manager.go)         │
         └──────┬──────────────────────────────┘
                │
         ┌──────▼──────────────────────────────┐
         │    Configuration Layer              │
         │  (pkg/config/config.go)             │
         │                                     │
         │  ┌─────────────────────────────┐   │
         │  │ Config                      │   │
         │  │ ├─ LXCReady: bool           │   │
         │  │ └─ Containers: map[string]  │   │
         │  │    ├─ Container "redroid"   │   │
         │  │    ├─ Container "android1"  │   │
         │  │    └─ Container "android2"  │   │
         ��  └─────────────────────────────┘   │
         └──────┬──────────────────────────────┘
                │
         ┌──────▼──────────────────────────────┐
         │    LXC System (Linux)               │
         │                                     │
         │  ┌─────────────────────────────┐   │
         │  │ LXC Bridge (lxcbr0)         │   │
         │  │ NAT Rules                   │   │
         │  │ IP Forwarding               │   │
         │  └─────────────────────────────┘   │
         │                                     │
         │  ┌─────────────────────────────┐   │
         │  │ Container: redroid          │   │
         │  │ ├─ /var/lib/lxc/redroid     │   │
         │  │ ├─ ~/data-redroid           │   │
         │  │ └─ redroid.log              │   │
         │  └───────────────��─────────────┘   │
         │                                     │
         │  ┌─────────────────────────────┐   │
         │  │ Container: android1         │   │
         │  │ ├─ /var/lib/lxc/android1    │   │
         │  │ ├─ ~/data-android1          │   │
         │  │ └─ android1.log             │   │
         │  └─────────────────────────────┘   │
         │                                     │
         │  ┌─────────────────────────────┐   │
         │  │ Container: android2         │   │
         │  │ ├─ /var/lib/lxc/android2    │   │
         │  │ ├─ ~/data-android2          │   │
         │  │ └─ android2.log             │   │
         │  └─────────────────────────────┘   │
         └─────────────────────────────────────┘
```

## Initialization Flow

### First Time Setup (LXC Preparation)

```
redway prepare-lxc
        │
        ▼
┌─���───────────────────────────────────┐
│ LXCPreparer.PrepareLXC()            │
├─────────────────────────────────────┤
│ 1. Check kernel modules (binder)    │
│ 2. Check LXC tools                  │
│ 3. Setup LXC networking (lxcbr0)    │
│ 4. Setup NAT rules                  │
│ 5. Enable IP forwarding             │
│ 6. Adjust OCI template              │
│ 7. Check required tools             │
└─────────────────────────────────────┘
        │
        ▼
   Config.LXCReady = true
   (saved to ~/.config/redway/config.json)
```

### Container Initialization

```
redway init [name] [image]
        │
        ▼
┌─────────────────────────────────────┐
│ Initializer.Initialize()            │
├─────────────────────────────────────┤
│ 1. Check if LXC is ready            │
│    └─ If not, run PrepareLXC()      │
│ 2. Create LXC container             │
│ 3. Fix container filesystem         │
│ 4. Create data directory            │
│ 5. Adjust container config          │
│ 6. Apply networking workaround      │
└─────────────────────────────────────┘
        │
        ▼
   Container.Initialized = true
   (saved to ~/.config/redway/config.json)
```

## Container Lifecycle

```
┌──────────────────────────────────────────────────────────┐
│                   Container Lifecycle                    │
└──────────────────────────────────────────────────────────┘

    redway init
        │
        ▼
    ┌─────────────┐
    │ CREATED     │
    │ (not init)  │
    └──────┬──────┘
           │
           │ (initialization steps)
           │
           ▼
    ┌─────────────┐
    │ INITIALIZED │
    │ (ready)     │
    └──────┬──────┘
           │
    ┌──────┴──────┐
    │             │
    │ redway      │ redway
    │ start       │ remove
    │             │
    ▼             ▼
┌─────────┐  ┌──────────┐
│ RUNNING │  │ REMOVED  │
└────┬────┘  └──────────┘
     │
     │ redway stop
     │
     ▼
┌─────────┐
│ STOPPED │
└────┬────┘
     │
     │ redway start
     │
     ▼
┌─────────┐
│ RUNNING │
└─────────┘
```

## Data Organization

```
~/.config/redway/
└── config.json
    {
      "lxc_ready": true,
      "containers": {
        "redroid": {
          "name": "redroid",
          "image_url": "docker://...",
          "data_path": "~/data-redroid",
          "log_file": "redroid.log",
          "gpu_mode": "guest",
          "initialized": true
        },
        "android1": {
          "name": "android1",
          "image_url": "docker://...",
          "data_path": "~/data-android1",
          "log_file": "android1.log",
          "gpu_mode": "guest",
          "initialized": true
        }
      }
    }

~/data-redroid/
└── (container data files)

~/data-android1/
└── (container data files)

/var/lib/lxc/
├── redroid/
│   ├── config
│   ├── rootfs/
│   └── ...
├── android1/
│   ├── config
│   ├── rootfs/
│   └── ...
└── ...
```

## Component Interaction

```
┌─────────────────────────────────────────────────────────┐
│                    Command Handler                      │
│                  (cmd/command.go)                       │
└────────────────────────┬────────────────────────────────┘
                         │
        ┌────────────────┼────────────────┐
        │                │                │
        ▼                ▼                ▼
   ┌─────────┐      ┌─────────┐      ┌─────────┐
   │LXCPrep  │      │Initializ│      │Manager  │
   │arer     │      │er       │      │         │
   └────┬────┘      └────┬────┘      └────┬────┘
        │                │                │
        └────────────────┼────────────────┘
                         │
                         ▼
                    ┌─────────────┐
                    │Config       │
                    │(persistent) │
                    └─────────────┘
                         │
                         ▼
                    ┌─────────────┐
                    │LXC System   │
                    │(Linux)      │
                    └─────────────┘
```

## Separation of Concerns

```
┌──────────────────────────────────────────────────────────┐
│                  LXC System Layer                        │
│  (One-time setup, shared by all containers)             │
│                                                          │
│  • Kernel modules                                        │
│  • LXC tools                                             │
│  • Network bridge (lxcbr0)                               │
│  • NAT rules                                             │
│  • IP forwarding                                         │
│  • OCI template                                          │
│  • Required tools (skopeo, umoci, jq)                    │
└──────────────────────────────────────────────────────────┘
                         ▲
                         │
                    (one-time)
                         │
                         │
┌──────────────────────────────────────────────────────────┐
│              Container Layer                            │
│  (Per-container setup, independent execution)           │
│                                                          │
│  Container 1:                                            │
│  • Create LXC container                                  │
│  • Fix filesystem                                        │
│  • Create data directory                                 │
│  • Adjust config                                         │
│  • Apply workarounds                                     │
│                                                          │
│  Container 2:                                            │
│  • Create LXC container                                  │
│  • Fix filesystem                                        │
│  • Create data directory                                 │
│  • Adjust config                                         │
│  • Apply workarounds                                     │
│                                                          │
│  Container N:                                            │
│  • ... (same as above)                                   │
└──────────────────────────────────────────────────────────┘
```

## Key Design Principles

1. **Separation of Concerns**
   - LXC system setup (one-time)
   - Container initialization (per-container)
   - Container management (per-container)

2. **Independence**
   - Each container has own data directory
   - Each container has own log file
   - Each container has own configuration
   - Containers can run simultaneously

3. **Efficiency**
   - LXC system setup only runs once
   - Subsequent containers skip system setup
   - Faster initialization for multiple containers

4. **Maintainability**
   - Clear code structure
   - Easy to extend
   - Easy to debug
   - Well-documented

5. **Backward Compatibility**
   - Default container name: "redroid"
   - Commands work without container name
   - Existing workflows unchanged
