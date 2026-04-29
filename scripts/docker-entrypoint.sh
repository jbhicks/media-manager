#!/bin/bash
set -e

NORDVPN_TOKEN="${NORDVPN_TOKEN:-}"
NORDVPN_SERVER="${NORDVPN_SERVER:-}"
VPN_PROTOCOL="${VPN_PROTOCOL:-nordlynx}"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

if [ ! -c /dev/net/tun ]; then
    log "ERROR: /dev/net/tun not found"
    exit 1
fi

dns_config() {
    if [ -n "$PIHOLE_IP" ]; then
        log "Configuring DNS..."
        echo "nameserver $PIHOLE_IP" > /etc/resolv.conf
        log "Using PiHole DNS: $PIHOLE_IP"
    fi
}

setup_vpn() {
    log "Setting up VPN..."
    
    # Check if nordvpn is installed
    if ! command -v nordvpn &> /dev/null; then
        log "WARNING: NordVPN not installed, skipping VPN setup"
        return 0
    fi
    
    # Check if we have credentials
    if [ -z "$NORDVPN_TOKEN" ] && [ -z "$NORDVPN_USERNAME" ]; then
        log "WARNING: No NordVPN credentials set, skipping VPN"
        return 0
    fi
    
    # Start NordVPN daemon (required in Docker without systemd)
    log "Starting NordVPN daemon..."
    if [ -f /etc/init.d/nordvpn ]; then
        /etc/init.d/nordvpn start
    elif [ -f /usr/sbin/nordvpnd ]; then
        /usr/sbin/nordvpnd &
        sleep 2
    fi
    
    # Wait for daemon to be ready
    for i in {1..10}; do
        if [ -S /run/nordvpn/nordvpnd.sock ]; then
            log "NordVPN daemon is ready"
            break
        fi
        sleep 1
    done
    
    log "Logging in to NordVPN..."
    if [ -n "$NORDVPN_TOKEN" ]; then
        # Try token-based login first
        # Pipe 'n' to auto-decline analytics prompt on fresh installs
        printf "n\n" | nordvpn login --token "$NORDVPN_TOKEN" || {
            log "WARNING: Token login failed, trying username/password..."
            # Fall back to username/password
            if [ -n "$NORDVPN_USERNAME" ] && [ -n "$NORDVPN_PASSWORD" ]; then
                printf "%s\n%s\n" "$NORDVPN_USERNAME" "$NORDVPN_PASSWORD" | nordvpn login || {
                    log "ERROR: Failed to login with username/password"
                    return 1
                }
            else
                log "ERROR: Failed to login"
                return 1
            fi
        }
    elif [ -n "$NORDVPN_USERNAME" ] && [ -n "$NORDVPN_PASSWORD" ]; then
        # Use username/password login
        printf "%s\n%s\n" "$NORDVPN_USERNAME" "$NORDVPN_PASSWORD" | nordvpn login || {
            log "ERROR: Failed to login"
            return 1
        }
    fi
    
    log "Configuring..."
    nordvpn set technology "$VPN_PROTOCOL"
    nordvpn set killswitch off
    nordvpn set dns off 2>/dev/null || true
    nordvpn set firewall off 2>/dev/null || true
    
    log "Waiting for daemon to stabilize..."
    sleep 3
    
    # Verify login actually worked
    log "Verifying login status..."
    nordvpn account || log "WARNING: Account check failed"
    
    # Log current settings for debugging
    log "Current NordVPN settings:"
    nordvpn settings || true
    
    log "Configuring whitelist..."
    # Add subnets to whitelist to allow Docker/LAN traffic through VPN
    # Default includes common private networks
    WHITELIST_SUBNETS="${VPN_WHITELIST_SUBNETS:-192.168.1.0/24,10.0.0.0/8,172.16.0.0/12}"
    if [ -n "$WHITELIST_SUBNETS" ]; then
        IFS=',' read -ra SUBNETS <<< "$WHITELIST_SUBNETS"
        for subnet in "${SUBNETS[@]}"; do
            subnet=$(echo "$subnet" | xargs) # trim whitespace
            if [ -n "$subnet" ]; then
                log "Adding whitelist subnet: $subnet"
                nordvpn whitelist add subnet "$subnet" 2>/dev/null || log "WARNING: Failed to whitelist $subnet"
            fi
        done
    fi
    
    # Verify basic connectivity before attempting VPN connection
    log "Checking internet connectivity..."
    if ! curl -fsSL --max-time 5 https://api.ipify.org >/dev/null 2>&1; then
        log "WARNING: No internet connectivity detected, VPN may not work"
    fi
    
    log "Connecting to VPN..."
    CONNECTED=0
    
    # Try NordLynx first (if configured), then fallback to OpenVPN
    for proto in "$VPN_PROTOCOL" "openvpn"; do
        if [ "$proto" != "$VPN_PROTOCOL" ]; then
            log "Trying fallback protocol: $proto"
            nordvpn set technology "$proto" || true
            sleep 2
        fi
        
        for attempt in {1..3}; do
            log "Connection attempt $attempt/3 using $proto..."
            if [ -n "$NORDVPN_SERVER" ]; then
                if nordvpn connect "$NORDVPN_SERVER"; then
                    CONNECTED=1
                    break 2
                fi
            else
                if nordvpn connect; then
                    CONNECTED=1
                    break 2
                fi
            fi
            log "Connect failed, retrying in 5s..."
            sleep 5
        done
    done
    
    if [ "$CONNECTED" -ne 1 ]; then
        log "WARNING: VPN connection failed after all attempts, continuing without VPN"
        return 0
    fi
    
    log "Waiting for connection..."
    for i in {1..30}; do
        if nordvpn status | grep -q "Status: Connected"; then
            log "VPN connected!"
            # Re-apply whitelist after connection (some NordVPN versions reset it)
            if [ -n "$WHITELIST_SUBNETS" ]; then
                IFS=',' read -ra SUBNETS <<< "$WHITELIST_SUBNETS"
                for subnet in "${SUBNETS[@]}"; do
                    subnet=$(echo "$subnet" | xargs)
                    if [ -n "$subnet" ]; then
                        nordvpn whitelist add subnet "$subnet" 2>/dev/null || true
                    fi
                done
            fi
            return 0
        fi
        sleep 1
    done
    
    log "WARNING: VPN status check failed, continuing without VPN"
    return 0
}

verify_vpn() {
    if command -v nordvpn &> /dev/null && nordvpn status | grep -q "Status: Connected"; then
        log "Verifying VPN..."
        PUBLIC_IP=$(curl -fsSL --max-time 10 https://api.ipify.org 2>/dev/null || echo "unknown")
        log "Public IP: $PUBLIC_IP"
    else
        log "VPN not active, skipping verification"
    fi
}

setup_firewall() {
    if ! command -v iptables-nft &> /dev/null; then
        log "iptables-nft not available, skipping firewall setup"
        return 0
    fi
    
    # Use iptables-nft to avoid legacy/nft conflicts
    IPT="iptables-nft"
    
    if ! command -v nordvpn &> /dev/null; then
        log "NordVPN not installed, using basic firewall rules"
        $IPT -P INPUT ACCEPT
        $IPT -P OUTPUT ACCEPT
        $IPT -P FORWARD DROP
        $IPT -F 2>/dev/null || true
        $IPT -A INPUT -p tcp --dport 8080 -j ACCEPT
        $IPT -A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT
        log "Basic firewall configured"
        return 0
    fi
    
    log "Setting up firewall..."
    
    # Add Docker bridge exceptions to mangle table BEFORE NordVPN adds its rules
    # This ensures Docker traffic is allowed
    $IPT -t mangle -I PREROUTING 1 -i eth0 -j ACCEPT 2>/dev/null || true
    $IPT -t mangle -I POSTROUTING 1 -o eth0 -j ACCEPT 2>/dev/null || true
    
    $IPT -F 2>/dev/null || true
    $IPT -t nat -F 2>/dev/null || true
    
    $IPT -P OUTPUT DROP
    $IPT -P INPUT DROP
    $IPT -P FORWARD DROP
    
    # Allow Docker bridge traffic
    $IPT -A INPUT -i eth0 -j ACCEPT
    $IPT -A OUTPUT -o eth0 -j ACCEPT
    
    $IPT -A INPUT -i lo -j ACCEPT
    $IPT -A OUTPUT -o lo -j ACCEPT
    
    $IPT -A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT
    $IPT -A OUTPUT -m state --state ESTABLISHED,RELATED -j ACCEPT
    
    $IPT -A OUTPUT -o tun+ -j ACCEPT
    $IPT -A OUTPUT -o nordlynx -j ACCEPT
    $IPT -A OUTPUT -o wg+ -j ACCEPT
    
    $IPT -A OUTPUT -p udp --dport 53 -j ACCEPT
    $IPT -A OUTPUT -p tcp --dport 80 -j ACCEPT
    $IPT -A OUTPUT -p tcp --dport 443 -j ACCEPT
    
    $IPT -A INPUT -p tcp --dport 8080 -j ACCEPT
    $IPT -A OUTPUT -p tcp --sport 8080 -j ACCEPT
    
    $IPT -A OUTPUT -p tcp --dport 6881:6889 -j ACCEPT
    $IPT -A OUTPUT -p udp --dport 6881:6889 -j ACCEPT
    
    log "Firewall configured"
}

main() {
    log "Starting Media Manager..."
    
    dns_config
    setup_vpn || log "WARNING: VPN setup failed, continuing without VPN"
    verify_vpn || true
    setup_firewall || log "WARNING: Firewall setup failed"
    
    mkdir -p "$(dirname "$DB_PATH")"
    
    log "Starting..."
    
    exec "$@"
}

cleanup() {
    log "Shutting down..."
    nordvpn disconnect 2>/dev/null || true
    exit 0
}

trap cleanup SIGTERM SIGINT

main "$@"
