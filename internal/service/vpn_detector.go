package service

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// VPNProvider represents the detected VPN provider
type VPNProvider struct {
	Name     string
	Type     string // "nordvpn", "openvpn", "wireguard", "protonvpn", "mullvad", "other"
	Active   bool
	Location string
	Server   string
}

// NordVPN country code to name mapping
var nordVPNCountryCodes = map[string]string{
	"us": "United States",
	"uk": "United Kingdom",
	"ca": "Canada",
	"au": "Australia",
	"de": "Germany",
	"fr": "France",
	"jp": "Japan",
	"nl": "Netherlands",
	"se": "Sweden",
	"ch": "Switzerland",
	"sg": "Singapore",
	"br": "Brazil",
	"in": "India",
	"kr": "South Korea",
	"mx": "Mexico",
	"es": "Spain",
	"it": "Italy",
	"pl": "Poland",
	"ru": "Russia",
	"za": "South Africa",
	"tr": "Turkey",
	"ae": "UAE",
	"ar": "Argentina",
	"at": "Austria",
	"be": "Belgium",
	"bg": "Bulgaria",
	"cz": "Czech Republic",
	"dk": "Denmark",
	"fi": "Finland",
	"gr": "Greece",
	"hk": "Hong Kong",
	"hu": "Hungary",
	"id": "Indonesia",
	"ie": "Ireland",
	"il": "Israel",
	"is": "Iceland",
	"lu": "Luxembourg",
	"my": "Malaysia",
	"no": "Norway",
	"nz": "New Zealand",
	"pt": "Portugal",
	"ro": "Romania",
	"rs": "Serbia",
	"sk": "Slovakia",
	"th": "Thailand",
	"tw": "Taiwan",
	"ua": "Ukraine",
	"vn": "Vietnam",
}

// VPNDetector handles detection of VPN connections
type VPNDetector struct {
	lastCheck      time.Time
	cachedResult   bool
	cachedProvider *VPNProvider
}

// NewVPNDetector creates a new VPN detector
func NewVPNDetector() *VPNDetector {
	return &VPNDetector{}
}

// IsVPNActive performs comprehensive VPN detection
func (vd *VPNDetector) IsVPNActive() (bool, *VPNProvider) {
	// Use cache for 30 seconds to avoid excessive checks
	if time.Since(vd.lastCheck) < 30*time.Second && vd.cachedProvider != nil {
		return vd.cachedResult, vd.cachedProvider
	}

	provider := &VPNProvider{
		Name:   "Unknown",
		Type:   "unknown",
		Active: false,
	}

	// Detection methods in order of reliability:
	// 1. Check NordVPN CLI (most reliable for NordVPN)
	if vd.checkNordVPN(provider) {
		vd.cacheResult(true, provider)
		return true, provider
	}

	// 2. Check for VPN network interfaces
	if vd.checkVPNInterfaces(provider) {
		// Even if interface exists, verify it's actually routing traffic
		if vd.verifyVPNTraffic() {
			vd.cacheResult(true, provider)
			return true, provider
		}
	}

	// 3. Check for VPN processes
	if vd.checkVPNProcesses(provider) {
		// Verify traffic is actually going through VPN
		if vd.verifyVPNTraffic() {
			vd.cacheResult(true, provider)
			return true, provider
		}
	}

	vd.cacheResult(false, provider)
	return false, provider
}

// cacheResult caches the detection result
func (vd *VPNDetector) cacheResult(active bool, provider *VPNProvider) {
	vd.lastCheck = time.Now()
	vd.cachedResult = active
	vd.cachedProvider = provider
}

// checkNordVPN specifically checks for NordVPN connection
func (vd *VPNDetector) checkNordVPN(provider *VPNProvider) bool {
	// Check if nordvpn CLI exists
	if _, err := exec.LookPath("nordvpn"); err != nil {
		log.Printf("[VPN] NordVPN CLI not found")
		return false
	}

	// Get NordVPN status
	cmd := exec.Command("nordvpn", "status")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("[VPN] Failed to get NordVPN status: %v", err)
		return false
	}

	status := string(output)
	log.Printf("[VPN] NordVPN status output: %s", status)

	// Check if connected
	if strings.Contains(status, "Status: Connected") {
		provider.Name = "NordVPN"
		provider.Type = "nordvpn"
		provider.Active = true

		// Parse server info
		provider.Server = vd.parseNordVPNOutput(status, "Server")

		// Parse country (this is the human-readable country name)
		if country := vd.parseNordVPNOutput(status, "Country"); country != "" {
			provider.Location = country
		} else if hostname := vd.parseNordVPNOutput(status, "Hostname"); hostname != "" {
			// Fallback: parse country code from hostname (e.g., "us1234.nordvpn.com" -> "us" -> "United States")
			countryCode := vd.parseNordVPNCountryCode(hostname)
			if countryName, ok := nordVPNCountryCodes[countryCode]; ok {
				provider.Location = countryName
			} else if countryCode != "" {
				provider.Location = strings.ToUpper(countryCode)
			} else {
				provider.Location = provider.Server
			}
		} else {
			provider.Location = provider.Server
		}

		log.Printf("[VPN] NordVPN is connected to: %s (Location: %s)", provider.Server, provider.Location)
		return true
	}

	log.Printf("[VPN] NordVPN is disconnected")
	return false
}

// parseNordVPNCountryCode extracts country code from server name (e.g., "us1234.nordvpn.com" -> "us")
func (vd *VPNDetector) parseNordVPNCountryCode(server string) string {
	// Remove domain part
	server = strings.TrimSuffix(server, ".nordvpn.com")
	server = strings.TrimSuffix(server, ".nordvpn")

	// Extract first 2 letters which is the country code
	if len(server) >= 2 {
		return strings.ToLower(server[:2])
	}
	return ""
}

// parseNordVPNOutput parses NordVPN status output for specific fields
func (vd *VPNDetector) parseNordVPNOutput(output, field string) string {
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, field+":") {
			return strings.TrimSpace(strings.TrimPrefix(line, field+":"))
		}
	}
	return ""
}

// checkVPNInterfaces checks for VPN network interfaces
func (vd *VPNDetector) checkVPNInterfaces(provider *VPNProvider) bool {
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Printf("[VPN] Failed to get network interfaces: %v", err)
		return false
	}

	for _, iface := range interfaces {
		name := iface.Name

		// Check for common VPN interface patterns
		if vd.isVPNInterface(name) {
			// Check if interface is up
			if iface.Flags&net.FlagUp != 0 {
				provider.Type = vd.getVPNTypeFromInterface(name)
				provider.Name = vd.getVPNNameFromInterface(name)
				provider.Active = true

				// Get IP address for this interface
				addrs, err := iface.Addrs()
				if err == nil && len(addrs) > 0 {
					log.Printf("[VPN] Found active VPN interface: %s (%s)", name, addrs[0])
				}

				return true
			}
		}
	}

	return false
}

// isVPNInterface checks if interface name matches VPN patterns
func (vd *VPNDetector) isVPNInterface(name string) bool {
	vpnPatterns := []string{
		"tun",        // OpenVPN, generic TUN
		"tap",        // OpenVPN TAP
		"wg",         // WireGuard
		"nordlynx",   // NordVPN WireGuard
		"nordtun",    // NordVPN TUN
		"ppp",        // PPTP/L2TP
		"ipsec",      // IPsec
		"proton",     // ProtonVPN
		"mullvad",    // Mullvad
		"windscribe", // Windscribe
	}

	lowerName := strings.ToLower(name)
	for _, pattern := range vpnPatterns {
		if strings.HasPrefix(lowerName, pattern) {
			return true
		}
	}
	return false
}

// getVPNTypeFromInterface determines VPN type from interface name
func (vd *VPNDetector) getVPNTypeFromInterface(name string) string {
	lowerName := strings.ToLower(name)
	switch {
	case strings.HasPrefix(lowerName, "nord"):
		return "nordvpn"
	case strings.HasPrefix(lowerName, "wg"):
		return "wireguard"
	case strings.HasPrefix(lowerName, "tun") || strings.HasPrefix(lowerName, "tap"):
		return "openvpn"
	case strings.HasPrefix(lowerName, "ppp"):
		return "pptp"
	case strings.HasPrefix(lowerName, "proton"):
		return "protonvpn"
	case strings.HasPrefix(lowerName, "mullvad"):
		return "mullvad"
	default:
		return "other"
	}
}

// getVPNNameFromInterface determines VPN name from interface name
func (vd *VPNDetector) getVPNNameFromInterface(name string) string {
	lowerName := strings.ToLower(name)
	switch {
	case strings.HasPrefix(lowerName, "nord"):
		return "NordVPN"
	case strings.HasPrefix(lowerName, "proton"):
		return "ProtonVPN"
	case strings.HasPrefix(lowerName, "mullvad"):
		return "Mullvad VPN"
	case strings.HasPrefix(lowerName, "wg"):
		return "WireGuard"
	case strings.HasPrefix(lowerName, "tun") || strings.HasPrefix(lowerName, "tap"):
		return "OpenVPN"
	default:
		return "VPN"
	}
}

// checkVPNProcesses checks for running VPN processes
func (vd *VPNDetector) checkVPNProcesses(provider *VPNProvider) bool {
	vpnProcesses := map[string]string{
		"nordvpnd":              "NordVPN",
		"openvpn":               "OpenVPN",
		"wg-quick":              "WireGuard",
		"wireguard-go":          "WireGuard",
		"protonvpn":             "ProtonVPN",
		"mullvad-vpn":           "Mullvad VPN",
		"expressvpnd":           "ExpressVPN",
		"windscribe":            "Windscribe",
		"surfshark":             "Surfshark",
		"cyberghost":            "CyberGhost",
		"privateinternetaccess": "PIA",
		"pia-client":            "PIA",
		"hotspotshield":         "Hotspot Shield",
		"tunnelbear":            "TunnelBear",
	}

	// Check /proc for running processes
	entries, err := os.ReadDir("/proc")
	if err != nil {
		log.Printf("[VPN] Failed to read /proc: %v", err)
		return false
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Check if directory name is a PID
		pid := entry.Name()
		if _, err := fmt.Sscanf(pid, "%d", new(int)); err != nil {
			continue
		}

		// Read process command line
		cmdlinePath := filepath.Join("/proc", pid, "cmdline")
		data, err := os.ReadFile(cmdlinePath)
		if err != nil {
			continue
		}

		cmdline := string(data)
		lowerCmdline := strings.ToLower(cmdline)

		// Check for VPN processes
		for processName, vpnName := range vpnProcesses {
			if strings.Contains(lowerCmdline, strings.ToLower(processName)) {
				// Found a VPN process
				provider.Name = vpnName
				provider.Type = strings.ToLower(strings.ReplaceAll(vpnName, " ", ""))
				log.Printf("[VPN] Found VPN process: %s (PID: %s)", vpnName, pid)
				return true
			}
		}
	}

	return false
}

// verifyVPNTraffic verifies that traffic is actually going through VPN
func (vd *VPNDetector) verifyVPNTraffic() bool {
	// Check default route
	cmd := exec.Command("ip", "route", "show", "default")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("[VPN] Failed to check default route: %v", err)
		return false
	}

	route := string(output)
	log.Printf("[VPN] Default route: %s", strings.TrimSpace(route))

	// Check if default route goes through a VPN interface
	if strings.Contains(route, "tun") ||
		strings.Contains(route, "wg") ||
		strings.Contains(route, "nord") ||
		strings.Contains(route, "ppp") {
		log.Printf("[VPN] Traffic is routing through VPN interface")
		return true
	}

	// If route goes through regular interface, VPN might be in split-tunnel mode
	// or not actually routing traffic
	return false
}

// isPrivateIP checks if an IP is private/local
func (vd *VPNDetector) isPrivateIP(ip string) bool {
	privateRanges := []string{
		"192.168.",
		"10.",
		"172.16.", "172.17.", "172.18.", "172.19.", "172.20.",
		"172.21.", "172.22.", "172.23.", "172.24.", "172.25.",
		"172.26.", "172.27.", "172.28.", "172.29.", "172.30.", "172.31.",
		"127.0.0.1",
		"::1",
		"fc00:",
		"fe80:",
	}

	for _, prefix := range privateRanges {
		if strings.HasPrefix(ip, prefix) {
			return true
		}
	}
	return false
}
