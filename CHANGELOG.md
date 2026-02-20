# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.5.0] - 2026-02-19

### Added

- Compact `lines=4` mode for 128×32 displays
- 5×7 bitmap font for improved readability on small screens
- Example configs installed to `/usr/share/doc/i2c-display/configs/`

### Fixed

- Corrected systemd `MemoryLimit` directive in docs

## [0.4.0] - 2026-02-17

### Added

- Rolling load average graph page
- Color-coded system metric text (green/yellow/red thresholds)
- Custom bitmap icons for system metrics (CPU, memory, disk, temperature)

## [0.3.1] - 2026-02-17

### Added

- Man page (`i2c-display.1`)
- PIE build mode (`-buildmode=pie`) by default
- `%check` section in RPM spec

### Fixed

- ARMv7 cross-compilation failures
- golangci-lint CI configuration
- Systemd unit installed to correct path (`/usr/lib/systemd/system`)
- Release tarball now includes all docs and packaging directories
- Removed sudo from Makefile install targets
- Multiple code quality and bad practice fixes

## [0.3.0] - 2026-02-17

### Added

- ST7735 TFT colour display support via SPI
- UCTRONICS Pi Rack Pro display type (`st7735_160x80_uctronics`)
- `-test-display` flag for hardware verification
- Example config for ST7735 160×80 (0.96" Waveshare)
- Vendored Go dependencies for offline/Fedora builds

### Fixed

- Display flicker eliminated by not flushing on `Clear()`
- Hostname rendered in green on colour displays
- Replaced SPI-based UCTRONICS driver with correct I2C bridge protocol
- Corrected MADCTL for 160×80 landscape panel orientation
- Removed INVON from ST7735 init sequence
- SPI writes chunked to respect sysfs 4096-byte tx limit
- Race conditions resolved with added concurrency safety
- Code quality issues (critical, high, and medium priority)
- Test coverage improved from 63.5% to 75.1%

### Changed

- CI and release workflows updated to Go 1.24
- Removed RPM/DEB builds from release workflow (build locally instead)
- Updated dependency versions (golang.org/x/image, periph.io/x/devices)

## [0.2.0] - 2026-02-12

### Added

- Improved small display layout

### Fixed

- All golangci-lint warnings resolved
- Go version compatibility for Go 1.21–1.23 (downgraded golang.org/x/image)
- golangci-lint configuration updated for v1/v2 compatibility

## [0.1.0] - 2026-02-12

### Added

- Adaptive layout system for different display sizes
- Comprehensive configuration guide and multi-platform documentation
- Fedora/RHEL/CentOS install instructions for i2c-tools

### Fixed

- Systemd unit updated for package installation and flexible I2C bus support

## [0.0.3] - 2026-02-11

### Fixed

- Display dimension auto-detection
- DEB package artifact paths
- Disabled dwz compression for Go binaries
- Removed deprecated `--with=systemd` from debian/rules
- RPM systemd unit directory macro definition
- Git tag detection in release workflow

## [0.0.2] - 2026-02-09

### Added

- Multi-display support with factory pattern
- Automatic display dimension detection
- Framework for multiple display types
- RPM and DEB package build support
- Build system, documentation, and deployment files
- License compatibility documentation
- CI workflows (golangci-lint, CodeQL, release)

### Changed

- Project renamed from ssd1306-display to i2c-display

### Fixed

- Packaging artifact paths in release workflow
- Maintainer name and email in packaging files

## 0.0.1 - 2026-02-09

### Added

- Initial implementation of SSD1306 display controller

[0.5.0]: https://github.com/ausil/i2c-display/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/ausil/i2c-display/compare/v0.3.1...v0.4.0
[0.3.1]: https://github.com/ausil/i2c-display/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/ausil/i2c-display/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/ausil/i2c-display/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/ausil/i2c-display/compare/v0.0.3...v0.1.0
[0.0.3]: https://github.com/ausil/i2c-display/compare/v0.0.2...v0.0.3
[0.0.2]: https://github.com/ausil/i2c-display/compare/v0.0.1...v0.0.2
