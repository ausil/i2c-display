# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 0.0.2   | :white_check_mark: |
| < 0.0.2 | :x:                |

## Security Considerations

### Hardware Access

This application requires I2C hardware access, which typically requires elevated privileges:

- The application must run as root or have appropriate I2C permissions
- Systemd service runs with restricted capabilities where possible
- See systemd service file for security hardening options

### Configuration Security

- Configuration files may contain sensitive information
- Set appropriate file permissions: `chmod 600 /etc/i2c-display/config.json`
- The application validates all configuration inputs
- Environment variables can override sensitive settings

### Network Exposure

If metrics are enabled:
- The metrics endpoint exposes system information
- Bind to localhost only in production: `"address": "127.0.0.1:9090"`
- Use firewall rules to restrict access
- Consider authentication proxy if exposing externally

### Dependency Management

- All dependencies are tracked in `go.mod` and `go.sum`
- Dependabot monitors for security updates
- Regular security scans via CodeQL and gosec in CI/CD

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

Instead, please report them via:

1. **GitHub Security Advisories** (preferred)
   - Go to https://github.com/ausil/i2c-display/security/advisories
   - Click "Report a vulnerability"

2. **Direct Email**
   - Contact the maintainer directly
   - Include detailed information about the vulnerability

### What to Include

- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)
- Your contact information

### Response Timeline

- **Initial Response**: Within 48 hours
- **Status Update**: Within 7 days
- **Fix Timeline**: Depends on severity
  - Critical: Within 7 days
  - High: Within 14 days
  - Medium: Within 30 days
  - Low: Next release cycle

### Disclosure Policy

- We follow coordinated disclosure
- We will credit reporters (unless anonymity is requested)
- Security advisories will be published after fixes are released
- CVEs will be requested for significant vulnerabilities

## Security Best Practices

### For Users

1. **Keep Updated**
   ```bash
   # Check for updates regularly
   sudo systemctl stop i2c-display
   # Update package
   sudo systemctl start i2c-display
   ```

2. **Secure Configuration**
   ```bash
   # Set correct permissions
   sudo chmod 600 /etc/i2c-display/config.json
   sudo chown root:root /etc/i2c-display/config.json
   ```

3. **Monitor Logs**
   ```bash
   # Check for suspicious activity
   sudo journalctl -u i2c-display -f
   ```

4. **Restrict Metrics Access**
   ```json
   {
     "metrics": {
       "enabled": true,
       "address": "127.0.0.1:9090"  // Localhost only
     }
   }
   ```

### For Developers

1. **Input Validation**
   - All configuration inputs are validated
   - Bounds checking on all array accesses
   - Safe type conversions

2. **Error Handling**
   - Never expose sensitive information in errors
   - Log errors with appropriate context
   - Fail securely

3. **Dependencies**
   - Review dependency changes
   - Keep dependencies minimal
   - Monitor security advisories

4. **Code Review**
   - All changes require review
   - Security-sensitive changes require thorough review
   - Automated security scanning in CI

## Systemd Security Hardening

The provided systemd service includes security hardening:

```ini
[Service]
# Restrict privileges
NoNewPrivileges=true

# Filesystem restrictions
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/dev/i2c-1
PrivateTmp=true

# Additional hardening (optional, may need adjustment)
# ProtectKernelTunables=true
# ProtectKernelModules=true
# ProtectControlGroups=true
```

Adjust based on your security requirements and system capabilities.

## Known Limitations

1. **I2C Access Required** - Application needs I2C device access
2. **Root or i2c Group** - Must run with appropriate permissions
3. **Local System Only** - Designed for local SBC use, not network exposure
4. **Hardware Dependencies** - Behavior depends on I2C hardware availability

## Security Checklist

- [ ] Application runs with minimum required privileges
- [ ] Configuration file permissions are restrictive (600)
- [ ] Metrics endpoint (if enabled) is localhost-only or firewalled
- [ ] Logs are monitored for anomalies
- [ ] System is kept up to date
- [ ] Only trusted displays are connected (I2C security)

## Additional Resources

- [OWASP Top Ten](https://owasp.org/www-project-top-ten/)
- [CWE/SANS Top 25](https://www.sans.org/top25-software-errors/)
- [Go Security Policy](https://go.dev/security/policy)

Thank you for helping keep this project secure!
