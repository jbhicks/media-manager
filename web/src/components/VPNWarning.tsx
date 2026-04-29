import { useVPNStatus } from '@/hooks/useApi'
import { ShieldAlert, X } from 'lucide-react'
import { useState } from 'react'
import { cn } from '@/lib/utils'

export function VPNWarning() {
  const { data: vpnStatus, isLoading } = useVPNStatus()
  const [dismissed, setDismissed] = useState(false)

  // Don't show while loading or when connected
  if (isLoading || vpnStatus?.active || dismissed) {
    return null
  }

  return (
    <div className={cn(
      'relative flex items-center gap-3 rounded-lg border px-4 py-3 mb-6',
      'bg-red-500/10 border-red-500/20 text-red-500'
    )}>
      <ShieldAlert className="h-5 w-5 flex-shrink-0" />
      <div className="flex-1">
        <p className="font-medium">VPN Disconnected</p>
        <p className="text-sm opacity-90">
          Downloads are disabled for your security. Please connect to your VPN to resume downloading.
        </p>
      </div>
      <button
        onClick={() => setDismissed(true)}
        className="flex-shrink-0 rounded-md p-1 hover:bg-red-500/20 transition-colors"
        aria-label="Dismiss warning"
      >
        <X className="h-4 w-4" />
      </button>
    </div>
  )
}
