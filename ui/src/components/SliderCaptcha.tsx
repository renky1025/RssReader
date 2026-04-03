import { Verify } from 'react-puzzle-captcha'
import 'react-puzzle-captcha/dist/react-puzzle-captcha.css'

interface SliderCaptchaProps {
  isOpen: boolean
  onClose: () => void
  onSuccess: () => void
  onRefresh?: () => void
}

export default function SliderCaptcha({
  isOpen,
  onClose,
  onSuccess,
  onRefresh,
}: SliderCaptchaProps) {
  const handleSuccess = () => {
    // 验证成功，调用回调
    onSuccess()
    onClose()
  }

  const handleFail = () => {
    // 验证失败，可以刷新
    if (onRefresh) {
      onRefresh()
    }
  }

  const handleRefresh = () => {
    if (onRefresh) {
      onRefresh()
    }
  }

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div className="bg-[var(--bg-secondary)] rounded-xl shadow-2xl overflow-hidden">
        {/* Header */}
        <div className="flex items-center justify-between px-4 py-3 border-b border-[var(--border-color)]">
          <span className="text-[var(--text-primary)] font-medium">安全验证</span>
          <button
            onClick={onClose}
            className="text-[var(--text-secondary)] hover:text-[var(--text-primary)] transition-colors text-xl leading-none"
          >
            ×
          </button>
        </div>

        {/* Verify 组件 */}
        <div className="p-4">
          <Verify
            width={320}
            height={160}
            visible={true}
            onSuccess={handleSuccess}
            onFail={handleFail}
            onRefresh={handleRefresh}
          />
        </div>
      </div>
    </div>
  )
}
