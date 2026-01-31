# Reddock - Redroid Docker Manager

Reddock is a lightweight Go-based **Docker container manager** for
[redroid](https://github.com/remote-android/redroid-doc). It provides a
streamlined CLI for managing Android containers using Docker, focusing on ease
of use, performance, and reliability.

## Key Features

- **Docker-Based** - Leverages Docker for robust container management
- **Simplified CLI** - easy-to-use commands for init, start, stop, and removal
- **Addons System (Injection-Based)** - Inject ARM translation and GAPPS into
  running containers
- **Kernel Module Management** - Automatically checks and attempts to load
  required kernel modules (`binder_linux`)
- **ADB Integration** - Built-in ADB connection management with automatic port
  mapping
- **GPU Acceleration Support** - Easy configuration for GPU rendering modes
  (`host`, `guest`, `auto`)
- **Persistent Storage** - Android data persists across container restarts using
  Docker volumes

## What is Redroid?

[Redroid](https://github.com/remote-android/redroid-doc) (Remote Android) is a
GPU-accelerated **AIC (Android In Container)** solution that allows you to run
Android in containers with near-native performance. It is commonly used for
cloud gaming, automated testing, and virtual mobile infrastructure.

## Installation

1. **Install Dependencies**

   You need to install go/golang to build, and then lzip, tar, and xz-utils on
   your Linux systems.

2. **Clone and Build**

   ```bash
   git clone https://github.com/jimed-rand/reddock.git
   cd reddock
   make build
   ```

3. **Install (Optional)**

   ```bash
   sudo make install
   ```

## Usage

> [!IMPORTANT]
> Most commands require `sudo` to manage container runtimes and kernel modules.

### 1. Initialize a Container

```bash
sudo reddock init my-android redroid/redroid:13.0.0-latest
```

### 2. Start the Container

```bash
sudo reddock start my-android
```

### 3. Connect via ADB

```bash
reddock adb-connect my-android
```

## Commands

| Command                 | Description                                         |
| ----------------------- | --------------------------------------------------- |
| `init <name> [image]`   | Initialize a new Redroid container                  |
| `start <name> [-v]`     | Start a container (use -v for logs)                 |
| `stop <name>`           | Stop a running container                            |
| `restart <name> [-v]`   | Restart a container                                 |
| `status <name>`         | Show container status and info                      |
| `shell <name>`          | Enter the container shell                           |
| `adb-connect <name>`    | Connect to the container via ADB                    |
| `addons list`           | List available addons (Gapps, ARM Translation)      |
| `addons inject <n> <a>` | Inject single addon to running container            |
| `addons inject-multi`   | Inject multiple addons at once                      |
| `addons register-gapps` | Fetch Android ID for Google Play registration       |
| `log <name>`            | Show container logs                                 |
| `list`                  | List all Reddock-managed containers                 |
| `remove <name> [--all]` | Remove a container, data, and optionally image (-a) |

## Addons System (Injection-Based)

Reddock features a unique **injection-based** addon system. Unlike traditional
methods that require building custom Docker images, Reddock can inject ARM
translation layers and Google Apps directly into a **running container**.

### Available Addons

#### 1. ARM Translation

- **Houdini**: Intel Houdini (Intel's ARM translation libraries on x86/x86_64).
  - Support: Android 8.1.0 - 16.0.0
  - Arch: x86, x86_64
- **NDK Translation**: Google NDK Translation for ARM compatibility.
  - Support: Android 8.1.0 - 16.0.0 (including 64-only)
  - Arch: x86, x86_64

#### 2. Google Apps (GAPPS)

- **LiteGapps**: Optimized lightweight GAPPS. (Android 8.1 - 16)
- **MindTheGapps**: Minimal GAPPS package. (Android 12 - 16)
- **OpenGapps**: Pico version. (Android 11 only)

### Injection Workflow

1. **Start** your container first: `sudo reddock start my-android`
2. **Inject** the addon: `sudo reddock addons inject my-android litegapps`
3. **Restart** to apply changes: `sudo reddock restart my-android`

### Google Play Certification

According to
[Waydroid documentation](https://docs.waydro.id/faq/google-play-certification),
you need to register your session before using Google Play.

1. **Start** your container and **Open** the Google Play Store once.
2. **Fetch** your Android ID:
   ```bash
   sudo reddock addons register-gapps my-android
   ```
3. **Submit** the ID to
   [Google's device registration page](https://www.google.com/android/uncertified/).

### Advanced Injection Examples

**Gaming Setup (Android 13 + GAPPS + ARM Support):**

```bash
sudo reddock init gaming redroid/redroid:13.0.0-latest
sudo reddock start gaming
sudo reddock addons inject-multi gaming litegapps ndk
sudo reddock restart gaming
```

**Development (Android 11 + OpenGapps + Houdini):**

```bash
sudo reddock init dev redroid/redroid:11.0.0-latest
sudo reddock start dev
sudo reddock addons inject-multi dev opengapps houdini
sudo reddock restart dev
```

## Troubleshooting

- **Container must be running**: Injection only works on active containers.
- **Permission Issues**: If apps crash after injection, try fixing permissions:
  `sudo reddock shell <name>` then `chmod -R 755 /system/priv-app`
- **Architecture Mismatch**: ARM translation addons (Houdini/NDK) only work on
  x86_64 hosts.

## Implementation Details

The injection system works by:

1. Downloading the requested addon to `~/.cache/reddock/downloads`.
2. Extracting files to a temporary workspace.
3. Using `docker cp` or `podman cp` to push files into the root filesystem of
   the running container.
4. Setting correct Linux permissions via `exec chmod`.
5. Cleaning up temporary files.

## Credits

- [remote-android/redroid-script](https://github.com/remote-android/redroid-script) -
  Inspiration for the addon logic.
- [redroid](https://github.com/remote-android/redroid-doc) - Remote-Android
  project.
- [LiteGapps](https://litegapps.github.io/) /
  [MindTheGapps](https://gitlab.com/MindTheGapps) /
  [OpenGapps](https://opengapps.org/).

## License

GPL-2.0 license
