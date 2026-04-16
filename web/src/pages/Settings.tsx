import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { useSources, useRules, useVPNStatus } from '@/hooks/useApi'
import { Settings2, Database, Server, Shield, ShieldCheck, ShieldAlert, MapPin, Globe, Wifi } from 'lucide-react'
import { cn } from '@/lib/utils'

export function Settings() {
  const { data: sources } = useSources()
  const { data: rules } = useRules()
  const { data: vpnStatus, isLoading: vpnLoading } = useVPNStatus()

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Settings</h1>
        <p className="text-muted-foreground">
          Configure your media manager
        </p>
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Database className="h-5 w-5" />
              Download Sources
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground mb-4">
              {sources?.length || 0} sources configured
            </p>
            <Button variant="outline" className="w-full">
              Manage Sources
            </Button>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Settings2 className="h-5 w-5" />
              Download Rules
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground mb-4">
              {rules?.length || 0} rules configured
            </p>
            <Button variant="outline" className="w-full">
              Manage Rules
            </Button>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Server className="h-5 w-5" />
              Server Settings
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground mb-4">
              Configure server and connection settings
            </p>
            <Button variant="outline" className="w-full">
              Configure Server
            </Button>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Shield className="h-5 w-5" />
              Security
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground mb-4">
              Manage API keys and authentication
            </p>
            <Button variant="outline" className="w-full">
              Security Settings
            </Button>
          </CardContent>
        </Card>

        {/* VPN Status Card */}
        <Card className={cn(
          vpnStatus?.active ? 'border-green-500/50' : 'border-red-500/50'
        )}>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              {vpnStatus?.active ? (
                <ShieldCheck className="h-5 w-5 text-green-500" />
              ) : (
                <ShieldAlert className="h-5 w-5 text-red-500" />
              )}
              VPN Status
            </CardTitle>
            <CardDescription>
              Connection security status
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium">Status</span>
                <Badge 
                  variant={vpnStatus?.active ? 'default' : 'destructive'}
                  className={cn(
                    vpnStatus?.active && 'bg-green-500 hover:bg-green-600'
                  )}
                >
                  {vpnLoading ? 'Checking...' : vpnStatus?.status === 'connected' ? 'Connected' : 'Disconnected'}
                </Badge>
              </div>
              
              {/* Show provider when connected */}
              {vpnStatus?.active && vpnStatus?.provider && (
                <div className="flex items-center gap-2 text-sm">
                  <Wifi className="h-4 w-4 text-muted-foreground" />
                  <span className="text-muted-foreground">Provider:</span>
                  <span className="font-medium">{vpnStatus.provider}</span>
                </div>
              )}
              
              {/* Show location when connected */}
              {vpnStatus?.active && (vpnStatus?.location || vpnStatus?.country) && (
                <div className="flex items-center gap-2 text-sm">
                  <MapPin className="h-4 w-4 text-muted-foreground" />
                  <span className="text-muted-foreground">Location:</span>
                  <span className="font-medium">
                    {vpnStatus.location}
                    {vpnStatus.location && vpnStatus.country && ', '}
                    {vpnStatus.country}
                  </span>
                </div>
              )}
              
              {/* Show IP when connected */}
              {vpnStatus?.active && vpnStatus?.ip && (
                <div className="flex items-center gap-2 text-sm">
                  <Globe className="h-4 w-4 text-muted-foreground" />
                  <span className="text-muted-foreground">IP Address:</span>
                  <span className="font-medium font-mono">{vpnStatus.ip}</span>
                </div>
              )}
              
              {!vpnStatus?.active && (
                <p className="text-sm text-muted-foreground">
                  {vpnStatus?.message || 'Checking VPN status...'}
                </p>
              )}

              <div className="rounded-md bg-muted p-3">
                <p className="text-xs text-muted-foreground">
                  VPN status is automatically checked every 30 seconds to ensure secure downloads.
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}