declare module 'react-puzzle-captcha' {
  import { FC } from 'react'

  interface VerifyProps {
    width?: number
    height?: number
    visible?: boolean
    onSuccess?: () => void
    onFail?: () => void
    onRefresh?: () => void
    imgUrl?: string
    text?: {
      loading?: string
      slideTip?: string
      success?: string
      fail?: string
    }
  }

  export const Verify: FC<VerifyProps>
}
