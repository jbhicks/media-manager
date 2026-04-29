import { Component, ReactNode, ErrorInfo } from 'react'
import { Button } from '@/components/ui/Button'
import { AlertTriangle, RotateCcw } from 'lucide-react'

interface Props {
  children: ReactNode
}

interface State {
  hasError: boolean
  error: Error | null
}

export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props)
    this.state = { hasError: false, error: null }
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('ErrorBoundary caught an error:', error, errorInfo)
  }

  handleReset = () => {
    this.setState({ hasError: false, error: null })
    window.location.reload()
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="min-h-screen flex items-center justify-center p-4">
          <div className="max-w-md w-full space-y-6 text-center">
            <div className="flex justify-center">
              <div className="rounded-full bg-red-500/10 p-4">
                <AlertTriangle className="h-12 w-12 text-red-500" />
              </div>
            </div>
            
            <h1 className="text-2xl font-bold tracking-tight">
              Something went wrong
            </h1>
            
            <p className="text-muted-foreground">
              The application encountered an unexpected error. We've logged the details for debugging.
            </p>
            
            {this.state.error && (
              <div className="rounded-lg border bg-muted p-4 text-left">
                <p className="text-sm font-mono text-red-500 break-words">
                  {this.state.error.message}
                </p>
              </div>
            )}
            
            <Button onClick={this.handleReset} className="gap-2">
              <RotateCcw className="h-4 w-4" />
              Reload Application
            </Button>
          </div>
        </div>
      )
    }

    return this.props.children
  }
}