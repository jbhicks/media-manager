import { useState, useRef, useEffect, useCallback } from 'react'
import { Play, Pause, Volume2, VolumeX, Maximize, SkipBack, SkipForward, Loader2 } from 'lucide-react'

interface StreamPlayerProps {
  jobId: string
  poster?: string
  title?: string
  onClose?: () => void
}

export function StreamPlayer({ jobId, poster, title, onClose }: StreamPlayerProps) {
  const videoRef = useRef<HTMLVideoElement>(null)
  const [isPlaying, setIsPlaying] = useState(false)
  const [currentTime, setCurrentTime] = useState(0)
  const [duration, setDuration] = useState(0)
  const [volume, setVolume] = useState(1)
  const [isMuted, setIsMuted] = useState(false)
  const [isFullscreen, setIsFullscreen] = useState(false)
  const [showControls, setShowControls] = useState(true)
  const [streamStatus, setStreamStatus] = useState<'preparing' | 'ready' | 'error'>('preparing')
  const [progress, setProgress] = useState(0)
  const controlsTimeoutRef = useRef<ReturnType<typeof setTimeout>>()
  const hlsRef = useRef<any>(null)
  const statusIntervalRef = useRef<ReturnType<typeof setInterval>>()

  const API_BASE = '/api'

  // Check stream status
  const checkStatus = useCallback(async () => {
    try {
      const res = await fetch(`${API_BASE}/stream/status?job=${jobId}`)
      if (res.ok) {
        const data = await res.json()
        setStreamStatus(data.status)
        setProgress(data.progress)
        
        if (data.status === 'ready') {
          initPlayer()
          if (statusIntervalRef.current) {
            clearInterval(statusIntervalRef.current)
          }
        } else if (data.status === 'error') {
          if (statusIntervalRef.current) {
            clearInterval(statusIntervalRef.current)
          }
        }
      }
    } catch (err) {
      console.error('Status check failed:', err)
    }
  }, [jobId])

  // Initialize HLS player
  const initPlayer = useCallback(async () => {
    const video = videoRef.current
    if (!video) return

    const playlistUrl = `${API_BASE}/stream/playlist?job=${jobId}`

    // Check if HLS is natively supported
    if (video.canPlayType('application/vnd.apple.mpegurl')) {
      video.src = playlistUrl
    } else {
      // Use hls.js for browsers that don't support HLS natively
      try {
        const Hls = (await import('hls.js')).default
        if (Hls.isSupported()) {
          const hls = new Hls({
            enableWorker: true,
            lowLatencyMode: true,
          })
          hls.loadSource(playlistUrl)
          hls.attachMedia(video)
          hlsRef.current = hls

          hls.on(Hls.Events.MANIFEST_PARSED, () => {
            video.play().catch(() => {})
            setIsPlaying(true)
          })

          hls.on(Hls.Events.ERROR, (_event: any, data: any) => {
            console.error('HLS error:', data)
            if (data.fatal) {
              setStreamStatus('error')
            }
          })
        }
      } catch (err) {
        console.error('Failed to load hls.js:', err)
        // Fallback: try direct video anyway
        video.src = playlistUrl
      }
    }
  }, [jobId])

  useEffect(() => {
    // Start checking status
    checkStatus()
    statusIntervalRef.current = setInterval(checkStatus, 2000)

    return () => {
      if (statusIntervalRef.current) {
        clearInterval(statusIntervalRef.current)
      }
      if (hlsRef.current) {
        hlsRef.current.destroy()
      }
    }
  }, [checkStatus])

  useEffect(() => {
    const video = videoRef.current
    if (!video) return

    const updateTime = () => setCurrentTime(video.currentTime)
    const updateDuration = () => setDuration(video.duration)
    const handleEnded = () => setIsPlaying(false)

    video.addEventListener('timeupdate', updateTime)
    video.addEventListener('loadedmetadata', updateDuration)
    video.addEventListener('ended', handleEnded)

    return () => {
      video.removeEventListener('timeupdate', updateTime)
      video.removeEventListener('loadedmetadata', updateDuration)
      video.removeEventListener('ended', handleEnded)
    }
  }, [streamStatus])

  const togglePlay = () => {
    const video = videoRef.current
    if (!video) return

    if (isPlaying) {
      video.pause()
    } else {
      video.play().catch(() => {})
    }
    setIsPlaying(!isPlaying)
  }

  const handleSeek = (e: React.ChangeEvent<HTMLInputElement>) => {
    const time = parseFloat(e.target.value)
    if (videoRef.current) {
      videoRef.current.currentTime = time
      setCurrentTime(time)
    }
  }

  const handleVolumeChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const vol = parseFloat(e.target.value)
    setVolume(vol)
    if (videoRef.current) {
      videoRef.current.volume = vol
      setIsMuted(vol === 0)
    }
  }

  const toggleMute = () => {
    if (videoRef.current) {
      videoRef.current.muted = !isMuted
      setIsMuted(!isMuted)
    }
  }

  const toggleFullscreen = () => {
    const container = videoRef.current?.parentElement
    if (!container) return

    if (!isFullscreen) {
      container.requestFullscreen?.().catch(() => {})
    } else {
      document.exitFullscreen?.().catch(() => {})
    }
    setIsFullscreen(!isFullscreen)
  }

  const showControlsTemporarily = () => {
    setShowControls(true)
    if (controlsTimeoutRef.current) {
      clearTimeout(controlsTimeoutRef.current)
    }
    controlsTimeoutRef.current = setTimeout(() => {
      if (isPlaying) {
        setShowControls(false)
      }
    }, 3000)
  }

  const formatTime = (seconds: number) => {
    const mins = Math.floor(seconds / 60)
    const secs = Math.floor(seconds % 60)
    return `${mins}:${secs.toString().padStart(2, '0')}`
  }

  const skip = (seconds: number) => {
    if (videoRef.current) {
      videoRef.current.currentTime += seconds
    }
  }

  // Loading state
  if (streamStatus === 'preparing') {
    return (
      <div className="relative w-full h-full bg-black flex items-center justify-center">
        <div className="text-center">
          <Loader2 className="w-12 h-12 text-[#1ed760] animate-spin mx-auto mb-4" />
          <p className="text-white text-lg mb-2">Preparing stream...</p>
          <p className="text-white/60 text-sm">{progress.toFixed(0)}% complete</p>
          <div className="w-64 h-1 bg-white/20 rounded-full mt-4 overflow-hidden">
            <div 
              className="h-full bg-[#1ed760] rounded-full transition-all"
              style={{ width: `${progress}%` }}
            />
          </div>
        </div>
      </div>
    )
  }

  // Error state
  if (streamStatus === 'error') {
    return (
      <div className="relative w-full h-full bg-black flex items-center justify-center">
        <div className="text-center">
          <p className="text-red-400 text-lg mb-2">Failed to start stream</p>
          <p className="text-white/60 text-sm mb-4">The video could not be transcoded</p>
          {onClose && (
            <button 
              onClick={onClose}
              className="px-4 py-2 bg-[#1f1f1f] text-white rounded-full hover:bg-[#2a2a2a] transition-colors"
            >
              Close
            </button>
          )}
        </div>
      </div>
    )
  }

  return (
    <div 
      className="relative w-full h-full bg-black group"
      onMouseMove={showControlsTemporarily}
      onClick={showControlsTemporarily}
    >
      {/* Video Element */}
      <video
        ref={videoRef}
        poster={poster}
        className="w-full h-full object-contain"
        onClick={togglePlay}
        playsInline
      />

      {/* Title Bar */}
      {showControls && title && (
        <div className="absolute top-0 left-0 right-0 p-4 bg-gradient-to-b from-black/80 to-transparent">
          <div className="flex items-center justify-between">
            <h3 className="text-white font-semibold text-lg">{title}</h3>
            {onClose && (
              <button 
                onClick={onClose}
                className="text-white/80 hover:text-white transition-colors"
              >
                ✕
              </button>
            )}
          </div>
        </div>
      )}

      {/* Center Play Button (when paused) */}
      {!isPlaying && streamStatus === 'ready' && (
        <div className="absolute inset-0 flex items-center justify-center">
          <button
            onClick={togglePlay}
            className="w-20 h-20 rounded-full bg-[#1ed760]/90 flex items-center justify-center hover:bg-[#1ed760] transition-colors"
          >
            <Play className="w-10 h-10 text-black ml-1" fill="currentColor" />
          </button>
        </div>
      )}

      {/* Controls Overlay */}
      {showControls && (
        <div className="absolute bottom-0 left-0 right-0 p-4 bg-gradient-to-t from-black/90 via-black/50 to-transparent">
          {/* Progress Bar */}
          <div className="relative mb-4">
            <input
              type="range"
              min={0}
              max={duration || 0}
              value={currentTime}
              onChange={handleSeek}
              className="w-full h-1 bg-white/20 rounded-full appearance-none cursor-pointer relative z-10
                [&::-webkit-slider-thumb]:appearance-none [&::-webkit-slider-thumb]:w-3 [&::-webkit-slider-thumb]:h-3 
                [&::-webkit-slider-thumb]:rounded-full [&::-webkit-slider-thumb]:bg-[#1ed760] [&::-webkit-slider-thumb]:cursor-pointer
                [&::-moz-range-thumb]:w-3 [&::-moz-range-thumb]:h-3 [&::-moz-range-thumb]:rounded-full 
                [&::-moz-range-thumb]:bg-[#1ed760] [&::-moz-range-thumb]:border-0 [&::-moz-range-thumb]:cursor-pointer"
              style={{
                background: `linear-gradient(to right, #1ed760 0%, #1ed760 ${(currentTime / duration) * 100}%, rgba(255,255,255,0.2) ${(currentTime / duration) * 100}%, rgba(255,255,255,0.2) 100%)`
              }}
            />
          </div>

          {/* Control Buttons */}
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              {/* Play/Pause */}
              <button
                onClick={togglePlay}
                className="text-white hover:text-[#1ed760] transition-colors"
              >
                {isPlaying ? (
                  <Pause className="w-6 h-6" />
                ) : (
                  <Play className="w-6 h-6" />
                )}
              </button>

              {/* Skip Back */}
              <button
                onClick={() => skip(-10)}
                className="text-white/70 hover:text-white transition-colors"
              >
                <SkipBack className="w-5 h-5" />
              </button>

              {/* Skip Forward */}
              <button
                onClick={() => skip(10)}
                className="text-white/70 hover:text-white transition-colors"
              >
                <SkipForward className="w-5 h-5" />
              </button>

              {/* Volume */}
              <div className="flex items-center gap-2 group/volume">
                <button
                  onClick={toggleMute}
                  className="text-white/70 hover:text-white transition-colors"
                >
                  {isMuted || volume === 0 ? (
                    <VolumeX className="w-5 h-5" />
                  ) : (
                    <Volume2 className="w-5 h-5" />
                  )}
                </button>
                <input
                  type="range"
                  min={0}
                  max={1}
                  step={0.1}
                  value={isMuted ? 0 : volume}
                  onChange={handleVolumeChange}
                  className="w-0 group-hover/volume:w-20 h-1 bg-white/20 rounded-full appearance-none cursor-pointer transition-all duration-200
                    [&::-webkit-slider-thumb]:appearance-none [&::-webkit-slider-thumb]:w-2.5 [&::-webkit-slider-thumb]:h-2.5 
                    [&::-webkit-slider-thumb]:rounded-full [&::-webkit-slider-thumb]:bg-white [&::-webkit-slider-thumb]:cursor-pointer
                    [&::-moz-range-thumb]:w-2.5 [&::-moz-range-thumb]:h-2.5 [&::-moz-range-thumb]:rounded-full 
                    [&::-moz-range-thumb]:bg-white [&::-moz-range-thumb]:border-0 [&::-moz-range-thumb]:cursor-pointer"
                  style={{
                    background: `linear-gradient(to right, white 0%, white ${(isMuted ? 0 : volume) * 100}%, rgba(255,255,255,0.2) ${(isMuted ? 0 : volume) * 100}%, rgba(255,255,255,0.2) 100%)`
                  }}
                />
              </div>

              {/* Time Display */}
              <span className="text-white/70 text-sm">
                {formatTime(currentTime)} / {formatTime(duration)}
              </span>
            </div>

            <div className="flex items-center gap-3">
              {/* Fullscreen */}
              <button
                onClick={toggleFullscreen}
                className="text-white/70 hover:text-white transition-colors"
              >
                <Maximize className="w-5 h-5" />
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
