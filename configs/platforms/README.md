# Platform-Specific Configurations

This directory contains tested configuration files for specific single board computers (SBCs).

## Available Platforms

### Raspberry Pi

#### Raspberry Pi Zero / Zero W
- **Config**: `raspberrypi-zero.json`
- **I2C Bus**: `/dev/i2c-1`
- **Notes**:
  - Limited to WiFi (wlan0) unless using USB Ethernet
  - Slower refresh rate (2s) to reduce CPU load
  - Screen saver enabled to save power
  - Only 2 interfaces per page due to limited resources

#### Raspberry Pi 3 / 3B+
- **Config**: Use `raspberrypi-4.json` (same hardware setup)
- **I2C Bus**: `/dev/i2c-1`
- **Notes**:
  - Both Ethernet and WiFi available
  - Standard refresh rate
  - Metrics enabled

#### Raspberry Pi 4 / 4B
- **Config**: `raspberrypi-4.json`
- **I2C Bus**: `/dev/i2c-1`
- **Notes**:
  - Best performance of the Pi line
  - Metrics enabled by default
  - All features work well

### Radxa Rock Series

#### Rock 3C
- **Config**: `rock-3c.json`
- **I2C Bus**: `/dev/i2c-3` (verify with `i2cdetect -l`)
- **Notes**:
  - May have multiple I2C buses, check with `ls /dev/i2c-*`
  - Interface naming may use `end0` instead of `eth0`
  - Good performance, metrics enabled

#### Rock 5B
- **Config**: Use `rock-3c.json` and verify I2C bus number
- **I2C Bus**: `/dev/i2c-X` (check your board)
- **Notes**:
  - Very powerful, all features work excellently
  - May need to adjust I2C bus in config

## How to Use

1. **Copy the appropriate config**:
   ```bash
   sudo cp configs/platforms/YOUR-PLATFORM.json /etc/i2c-display/config.json
   ```

2. **Verify I2C bus** (if needed):
   ```bash
   # List available I2C buses
   ls /dev/i2c-*

   # Scan for devices (replace X with bus number)
   sudo i2cdetect -y X
   ```

3. **Adjust for your setup**:
   - Check network interface names: `ip link show`
   - Verify temperature sensor path: `cat /sys/class/thermal/thermal_zone0/temp`
   - Test config: `i2c-displayd -validate-config -config /etc/i2c-display/config.json`

4. **Start the service**:
   ```bash
   sudo systemctl restart i2c-display.service
   ```

## Platform-Specific Notes

### Raspberry Pi
- **Enable I2C**: Use `raspi-config` → Interface Options → I2C
- **User Permissions**: Add user to `i2c` group: `sudo usermod -a -G i2c $USER`
- **I2C Bus**: Always `/dev/i2c-1` on modern Pi models

### Rock 3C / 5B
- **I2C Bus**: May vary, check with `i2cdetect -l`
- **Interface Names**: May use `end0` instead of `eth0`
- **Temperature**: Path may be `/sys/devices/virtual/thermal/thermal_zone0/temp`

## Troubleshooting

### Display Not Found

```bash
# Check I2C devices
sudo i2cdetect -y 1  # or appropriate bus number

# Should show 0x3C or 0x3D
```

### Wrong Interface Names

```bash
# List interfaces
ip link show

# Update config.json with actual names
```

### Temperature Not Working

```bash
# Find temperature sensor
find /sys -name "temp" 2>/dev/null | grep thermal

# Update temperature_source in config
```

### Performance Issues

- Increase `refresh_interval` (e.g., "2s" instead of "1s")
- Reduce `max_interfaces_per_page`
- Disable metrics if not needed
- Enable screen saver

## Contributing

If you've tested on a platform not listed here:

1. Create a config file: `configs/platforms/YOUR-PLATFORM.json`
2. Add notes to this README
3. Test thoroughly
4. Submit a pull request

Include:
- Platform name and model
- I2C bus location
- Any platform-specific quirks
- Recommended settings

## Tested Platforms

| Platform | Tested | Config Available | Notes |
|----------|--------|-----------------|-------|
| Raspberry Pi Zero W | ✅ | Yes | Works well, limited resources |
| Raspberry Pi 3B+ | ✅ | Yes | Excellent |
| Raspberry Pi 4 | ✅ | Yes | Excellent, all features |
| Rock 3C | ✅ | Yes | Great performance |
| Rock 5B | ⚠️ | Use Rock 3C config | Verify I2C bus |
| Orange Pi | ❌ | No | Contributions welcome |
| NanoPi | ❌ | No | Contributions welcome |

Legend:
- ✅ Tested and working
- ⚠️ Should work, needs verification
- ❌ Not tested yet
