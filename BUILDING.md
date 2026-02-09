# Building and Packaging Guide

This document describes how to build release packages for the SSD1306 Display Controller.

## Version

Current version: **0.0.1**

Version is managed in the `VERSION` file at the root of the repository.

## Prerequisites

### For RPM packages (Fedora, RHEL, CentOS, etc.)
```bash
sudo dnf install rpm-build rpmdevtools golang
```

### For DEB packages (Debian, Ubuntu, Raspberry Pi OS, etc.)
```bash
sudo apt-get install build-essential debhelper dh-golang golang-go
```

## Building Release Tarball

Create a source tarball suitable for distribution:

```bash
make dist
```

This creates `dist/ssd1306-display-0.0.1.tar.gz` containing all source files.

## Building RPM Packages

### Build both source and binary RPMs:
```bash
make rpm
```

This creates:
- `rpm-build/RPMS/*/ssd1306-display-0.0.1-1.*.rpm` (binary RPM)
- `rpm-build/SRPMS/ssd1306-display-0.0.1-1.src.rpm` (source RPM)

### Build only source RPM:
```bash
make srpm
```

### Install the RPM:
```bash
make install-rpm
# Or manually:
sudo rpm -Uvh rpm-build/RPMS/*/ssd1306-display-*.rpm
```

## Building DEB Packages

### Build binary package:
```bash
make deb
```

This creates `../ssd1306-display_0.0.1-1_*.deb`

### Build source package:
```bash
make deb-src
```

This creates:
- `../ssd1306-display_0.0.1-1.dsc`
- `../ssd1306-display_0.0.1-1.tar.xz`

### Install the DEB:
```bash
make install-deb
# Or manually:
sudo dpkg -i ../ssd1306-display_*.deb
sudo apt-get install -f  # Install dependencies if needed
```

## Cross-Compilation

Build binaries for different ARM architectures:

```bash
# For Raspberry Pi 2/3 (ARMv7)
make build-arm7

# For Raspberry Pi 4, Rock 3C (ARM64)
make build-arm64

# Build all architectures
make build-all
```

Binaries will be in:
- `bin/ssd1306d` (native architecture)
- `bin/ssd1306d-arm7` (ARMv7 32-bit)
- `bin/ssd1306d-arm64` (ARM64)

## Package Contents

All packages include:

- Binary: `/usr/bin/ssd1306d`
- Config: `/etc/ssd1306-display/config.json`
- Service: `/lib/systemd/system/ssd1306-display.service`
- Docs: `/usr/share/doc/ssd1306-display/`

## Testing Packages

### Test RPM installation:
```bash
# Build and install
make rpm
sudo rpm -Uvh rpm-build/RPMS/*/ssd1306-display-*.rpm

# Test the service
sudo systemctl start ssd1306-display.service
sudo systemctl status ssd1306-display.service
journalctl -u ssd1306-display.service -f

# Uninstall
sudo rpm -e ssd1306-display
```

### Test DEB installation:
```bash
# Build and install
make deb
sudo dpkg -i ../ssd1306-display_*.deb
sudo apt-get install -f

# Test the service
sudo systemctl start ssd1306-display.service
sudo systemctl status ssd1306-display.service
journalctl -u ssd1306-display.service -f

# Uninstall
sudo apt-get remove ssd1306-display
```

## Releasing a New Version

1. Update the `VERSION` file:
   ```bash
   echo "0.0.2" > VERSION
   ```

2. Update changelogs:
   - `rpm/ssd1306-display.spec` - Add entry in `%changelog` section
   - `debian/changelog` - Use `dch -v 0.0.2-1` or edit manually

3. Commit the version bump:
   ```bash
   git add VERSION rpm/ssd1306-display.spec debian/changelog
   git commit -m "Bump version to 0.0.2"
   git tag -a v0.0.2 -m "Release v0.0.2"
   ```

4. Build packages:
   ```bash
   make dist
   make rpm
   make deb
   ```

5. Test packages on target systems

6. Push to repository:
   ```bash
   git push origin main
   git push origin v0.0.2
   ```

## Cleaning Build Artifacts

Remove all build artifacts:

```bash
make clean
```

This removes:
- `bin/` - Built binaries
- `dist/` - Release tarballs
- `rpm-build/` - RPM build directory
- `debian/` build files
- DEB packages in parent directory

## Troubleshooting

### RPM build fails with "go: command not found"
Install Go development tools:
```bash
sudo dnf install golang
```

### DEB build fails with missing dependencies
Install build dependencies:
```bash
sudo apt-get install build-essential debhelper dh-golang golang-go
```

### Permission denied when accessing /dev/i2c-*
The service runs as root. For manual testing:
```bash
sudo usermod -a -G i2c $USER
# Log out and back in
```

### Service fails to start
Check logs:
```bash
sudo journalctl -u ssd1306-display.service -xe
```

Common issues:
- I2C not enabled (see README.md for enabling I2C)
- Display not connected or wrong address
- Configuration file syntax errors

## Directory Structure

```
.
├── VERSION                 # Version file
├── rpm/
│   └── ssd1306-display.spec   # RPM spec file
├── debian/
│   ├── control            # Package metadata
│   ├── changelog          # Debian changelog
│   ├── rules              # Build rules
│   ├── copyright          # License info
│   └── ...                # Other Debian files
├── Makefile               # Build system
└── ...
```
