Name:           ssd1306-display
Version:        0.0.1
Release:        1%{?dist}
Summary:        SSD1306 OLED display controller for single board computers

License:        BSD
URL:            https://github.com/denniskorablev/ssd1306-display
Source0:        %{name}-%{version}.tar.gz

BuildRequires:  golang >= 1.19
BuildRequires:  systemd-rpm-macros
Requires:       systemd

%description
A Go application for Single Board Computers (Raspberry Pi 3/4, Rock 3C) that
controls an SSD1306 128x64 OLED display via I2C, showing system stats and
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
install -d %{buildroot}%{_sysconfdir}/ssd1306-display
install -d %{buildroot}%{_unitdir}

# Install binary
install -m 0755 bin/ssd1306d %{buildroot}%{_bindir}/ssd1306d

# Install config
install -m 0644 configs/config.example.json %{buildroot}%{_sysconfdir}/ssd1306-display/config.json

# Install systemd service
install -m 0644 systemd/ssd1306-display.service %{buildroot}%{_unitdir}/ssd1306-display.service

%post
%systemd_post ssd1306-display.service

%preun
%systemd_preun ssd1306-display.service

%postun
%systemd_postun_with_restart ssd1306-display.service

%files
%license LICENSE
%doc README.md LICENSES.md
%{_bindir}/ssd1306d
%config(noreplace) %{_sysconfdir}/ssd1306-display/config.json
%{_unitdir}/ssd1306-display.service

%changelog
* Sun Feb 09 2025 Dennis Korablev <dennis@example.com> - 0.0.1-1
- Initial release
- System monitoring with disk, RAM, and CPU temperature display
- Network interface information display
- Automatic page rotation
- Systemd service integration
- JSON-based configuration
- Mock display support for testing
- Comprehensive test coverage
