import { useState } from 'react'
import { StreamPlayer } from './StreamPlayer'
import { VideoPlayer } from './VideoPlayer'
import { X } from 'lucide-react'

interface VideoModalProps {
  isOpen: boolean
  onClose: () => void
  videoPath?: string
  jobId?: string
  poster?: string
  title?: string
}

export function VideoModal({ isOpen, onClose, videoPath, jobId, poster, title }: VideoModalProps) {
  const [useHLS, setUseHLS] = useState(true)

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 z-50 bg-black/95 flex items-center justify-center"
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose()
      }}
    >
      <div className="relative w-full h-full max-w-[1920px] mx-auto">
        {/* Close button */}
        <button
          onClick={onClose}
          className="absolute top-4 right-4 z-10 w-10 h-10 rounded-full bg-black/50 flex items-center justify-center text-white hover:bg-black/70 transition-colors"
        >
          <X className="w-6 h-6" />
        </button>

        {/* Stream type toggle */}
        <div className="absolute top-4 left-4 z-10 flex items-center gap-2 bg-black/50 rounded-full px-3 py-1.5"
          onClick={(e) => e.stopPropagation()}
        >
          <span className="text-white/70 text-sm">Direct</span>
          <button
            onClick={() => setUseHLS(!useHLS)}
            className={`w-10 h-5 rounded-full transition-colors relative ${useHLS ? 'bg-[#1ed760]' : 'bg-white/30'}`}
          >
            <div className={`w-4 h-4 rounded-full bg-white absolute top-0.5 transition-transform ${useHLS ? 'translate-x-5' : 'translate-x-0.5'}`} />
          </button>
          <span className="text-white/70 text-sm">HLS</span>
        </div>

        {/* Player */}
        {useHLS && jobId ? (
          <StreamPlayer
            jobId={jobId}
            poster={poster}
            title={title}
            onClose={onClose}
          />
        ) : (
          <VideoPlayer
            src={`/api/stream/direct?path=${encodeURIComponent(videoPath || '')}`}
            poster={poster}
            title={title}
            onClose={onClose}
            autoPlay
          />
        )}
      </div>
    </div>
  )
}
