Name:           i2c-display
Version:        0.4.0
Release:        1%{?dist}
Summary:        I2C OLED display controller for single board computers

License:        BSD-3-Clause AND Apache-2.0 AND MIT AND BSD-2-Clause
URL:            https://github.com/ausil/i2c-display
Source0:        %{name}-%{version}.tar.gz

BuildRequires:  golang >= 1.24
BuildRequires:  systemd-rpm-macros
Requires:       systemd

# Bundled Go dependencies (vendored)
Provides:       bundled(golang(github.com/beorn7/perks)) = 1.0.1
Provides:       bundled(golang(github.com/cespare/xxhash/v2)) = 2.3.0
Provides:       bundled(golang(github.com/mattn/go-colorable)) = 0.1.13
Provides:       bundled(golang(github.com/mattn/go-isatty)) = 0.0.20
Provides:       bundled(golang(github.com/munnerz/goautoneg)) = 0.0.0
Provides:       bundled(golang(github.com/prometheus/client_golang)) = 1.23.2
Provides:       bundled(golang(github.com/prometheus/client_model)) = 0.6.2
Provides:       bundled(golang(github.com/prometheus/common)) = 0.66.1
Provides:       bundled(golang(github.com/prometheus/procfs)) = 0.16.1
Provides:       bundled(golang(github.com/rs/zerolog)) = 1.34.0
Provides:       bundled(golang(go.yaml.in/yaml/v2)) = 2.4.2
Provides:       bundled(golang(golang.org/x/image)) = 0.36.0
Provides:       bundled(golang(golang.org/x/sys)) = 0.35.0
Provides:       bundled(golang(google.golang.org/protobuf)) = 1.36.8
Provides:       bundled(golang(periph.io/x/conn/v3)) = 3.7.2
Provides:       bundled(golang(periph.io/x/devices/v3)) = 3.7.4
Provides:       bundled(golang(periph.io/x/host/v3)) = 3.8.5

ExclusiveArch:  %{go_arches}

%description
A Go application for Single Board Computers (Raspberry Pi 3/4, Rock 3C) that
controls I2C OLED displays (SSD1306, SH1106, etc.) showing system stats and
network information with rotating pages.

Features:
- System monitoring (disk, RAM, CPU temperature)
- Network interface information
- Automatic page rotation
- Systemd integration
- JSON configuration

%prep
%setup -q

%build
# Build with vendored dependencies (no network access required)
make build GOBUILD="go build -mod=vendor -buildmode=pie"

%check
make test GOTEST="go test -mod=vendor"

%install
# Create directories
install -d %{buildroot}%{_bindir}
install -d %{buildroot}%{_sysconfdir}/i2c-display
install -d %{buildroot}%{_unitdir}
install -d %{buildroot}%{_mandir}/man1
install -d %{buildroot}%{_docdir}/i2c-display/configs/platforms

# Install binary
install -m 0755 bin/i2c-displayd %{buildroot}%{_bindir}/i2c-displayd

# Install config
install -m 0644 configs/config.example.json %{buildroot}%{_sysconfdir}/i2c-display/config.json

# Install example configs to docdir
install -m 0644 configs/*.json %{buildroot}%{_docdir}/i2c-display/configs/
install -m 0644 configs/platforms/*.json %{buildroot}%{_docdir}/i2c-display/configs/platforms/
install -m 0644 configs/platforms/README.md %{buildroot}%{_docdir}/i2c-display/configs/platforms/

# Install systemd service
install -m 0644 systemd/i2c-display.service %{buildroot}%{_unitdir}/i2c-display.service

# Install man page
install -m 0644 man/i2c-displayd.1 %{buildroot}%{_mandir}/man1/i2c-displayd.1

%post
%systemd_post i2c-display.service

%preun
%systemd_preun i2c-display.service

%postun
%systemd_postun_with_restart i2c-display.service

%files
%license LICENSE
%license vendor/github.com/beorn7/perks/LICENSE
%license vendor/github.com/cespare/xxhash/v2/LICENSE.txt
%license vendor/github.com/mattn/go-colorable/LICENSE
%license vendor/github.com/mattn/go-isatty/LICENSE
%license vendor/github.com/munnerz/goautoneg/LICENSE
%license vendor/github.com/prometheus/client_golang/LICENSE
%license vendor/github.com/prometheus/client_model/LICENSE
%license vendor/github.com/prometheus/common/LICENSE
%license vendor/github.com/prometheus/procfs/LICENSE
%license vendor/github.com/rs/zerolog/LICENSE
%license vendor/go.yaml.in/yaml/v2/LICENSE
%license vendor/golang.org/x/image/LICENSE
%license vendor/golang.org/x/sys/LICENSE
%license vendor/google.golang.org/protobuf/LICENSE
%license vendor/periph.io/x/conn/v3/LICENSE
%license vendor/periph.io/x/devices/v3/LICENSE
%license vendor/periph.io/x/host/v3/LICENSE
%doc README.md LICENSES.md
%doc %{_docdir}/i2c-display/configs/
%{_bindir}/i2c-displayd
%{_mandir}/man1/i2c-displayd.1*
%config(noreplace) %{_sysconfdir}/i2c-display/config.json
%{_unitdir}/i2c-display.service

%changelog
* Mon Feb 17 2025 Dennis Gilmore <dennis@ausil.us> - 0.4.0-1
- Add rolling load average graph page with color-coded bars
- Read /proc/loadavg for 1m, 5m, 15m load averages
- Color thresholds based on per-core load (green/yellow/red)
- Text-only fallback for small displays (128x32)

* Mon Feb 17 2025 Dennis Gilmore <dennis@ausil.us> - 0.3.1-1
- Pin gosec CI action to v2.23.0 and enforce security scan failures
- Default metrics endpoint to 127.0.0.1:9090 to prevent network exposure
- Align Makefile install path with systemd service ExecStart
- Log ST7735 SPI port cleanup errors instead of suppressing
- Add systemd security hardening (ProtectKernelLogs, RestrictRealtime, etc.)
- Remove unnecessary network.target dependency from systemd service
- Remove sudo from Makefile install targets
- Vendor Go dependencies for offline/Fedora builds
- Fix display flicker by not flushing framebuffer on Clear()
- Install systemd unit to /usr/lib/systemd/system instead of /etc/systemd/system

* Sun Feb 16 2025 Dennis Gilmore <dennis@ausil.us> - 0.3.0-1
- Add UCTRONICS colour display support (uctronics_colour) via I2C bridge MCU
- Render hostname in green on colour displays
- Preserve source colours in DrawImage on colour displays (ST7735, UCTRONICS)
- Fix display flicker by not flushing framebuffer on Clear()
- Use max-channel brightness for monochrome DrawImage so saturated colours render as white

* Wed Feb 12 2025 Dennis Gilmore <dennis@ausil.us> - 0.2.0-1
- Enhanced small display support for 128x32 screens
- Split system metrics into individual rotating pages
- Show actual GB usage instead of percentages for disk and memory
- Reduce page rotation interval to 2 seconds for faster updates
- Add separator line after header on all display sizes
- Fix text rendering to ensure hostname always visible
- Reduce margins to 1px for maximum text space
- Downgrade golang.org/x/image for Go 1.21-1.23 compatibility
- Resolve all golangci-lint warnings and code quality issues
- Fix CI configuration and build process

* Wed Feb 12 2025 Dennis Gilmore <dennis@ausil.us> - 0.1.0-1
- Adaptive layout system for different display sizes
- Fix text rendering to prevent content from overwriting header
- Optimized layout for 128x32 displays with compact formatting
- Hostname header now always visible on small displays
- Fixed systemd unit file to use /usr/bin path
- Improved documentation for multi-platform SBC support

* Wed Feb 12 2025 Dennis Gilmore <dennis@ausil.us> - 0.0.3-1
- Fix display dimension auto-detection for different display types
- Display dimensions now automatically set based on display type
- Fixes configuration validation error for ssd1306_128x32 displays

* Sun Feb 09 2025 Dennis Gilmore <dennis@ausil.us> - 0.0.2-1
- Production-ready enhancements
- Added structured logging with zerolog
- Added Prometheus metrics endpoint
- Added configuration validation and hot reload
- Added screen saver with brightness control
- Added retry logic and health monitoring
- Complete CI/CD pipeline with multi-arch builds
- Platform-specific configurations
- Security hardening and code quality improvements

* Sun Feb 09 2025 Dennis Gilmore <dennis@ausil.us> - 0.0.1-1
- Initial release
- System monitoring with disk, RAM, and CPU temperature display
- Network interface information display
- Automatic page rotation
- Systemd service integration
- JSON-based configuration
- Mock display support for testing
- Comprehensive test coverage
