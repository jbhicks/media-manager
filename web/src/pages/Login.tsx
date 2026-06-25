import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { motion } from 'framer-motion'
import { Film, LogIn, UserPlus } from 'lucide-react'
import { useAuth } from '@/contexts/AuthContext'

export function Login() {
  const [isRegistering, setIsRegistering] = useState(false)
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [email, setEmail] = useState('')
  const [error, setError] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const { login, register } = useAuth()
  const navigate = useNavigate()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setIsLoading(true)

    try {
      let success
      if (isRegistering) {
        success = await register(username, password, email)
      } else {
        success = await login(username, password)
      }

      if (success) {
        navigate('/')
      } else {
        setError(isRegistering ? 'Registration failed' : 'Invalid username or password')
      }
    } catch (err) {
      setError('An error occurred. Please try again.')
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-[#121212] flex items-center justify-center p-4">
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="w-full max-w-md"
      >
        {/* Logo */}
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center w-16 h-16 bg-[#1ed760] rounded-full mb-4">
            <Film className="w-8 h-8 text-black" />
          </div>
          <h1 className="text-3xl font-bold text-white">Media Manager</h1>
          <p className="text-[#b3b3b3] mt-2">Your personal entertainment center</p>
        </div>

        {/* Form */}
        <div className="bg-[#1f1f1f] rounded-lg p-8">
          <h2 className="text-2xl font-bold text-white mb-6">
            {isRegistering ? 'Create Account' : 'Sign In'}
          </h2>

          {error && (
            <div className="mb-4 p-3 bg-red-500/20 border border-red-500/50 rounded-lg text-red-400 text-sm">
              {error}
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-[#b3b3b3] mb-1">
                Username
              </label>
              <input
                type="text"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                className="w-full px-4 py-3 bg-[#121212] border border-[#4d4d4d] rounded-lg text-white placeholder-[#666] focus:outline-none focus:border-[#1ed760] transition-colors"
                placeholder="Enter your username"
                required
              />
            </div>

            {isRegistering && (
              <div>
                <label className="block text-sm font-medium text-[#b3b3b3] mb-1">
                  Email (optional)
                </label>
                <input
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  className="w-full px-4 py-3 bg-[#121212] border border-[#4d4d4d] rounded-lg text-white placeholder-[#666] focus:outline-none focus:border-[#1ed760] transition-colors"
                  placeholder="Enter your email"
                />
              </div>
            )}

            <div>
              <label className="block text-sm font-medium text-[#b3b3b3] mb-1">
                Password
              </label>
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="w-full px-4 py-3 bg-[#121212] border border-[#4d4d4d] rounded-lg text-white placeholder-[#666] focus:outline-none focus:border-[#1ed760] transition-colors"
                placeholder="Enter your password"
                required
              />
            </div>

            <button
              type="submit"
              disabled={isLoading}
              className="w-full flex items-center justify-center gap-2 px-6 py-3 bg-[#1ed760] text-black rounded-full font-semibold hover:bg-[#1ed760]/90 transition-colors disabled:opacity-50"
            >
              {isLoading ? (
                <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-black" />
              ) : isRegistering ? (
                <>
                  <UserPlus className="w-5 h-5" />
                  Create Account
                </>
              ) : (
                <>
                  <LogIn className="w-5 h-5" />
                  Sign In
                </>
              )}
            </button>
          </form>

          <div className="mt-6 text-center">
            <button
              onClick={() => {
                setIsRegistering(!isRegistering)
                setError('')
              }}
              className="text-[#1ed760] hover:underline text-sm"
            >
              {isRegistering
                ? 'Already have an account? Sign in'
                : "Don't have an account? Create one"}
            </button>
          </div>
        </div>
      </motion.div>
    </div>
  )
}
