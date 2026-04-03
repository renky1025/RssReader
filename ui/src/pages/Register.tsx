import { useState, useEffect } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { Rss, Check, X, Loader2 } from 'lucide-react'
import { useRegister, useCheckAvailability } from '../api/hooks'
import { useAppStore } from '../stores/appStore'

export default function Register() {
  const [username, setUsername] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [error, setError] = useState('')
  const [usernameAvailable, setUsernameAvailable] = useState<boolean | null>(null)
  const [emailAvailable, setEmailAvailable] = useState<boolean | null>(null)
  const navigate = useNavigate()
  const register = useRegister()
  const checkAvailability = useCheckAvailability()
  const setAuthenticated = useAppStore((state) => state.setAuthenticated)

  // Check username availability with debounce
  useEffect(() => {
    if (username.length < 3) {
      setUsernameAvailable(null)
      return
    }
    const timer = setTimeout(async () => {
      try {
        const result = await checkAvailability.mutateAsync({ username })
        setUsernameAvailable(result.username_available ?? null)
      } catch {
        setUsernameAvailable(null)
      }
    }, 500)
    return () => clearTimeout(timer)
  }, [username])

  // Check email availability with debounce
  useEffect(() => {
    if (!email || !email.includes('@')) {
      setEmailAvailable(null)
      return
    }
    const timer = setTimeout(async () => {
      try {
        const result = await checkAvailability.mutateAsync({ email })
        setEmailAvailable(result.email_available ?? null)
      } catch {
        setEmailAvailable(null)
      }
    }, 500)
    return () => clearTimeout(timer)
  }, [email])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    if (username.length < 3) {
      setError('Username must be at least 3 characters')
      return
    }

    if (usernameAvailable === false) {
      setError('Username is already taken')
      return
    }

    if (!email) {
      setError('Email is required')
      return
    }

    if (emailAvailable === false) {
      setError('Email is already registered')
      return
    }

    if (password.length < 6) {
      setError('Password must be at least 6 characters')
      return
    }

    if (password !== confirmPassword) {
      setError('Passwords do not match')
      return
    }

    try {
      await register.mutateAsync({ username, email, password })
      setAuthenticated(true)
      navigate('/')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Registration failed')
    }
  }

  return (
    <div className="min-h-screen bg-[var(--bg-primary)] flex items-center justify-center p-4">
      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center w-16 h-16 bg-[var(--accent)] rounded-2xl mb-4">
            <Rss className="w-8 h-8 text-white" />
          </div>
          <h1 className="text-3xl font-bold text-[var(--text-primary)]">Fread</h1>
          <p className="text-[var(--text-secondary)] mt-2">Create your account</p>
        </div>

        <form onSubmit={handleSubmit} className="bg-[var(--bg-secondary)] rounded-xl p-6 shadow-xl">
          {error && (
            <div className="bg-red-500/20 border border-red-500 text-red-400 px-4 py-2 rounded-lg mb-4">
              {error}
            </div>
          )}

          <div className="mb-4">
            <label className="block text-[var(--text-secondary)] text-sm font-medium mb-2">
              Username
            </label>
            <div className="relative">
              <input
                type="text"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                className={`w-full px-4 py-3 pr-10 bg-[var(--bg-primary)] border rounded-lg text-[var(--text-primary)] focus:outline-none transition-colors ${
                  usernameAvailable === false 
                    ? 'border-red-500 focus:border-red-500' 
                    : usernameAvailable === true 
                    ? 'border-green-500 focus:border-green-500' 
                    : 'border-[var(--border-color)] focus:border-[var(--accent)]'
                }`}
                placeholder="Choose a username"
                required
                minLength={3}
              />
              {username.length >= 3 && (
                <div className="absolute right-3 top-1/2 -translate-y-1/2">
                  {checkAvailability.isPending ? (
                    <Loader2 className="w-5 h-5 text-[var(--text-secondary)] animate-spin" />
                  ) : usernameAvailable === true ? (
                    <Check className="w-5 h-5 text-green-500" />
                  ) : usernameAvailable === false ? (
                    <X className="w-5 h-5 text-red-500" />
                  ) : null}
                </div>
              )}
            </div>
            {usernameAvailable === false && (
              <p className="text-red-400 text-xs mt-1">Username is already taken</p>
            )}
          </div>

          <div className="mb-4">
            <label className="block text-[var(--text-secondary)] text-sm font-medium mb-2">
              Email
            </label>
            <div className="relative">
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className={`w-full px-4 py-3 pr-10 bg-[var(--bg-primary)] border rounded-lg text-[var(--text-primary)] focus:outline-none transition-colors ${
                  emailAvailable === false 
                    ? 'border-red-500 focus:border-red-500' 
                    : emailAvailable === true 
                    ? 'border-green-500 focus:border-green-500' 
                    : 'border-[var(--border-color)] focus:border-[var(--accent)]'
                }`}
                placeholder="Enter your email"
                required
              />
              {email && email.includes('@') && (
                <div className="absolute right-3 top-1/2 -translate-y-1/2">
                  {checkAvailability.isPending ? (
                    <Loader2 className="w-5 h-5 text-[var(--text-secondary)] animate-spin" />
                  ) : emailAvailable === true ? (
                    <Check className="w-5 h-5 text-green-500" />
                  ) : emailAvailable === false ? (
                    <X className="w-5 h-5 text-red-500" />
                  ) : null}
                </div>
              )}
            </div>
            {emailAvailable === false && (
              <p className="text-red-400 text-xs mt-1">Email is already registered</p>
            )}
          </div>

          <div className="mb-4">
            <label className="block text-[var(--text-secondary)] text-sm font-medium mb-2">
              Password
            </label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full px-4 py-3 bg-[var(--bg-primary)] border border-[var(--border-color)] rounded-lg text-[var(--text-primary)] focus:outline-none focus:border-[var(--accent)] transition-colors"
              placeholder="Create a password"
              required
              minLength={6}
            />
          </div>

          <div className="mb-6">
            <label className="block text-[var(--text-secondary)] text-sm font-medium mb-2">
              Confirm Password
            </label>
            <input
              type="password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              className="w-full px-4 py-3 bg-[var(--bg-primary)] border border-[var(--border-color)] rounded-lg text-[var(--text-primary)] focus:outline-none focus:border-[var(--accent)] transition-colors"
              placeholder="Confirm your password"
              required
            />
          </div>

          <button
            type="submit"
            disabled={register.isPending}
            className="w-full py-3 bg-[var(--accent)] hover:opacity-90 text-white font-medium rounded-lg transition-colors disabled:opacity-50"
          >
            {register.isPending ? 'Creating account...' : 'Create Account'}
          </button>

          <p className="text-[var(--text-secondary)] text-sm text-center mt-4">
            Already have an account?{' '}
            <Link to="/login" className="text-[var(--accent)] hover:underline">
              Sign In
            </Link>
          </p>
        </form>
      </div>
    </div>
  )
}
