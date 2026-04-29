import { useAppStore } from '@/store/appStore'
import { cn } from '@/lib/utils'
import { X } from 'lucide-react'

export function Toaster() {
  const { toasts, removeToast } = useAppStore()

  return (
    <div className="fixed bottom-4 right-4 z-50 flex flex-col gap-2">
      {toasts.map((toast) => (
        <div
          key={toast.id}
          className={cn(
            'flex items-center gap-3 rounded-md border px-4 py-3 shadow-lg animate-in slide-in-from-right',
            toast.type === 'success' && 'border-green-500/20 bg-green-500/10 text-green-500',
            toast.type === 'error' && 'border-red-500/20 bg-red-500/10 text-red-500',
            toast.type === 'info' && 'border-blue-500/20 bg-blue-500/10 text-blue-500'
          )}
        >
          <span className="text-sm font-medium max-w-md break-words">{toast.message}</span>
          <button
            onClick={() => removeToast(toast.id)}
            className="inline-flex items-center justify-center rounded-md text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground h-6 w-6"
          >
            <X className="h-4 w-4" />
          </button>
        </div>
      ))}
    </div>
  )
}