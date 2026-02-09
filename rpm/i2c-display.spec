Name:           i2c-display
Version:        0.0.2
Release:        1%{?dist}
Summary:        I2C OLED display controller for single board computers

License:        BSD
URL:            https://github.com/ausil/i2c-display
Source0:        %{name}-%{version}.tar.gz

BuildRequires:  golang >= 1.19
BuildRequires:  systemd-rpm-macros
Requires:       systemd

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
# Build the binary
make build

%install
# Create directories
install -d %{buildroot}%{_bindir}
install -d %{buildroot}%{_sysconfdir}/i2c-display
install -d %{buildroot}%{_unitdir}

# Install binary
install -m 0755 bin/i2c-displayd %{buildroot}%{_bindir}/i2c-displayd

# Install config
install -m 0644 configs/config.example.json %{buildroot}%{_sysconfdir}/i2c-display/config.json

# Install systemd service
install -m 0644 systemd/i2c-display.service %{buildroot}%{_unitdir}/i2c-display.service

%post
%systemd_post i2c-display.service

%preun
%systemd_preun i2c-display.service

%postun
%systemd_postun_with_restart i2c-display.service

%files
%license LICENSE
%doc README.md LICENSES.md
%{_bindir}/i2c-displayd
%config(noreplace) %{_sysconfdir}/i2c-display/config.json
%{_unitdir}/i2c-display.service

%changelog
* Sun Feb 09 2025 Dennis Korablev <dennis@example.com> - 0.0.2-1
- Production-ready enhancements
- Added structured logging with zerolog
- Added Prometheus metrics endpoint
- Added configuration validation and hot reload
- Added screen saver with brightness control
- Added retry logic and health monitoring
- Complete CI/CD pipeline with multi-arch builds
- Platform-specific configurations
- Security hardening and code quality improvements

* Sun Feb 09 2025 Dennis Korablev <dennis@example.com> - 0.0.1-1
- Initial release
- System monitoring with disk, RAM, and CPU temperature display
- Network interface information display
- Automatic page rotation
- Systemd service integration
- JSON-based configuration
- Mock display support for testing
- Comprehensive test coverage
