import { Card, CardContent } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { useTasks, useCancelTask, useRestartTask, useDeleteTask, useClearCompletedTasks, useClearFailedTasks } from '@/hooks/useApi'
import { useAppStore } from '@/store/appStore'
import { getStatusBgColor, formatBytes, formatDateTime } from '@/lib/utils'
import { Pause, Trash2, RotateCcw, CheckCircle, XCircle } from 'lucide-react'

export function Downloads() {
  const { data: tasks, isLoading } = useTasks()
  const { addToast } = useAppStore()
  
  const cancelTask = useCancelTask()
  const restartTask = useRestartTask()
  const deleteTask = useDeleteTask()
  const clearCompleted = useClearCompletedTasks()
  const clearFailed = useClearFailedTasks()

  const handleCancel = async (id: number) => {
    try {
      await cancelTask.mutateAsync(id)
      addToast('Task cancelled', 'success')
    } catch {
      addToast('Failed to cancel task', 'error')
    }
  }

  const handleRestart = async (id: number) => {
    try {
      await restartTask.mutateAsync(id)
      addToast('Task restarted', 'success')
    } catch {
      addToast('Failed to restart task', 'error')
    }
  }

  const handleDelete = async (id: number) => {
    try {
      await deleteTask.mutateAsync(id)
      addToast('Task deleted', 'success')
    } catch {
      addToast('Failed to delete task', 'error')
    }
  }

  const handleClearCompleted = async () => {
    try {
      await clearCompleted.mutateAsync()
      addToast('Completed tasks cleared', 'success')
    } catch {
      addToast('Failed to clear completed tasks', 'error')
    }
  }

  const handleClearFailed = async () => {
    try {
      await clearFailed.mutateAsync()
      addToast('Failed tasks cleared', 'success')
    } catch {
      addToast('Failed to clear failed tasks', 'error')
    }
  }

  if (isLoading) {
    return <div>Loading...</div>
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Downloads</h1>
          <p className="text-muted-foreground">
            Manage your download queue
          </p>
        </div>
        <div className="flex gap-2">
          <Button
            variant="outline"
            onClick={handleClearCompleted}
            disabled={clearCompleted.isPending}
          >
            <CheckCircle className="mr-2 h-4 w-4" />
            Clear Completed
          </Button>
          <Button
            variant="outline"
            onClick={handleClearFailed}
            disabled={clearFailed.isPending}
          >
            <XCircle className="mr-2 h-4 w-4" />
            Clear Failed
          </Button>
        </div>
      </div>

      <div className="space-y-4">
        {tasks?.map((task) => (
          <Card key={task.id}>
            <CardContent className="p-6">
              <div className="flex items-start justify-between">
                <div className="space-y-1">
                  <h3 className="font-semibold">{task.title}</h3>
                  <div className="flex items-center gap-2 text-sm text-muted-foreground">
                    <Badge className={getStatusBgColor(task.status)}>
                      {task.status}
                    </Badge>
                    <span>{formatBytes(task.size)}</span>
                    <span>•</span>
                    <span>{task.seeders} seeders</span>
                    {task.started_at && (
                      <>
                        <span>•</span>
                        <span>Started {formatDateTime(task.started_at)}</span>
                      </>
                    )}
                  </div>
                </div>

                <div className="flex items-center gap-2">
                  {task.status === 'downloading' && (
                    <Button
                      variant="outline"
                      size="icon"
                      onClick={() => handleCancel(task.id)}
                      disabled={cancelTask.isPending}
                    >
                      <Pause className="h-4 w-4" />
                    </Button>
                  )}
                  
                  {(task.status === 'failed' || task.status === 'cancelled') && (
                    <Button
                      variant="outline"
                      size="icon"
                      onClick={() => handleRestart(task.id)}
                      disabled={restartTask.isPending}
                    >
                      <RotateCcw className="h-4 w-4" />
                    </Button>
                  )}

                  <Button
                    variant="outline"
                    size="icon"
                    onClick={() => handleDelete(task.id)}
                    disabled={deleteTask.isPending}
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </div>
              </div>

              {/* Progress bar */}
              <div className="mt-4">
                <div className="h-2 w-full rounded-full bg-secondary">
                  <div
                    className="h-2 rounded-full bg-primary transition-all"
                    style={{ width: `${task.progress}%` }}
                  />
                </div>
                <p className="mt-1 text-sm text-muted-foreground">
                  {task.progress.toFixed(1)}% complete
                </p>
              </div>

              {task.error && (
                <p className="mt-2 text-sm text-red-500">
                  Error: {task.error}
                </p>
              )}
            </CardContent>
          </Card>
        ))}

        {tasks?.length === 0 && (
          <Card>
            <CardContent className="p-6 text-center">
              <p className="text-muted-foreground">No downloads yet</p>
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  )
}