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
    log "Configuring DNS..."
    if [ -n "$PIHOLE_IP" ]; then
        echo "nameserver $PIHOLE_IP" > /etc/resolv.conf
        log "Using PiHole DNS: $PIHOLE_IP"
    else
        echo "nameserver 9.9.9.9" > /etc/resolv.conf
        echo "nameserver 149.112.112.112" >> /etc/resolv.conf
        log "Using Quad9 DNS"
    fi
}

setup_vpn() {
    log "Setting up VPN..."
    
    # Check if nordvpn is installed
    if ! command -v nordvpn &> /dev/null; then
        log "WARNING: NordVPN not installed, skipping VPN setup"
        return 0
    fi
    
    if [ -z "$NORDVPN_TOKEN" ]; then
        log "WARNING: NORDVPN_TOKEN not set, skipping VPN"
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
    nordvpn login --token "$NORDVPN_TOKEN" || {
        log "ERROR: Failed to login"
        return 1
    }
    
    log "Configuring..."
    nordvpn set technology "$VPN_PROTOCOL"
    nordvpn set killswitch off
    nordvpn set dns off 2>/dev/null || true
    
    log "Connecting to VPN..."
    if [ -n "$NORDVPN_SERVER" ]; then
        nordvpn connect "$NORDVPN_SERVER"
    else
        nordvpn connect
    fi
    
    log "Waiting for connection..."
    for i in {1..30}; do
        if nordvpn status | grep -q "Status: Connected"; then
            log "VPN connected!"
            return 0
        fi
        sleep 1
    done
    
    log "WARNING: VPN connection failed, continuing without VPN"
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
    if ! command -v iptables &> /dev/null; then
        log "iptables not available, skipping firewall setup"
        return 0
    fi
    
    if ! command -v nordvpn &> /dev/null; then
        log "NordVPN not installed, using basic firewall rules"
        iptables -P INPUT ACCEPT
        iptables -P OUTPUT ACCEPT
        iptables -P FORWARD DROP
        iptables -F 2>/dev/null || true
        iptables -A INPUT -p tcp --dport 8080 -j ACCEPT
        iptables -A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT
        log "Basic firewall configured"
        return 0
    fi
    
    log "Setting up firewall..."
    
    iptables -F 2>/dev/null || true
    iptables -t nat -F 2>/dev/null || true
    
    iptables -P OUTPUT DROP
    iptables -P INPUT DROP
    iptables -P FORWARD DROP
    
    iptables -A INPUT -i lo -j ACCEPT
    iptables -A OUTPUT -o lo -j ACCEPT
    
    iptables -A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT
    iptables -A OUTPUT -m state --state ESTABLISHED,RELATED -j ACCEPT
    
    iptables -A OUTPUT -o tun+ -j ACCEPT
    iptables -A OUTPUT -o nordlynx -j ACCEPT
    iptables -A OUTPUT -o wg+ -j ACCEPT
    
    iptables -A OUTPUT -p udp --dport 53 -j ACCEPT
    iptables -A OUTPUT -p tcp --dport 80 -j ACCEPT
    iptables -A OUTPUT -p tcp --dport 443 -j ACCEPT
    
    iptables -A INPUT -p tcp --dport 8080 -j ACCEPT
    iptables -A OUTPUT -p tcp --sport 8080 -j ACCEPT
    
    iptables -A OUTPUT -p tcp --dport 6881:6889 -j ACCEPT
    iptables -A OUTPUT -p udp --dport 6881:6889 -j ACCEPT
    
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
