import { useVPNStatus } from '@/hooks/useApi'
import { Shield, ShieldAlert, MapPin, Network } from 'lucide-react'
import { cn } from '@/lib/utils'

export function VPNStatus() {
  const { data: vpnStatus, isLoading } = useVPNStatus()

  if (isLoading) {
    return (
      <div className="flex items-center gap-2 text-sm text-muted-foreground">
        <div className="h-4 w-4 animate-pulse rounded-full bg-muted" />
        <span className="hidden sm:inline">Checking VPN...</span>
      </div>
    )
  }

  const isConnected = vpnStatus?.active ?? false
  const hasLocation = vpnStatus?.location || vpnStatus?.country
  const hasIP = vpnStatus?.ip
  const hasProvider = vpnStatus?.provider

  return (
    <div
      className={cn(
        'flex items-center gap-2 rounded-md px-3 py-1.5 text-sm font-medium transition-colors',
        isConnected
          ? 'bg-green-500/10 text-green-500'
          : 'bg-red-500/10 text-red-500'
      )}
      title={vpnStatus?.message || 'VPN Status Unknown'}
    >
      {isConnected ? (
        <Shield className="h-4 w-4" />
      ) : (
        <ShieldAlert className="h-4 w-4" />
      )}
      <span className="hidden sm:inline">
        {isConnected ? 'VPN Connected' : 'VPN Disconnected'}
      </span>
      <span className="sm:hidden">
        {isConnected ? 'VPN' : 'No VPN'}
      </span>
      
      {isConnected && hasLocation && (
        <div className="flex items-center gap-1 ml-1 text-xs opacity-80">
          <MapPin className="h-3 w-3" />
          <span>
            {vpnStatus.location}
            {vpnStatus.country && `, ${vpnStatus.country}`}
          </span>
        </div>
      )}
      
      {isConnected && hasProvider && (
        <div className="flex items-center gap-1 ml-1 text-xs opacity-80">
          <Network className="h-3 w-3" />
          <span>{vpnStatus.provider}</span>
        </div>
      )}
      
      {isConnected && hasIP && !hasLocation && !hasProvider && (
        <div className="flex items-center gap-1 ml-1 text-xs opacity-80">
          <span>IP: {vpnStatus.ip}</span>
        </div>
      )}
    </div>
  )
}
