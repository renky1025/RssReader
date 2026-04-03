import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { Rss, Shield } from 'lucide-react'
import { useLogin, useCaptchaStatus } from '../api/hooks'
import { useAppStore } from '../stores/appStore'
import SliderCaptcha from '../components/SliderCaptcha'

export default function Login() {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [showCaptcha, setShowCaptcha] = useState(false)
  const [pendingLogin, setPendingLogin] = useState(false)
  
  const navigate = useNavigate()
  const login = useLogin()
  const setAuthenticated = useAppStore((state) => state.setAuthenticated)
  
  // Check captcha status when username changes
  const { data: captchaStatus, refetch: refetchCaptchaStatus } = useCaptchaStatus(username)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    // Check if captcha is required
    if (captchaStatus?.required) {
      setShowCaptcha(true)
      setPendingLogin(true)
      return
    }

    // Normal login without captcha
    await performLogin()
  }

  const performLogin = async (captchaVerified?: boolean) => {
    try {
      await login.mutateAsync({
        username,
        password,
        captcha_verified: captchaVerified,
      })
      setAuthenticated(true)
      navigate('/')
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Login failed'
      
      // Check if captcha is now required
      if (errorMessage.includes('captcha_required')) {
        setError('登录失败次数过多，请完成验证')
        refetchCaptchaStatus()
        setShowCaptcha(true)
      } else {
        setError(errorMessage === 'invalid credentials' ? '用户名或密码错误' : errorMessage)
        refetchCaptchaStatus()
      }
    }
  }

  const handleCaptchaSuccess = async () => {
    setShowCaptcha(false)
    
    if (pendingLogin) {
      setPendingLogin(false)
      await performLogin(true) // 验证码已通过
    }
  }

  const handleCaptchaClose = () => {
    setShowCaptcha(false)
    setPendingLogin(false)
  }

  return (
    <div className="min-h-screen bg-[var(--bg-primary)] flex items-center justify-center p-4">
      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center w-16 h-16 bg-[var(--accent)] rounded-2xl mb-4">
            <Rss className="w-8 h-8 text-white" />
          </div>
          <h1 className="text-3xl font-bold text-[var(--text-primary)]">Fread</h1>
          <p className="text-[var(--text-secondary)] mt-2">RSS Reader</p>
        </div>

        <form onSubmit={handleSubmit} className="bg-[var(--bg-secondary)] rounded-xl p-6 shadow-xl">
          {error && (
            <div className="bg-red-500/20 border border-red-500 text-red-400 px-4 py-2 rounded-lg mb-4">
              {error}
            </div>
          )}

          {/* Security indicator */}
          <div className="flex items-center gap-2 text-[var(--text-secondary)] text-xs mb-4 bg-[var(--bg-primary)] px-3 py-2 rounded-lg">
            <Shield className="w-4 h-4 text-green-500" />
            <span>密码已加密传输</span>
          </div>

          <div className="mb-4">
            <label className="block text-[var(--text-secondary)] text-sm font-medium mb-2">
              用户名
            </label>
            <input
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              className="w-full px-4 py-3 bg-[var(--bg-primary)] border border-[var(--border-color)] rounded-lg text-[var(--text-primary)] focus:outline-none focus:border-[var(--accent)] transition-colors"
              placeholder="请输入用户名"
              required
            />
          </div>

          <div className="mb-6">
            <label className="block text-[var(--text-secondary)] text-sm font-medium mb-2">
              密码
            </label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full px-4 py-3 bg-[var(--bg-primary)] border border-[var(--border-color)] rounded-lg text-[var(--text-primary)] focus:outline-none focus:border-[var(--accent)] transition-colors"
              placeholder="请输入密码"
              required
            />
          </div>

          {/* Show warning if captcha will be required */}
          {captchaStatus && captchaStatus.failed_attempts > 0 && captchaStatus.failed_attempts < 3 && (
            <div className="text-yellow-500 text-sm mb-4">
              ⚠️ 已失败 {captchaStatus.failed_attempts} 次，再失败 {3 - captchaStatus.failed_attempts} 次将需要验证
            </div>
          )}

          <button
            type="submit"
            disabled={login.isPending}
            className="w-full py-3 bg-[var(--accent)] hover:opacity-90 text-white font-medium rounded-lg transition-colors disabled:opacity-50"
          >
            {login.isPending ? '登录中...' : '登录'}
          </button>

          <p className="text-[var(--text-secondary)] text-sm text-center mt-4">
            还没有账号？{' '}
            <Link to="/register" className="text-[var(--accent)] hover:underline">
              注册
            </Link>
          </p>
        </form>
      </div>

      {/* Slider Captcha Modal */}
      {showCaptcha && (
        <SliderCaptcha
          isOpen={showCaptcha}
          onClose={handleCaptchaClose}
          onSuccess={handleCaptchaSuccess}
          onRefresh={() => {
            // react-puzzle-captcha 自带刷新功能
          }}
        />
      )}
    </div>
  )
}
