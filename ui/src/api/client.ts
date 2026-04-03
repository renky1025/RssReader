const API_BASE = import.meta.env.VITE_API_URL || '/api/v1'

// Unified API response format
interface APIResponse<T> {
  success: boolean
  data?: T
  error?: {
    code: number
    message: string
  }
}

class ApiClient {
  private token: string | null = null

  constructor() {
    this.token = localStorage.getItem('token')
  }

  setToken(token: string | null) {
    this.token = token
    if (token) {
      localStorage.setItem('token', token)
    } else {
      localStorage.removeItem('token')
    }
  }

  getToken() {
    return this.token
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    }

    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`
    }

    const response = await fetch(`${API_BASE}${endpoint}`, {
      ...options,
      headers,
    })

    if (response.status === 401) {
      this.setToken(null)
      window.location.href = '/login'
      throw new Error('Unauthorized')
    }

    if (response.status === 204) {
      return {} as T
    }

    const json: APIResponse<T> = await response.json().catch(() => ({
      success: false,
      error: { code: response.status, message: 'Request failed' }
    }))

    if (!json.success || json.error) {
      throw new Error(json.error?.message || 'Request failed')
    }

    return json.data as T
  }

  get<T>(endpoint: string) {
    return this.request<T>(endpoint)
  }

  post<T>(endpoint: string, data?: unknown) {
    return this.request<T>(endpoint, {
      method: 'POST',
      body: data ? JSON.stringify(data) : undefined,
    })
  }

  patch<T>(endpoint: string, data: unknown) {
    return this.request<T>(endpoint, {
      method: 'PATCH',
      body: JSON.stringify(data),
    })
  }

  delete<T>(endpoint: string) {
    return this.request<T>(endpoint, {
      method: 'DELETE',
    })
  }

  async uploadFile<T>(endpoint: string, file: File): Promise<T> {
    const formData = new FormData()
    formData.append('file', file)

    const headers: HeadersInit = {}
    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`
    }

    const response = await fetch(`${API_BASE}${endpoint}`, {
      method: 'POST',
      headers,
      body: formData,
    })

    const json: APIResponse<T> = await response.json().catch(() => ({
      success: false,
      error: { code: response.status, message: 'Upload failed' }
    }))

    if (!json.success || json.error) {
      throw new Error(json.error?.message || 'Upload failed')
    }

    return json.data as T
  }
}

export const apiClient = new ApiClient()

// Security-related API methods
export const securityApi = {
  getPublicKey: () => apiClient.get<{ public_key: string; key_id: string; expires_at: number }>('/auth/public-key'),
  getCaptchaStatus: (username: string) => apiClient.get<{ required: boolean; failed_attempts: number }>(`/auth/captcha-status?username=${encodeURIComponent(username)}`),
  generateCaptcha: () => apiClient.post<{ token: string; image_index: number; target_x: number }>('/auth/captcha'),
}
