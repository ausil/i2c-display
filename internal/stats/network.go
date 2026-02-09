package stats

import (
	"fmt"
	"net"
	"path/filepath"

	"github.com/denniskorablev/ssd1306-display/internal/config"
)

// NetworkCollector collects network interface information
type NetworkCollector struct {
	config config.NetworkConfig
}

// NewNetworkCollector creates a new network collector
func NewNetworkCollector(cfg config.NetworkConfig) *NetworkCollector {
	return &NetworkCollector{
		config: cfg,
	}
}

// GetInterfaces returns filtered network interfaces with their addresses
func (n *NetworkCollector) GetInterfaces() ([]NetInterface, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get network interfaces: %w", err)
	}

	var result []NetInterface

	for _, iface := range ifaces {
		// Skip down interfaces
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		// Apply filters
		if !n.shouldInclude(iface.Name) {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		netIface := NetInterface{
			Name: iface.Name,
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			ip := ipNet.IP
			if ip == nil {
				continue
			}

			// Check if it's IPv4 or IPv6
			if ip.To4() != nil {
				if n.config.ShowIPv4 {
					netIface.IPv4Addrs = append(netIface.IPv4Addrs, ip.String())
				}
			} else {
				if n.config.ShowIPv6 {
					// Filter out link-local IPv6 addresses
					if !ip.IsLinkLocalUnicast() {
						netIface.IPv6Addrs = append(netIface.IPv6Addrs, ip.String())
					}
				}
			}
		}

		// Only add interface if it has addresses we care about
		if len(netIface.IPv4Addrs) > 0 || len(netIface.IPv6Addrs) > 0 {
			result = append(result, netIface)
		}
	}

	return result, nil
}

// shouldInclude checks if an interface should be included based on filters
func (n *NetworkCollector) shouldInclude(name string) bool {
	// First check exclude patterns
	for _, pattern := range n.config.InterfaceFilter.Exclude {
		matched, _ := filepath.Match(pattern, name)
		if matched {
			return false
		}
	}

	// If auto_detect is true and no include patterns, include all
	if n.config.AutoDetect && len(n.config.InterfaceFilter.Include) == 0 {
		return true
	}

	// Check include patterns
	for _, pattern := range n.config.InterfaceFilter.Include {
		matched, _ := filepath.Match(pattern, name)
		if matched {
			return true
		}
	}

	return false
}
