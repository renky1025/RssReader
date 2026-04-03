import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { apiClient, securityApi } from './client'
import type { 
  Feed, Article, Folder, Stats, User,
  LoginResponse, PaginatedResponse, RecommendedFeed,
  RegisterRequest, AdminStats, UserStats, UpdateUserRequest, CreateUserRequest,
  AdminFeedListParams, AdminUpdateFeedRequest, AdminCreateFeedRequest,
  SecureLoginRequest, AdminSystemSettings
} from '../types'
import { encryptPassword, generateNonce, getTimestamp } from '../utils/crypto'

// Auth
export function useLogin() {
  return useMutation({
    mutationFn: async (data: SecureLoginRequest) => {
      // If encryption is requested, get public key and encrypt
      if (data.encrypted !== false) {
        try {
          const keyInfo = await securityApi.getPublicKey()
          const encryptedPassword = await encryptPassword(data.password, keyInfo.public_key)
          
          const secureRequest: SecureLoginRequest = {
            username: data.username,
            password: encryptedPassword,
            encrypted: true,
            key_id: keyInfo.key_id,
            nonce: data.nonce || generateNonce(),
            timestamp: data.timestamp || getTimestamp(),
            captcha_verified: data.captcha_verified,
          }
          
          return apiClient.post<LoginResponse>('/login', secureRequest)
        } catch (err) {
          // Fallback to plain text if encryption fails (for compatibility)
          console.warn('RSA encryption failed, falling back to plain text:', err)
        }
      }
      
      // Plain text login (fallback)
      return apiClient.post<LoginResponse>('/login', {
        username: data.username,
        password: data.password,
        encrypted: false,
        captcha_verified: data.captcha_verified,
      })
    },
    onSuccess: (data) => {
      apiClient.setToken(data.token)
    },
  })
}

// Security hooks
export function usePublicKey() {
  return useQuery({
    queryKey: ['publicKey'],
    queryFn: () => securityApi.getPublicKey(),
    staleTime: 60 * 60 * 1000, // 1 hour
  })
}

export function useCaptchaStatus(username: string) {
  return useQuery({
    queryKey: ['captchaStatus', username],
    queryFn: () => securityApi.getCaptchaStatus(username),
    enabled: !!username,
    staleTime: 0, // Always fresh
  })
}

export function useGenerateCaptcha() {
  return useMutation({
    mutationFn: () => securityApi.generateCaptcha(),
  })
}

export function useRegister() {
  return useMutation({
    mutationFn: (data: RegisterRequest) =>
      apiClient.post<LoginResponse>('/register', data),
    onSuccess: (data) => {
      apiClient.setToken(data.token)
    },
  })
}

export function useCheckAvailability() {
  return useMutation({
    mutationFn: (params: { username?: string; email?: string }) => {
      const searchParams = new URLSearchParams()
      if (params.username) searchParams.set('username', params.username)
      if (params.email) searchParams.set('email', params.email)
      return apiClient.get<{ username_available?: boolean; email_available?: boolean }>(
        `/check-availability?${searchParams.toString()}`
      )
    },
  })
}

export function useCurrentUser() {
  return useQuery({
    queryKey: ['currentUser'],
    queryFn: () => apiClient.get<User>('/me'),
    enabled: !!apiClient.getToken(),
  })
}

export function useCompleteOnboarding() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: () => apiClient.post('/onboarding/complete'),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['currentUser'] })
    },
  })
}

// Feeds
export function useFeeds() {
  return useQuery({
    queryKey: ['feeds'],
    queryFn: () => apiClient.get<Feed[]>('/feeds'),
  })
}

export function useCreateFeed() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: { url: string; folder_id?: number }) =>
      apiClient.post<Feed>('/feeds', data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['feeds'] })
      queryClient.invalidateQueries({ queryKey: ['stats'] })
    },
  })
}

export function useDeleteFeed() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => apiClient.delete(`/feeds/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['feeds'] })
      queryClient.invalidateQueries({ queryKey: ['articles'] })
      queryClient.invalidateQueries({ queryKey: ['stats'] })
    },
  })
}

export function useFetchFeed() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => apiClient.post(`/feeds/${id}/fetch`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['articles'] })
      queryClient.invalidateQueries({ queryKey: ['feeds'] })
    },
  })
}

export function useUpdateFeed() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, ...data }: { id: number; folder_id?: number | null; title?: string }) =>
      apiClient.patch<Feed>(`/feeds/${id}`, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['feeds'] })
      queryClient.invalidateQueries({ queryKey: ['stats'] })
    },
  })
}

// Articles
interface ArticleParams {
  feed_id?: number
  folder_id?: number
  is_read?: boolean
  is_starred?: boolean
  is_read_later?: boolean
  q?: string
  limit?: number
  offset?: number
}

export function useArticles(params: ArticleParams = {}) {
  const searchParams = new URLSearchParams()
  if (params.feed_id) searchParams.set('feed_id', String(params.feed_id))
  if (params.folder_id) searchParams.set('folder_id', String(params.folder_id))
  if (params.is_read !== undefined) searchParams.set('is_read', String(params.is_read))
  if (params.is_starred !== undefined) searchParams.set('is_starred', String(params.is_starred))
  if (params.is_read_later !== undefined) searchParams.set('is_read_later', String(params.is_read_later))
  if (params.q) searchParams.set('q', params.q)
  if (params.limit) searchParams.set('limit', String(params.limit))
  if (params.offset) searchParams.set('offset', String(params.offset))

  const queryString = searchParams.toString()
  
  return useQuery({
    queryKey: ['articles', params],
    queryFn: () => apiClient.get<PaginatedResponse<Article>>(`/articles${queryString ? `?${queryString}` : ''}`),
    // 优化缓存策略
    staleTime: params.is_read ? 5 * 60 * 1000 : 60 * 1000, // 已读文章缓存5分钟，未读文章缓存1分钟
    gcTime: 10 * 60 * 1000, // 10分钟后清理缓存
    // 对于特定feed的查询，可以更积极地缓存
    refetchOnWindowFocus: !params.feed_id || !params.is_read,
  })
}

export function useArticle(id: number) {
  return useQuery({
    queryKey: ['article', id],
    queryFn: () => apiClient.get<Article>(`/articles/${id}`),
    enabled: id > 0,
  })
}

export function useUpdateArticle() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, ...data }: { id: number; is_read?: boolean; is_starred?: boolean; is_read_later?: boolean }) =>
      apiClient.patch<Article>(`/articles/${id}`, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['articles'] })
      queryClient.invalidateQueries({ queryKey: ['article', variables.id] })
      queryClient.invalidateQueries({ queryKey: ['feeds'] })
      queryClient.invalidateQueries({ queryKey: ['stats'] })
    },
  })
}

export function useMarkAllRead() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (params: { feed_id?: number; folder_id?: number }) => {
      const searchParams = new URLSearchParams()
      if (params.feed_id) searchParams.set('feed_id', String(params.feed_id))
      if (params.folder_id) searchParams.set('folder_id', String(params.folder_id))
      const queryString = searchParams.toString()
      return apiClient.post(`/articles/mark-all-read${queryString ? `?${queryString}` : ''}`)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['articles'] })
      queryClient.invalidateQueries({ queryKey: ['feeds'] })
      queryClient.invalidateQueries({ queryKey: ['stats'] })
    },
  })
}

// Folders
export function useFolders() {
  return useQuery({
    queryKey: ['folders'],
    queryFn: () => apiClient.get<Folder[]>('/folders'),
  })
}

export function useCreateFolder() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: { name: string; parent_id?: number }) =>
      apiClient.post<Folder>('/folders', data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['folders'] })
    },
  })
}

export function useDeleteFolder() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => apiClient.delete(`/folders/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['folders'] })
      queryClient.invalidateQueries({ queryKey: ['feeds'] })
    },
  })
}

// Stats
export function useStats() {
  return useQuery({
    queryKey: ['stats'],
    queryFn: () => apiClient.get<Stats>('/stats'),
  })
}

// OPML
export function useImportOPML() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (file: File) => apiClient.uploadFile('/opml/import', file),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['feeds'] })
      queryClient.invalidateQueries({ queryKey: ['folders'] })
      queryClient.invalidateQueries({ queryKey: ['stats'] })
    },
  })
}

// Recommended Feeds
export function useRecommendedFeeds() {
  return useQuery({
    queryKey: ['recommendedFeeds'],
    queryFn: () => apiClient.get<RecommendedFeed[]>('/recommended-feeds'),
  })
}

export function useBatchCreateFeeds() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (urls: string[]) =>
      apiClient.post<{ created: Feed[]; errors: string[] }>('/feeds/batch', { urls }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['feeds'] })
      queryClient.invalidateQueries({ queryKey: ['stats'] })
    },
  })
}

// Admin hooks
export function useAdminStats() {
  return useQuery({
    queryKey: ['adminStats'],
    queryFn: () => apiClient.get<AdminStats>('/admin/stats'),
  })
}

export function useAdminUsers(params: { status?: number; q?: string; limit?: number; offset?: number } = {}) {
  const searchParams = new URLSearchParams()
  if (params.status !== undefined) searchParams.set('status', String(params.status))
  if (params.q) searchParams.set('q', params.q)
  if (params.limit) searchParams.set('limit', String(params.limit))
  if (params.offset) searchParams.set('offset', String(params.offset))
  const queryString = searchParams.toString()

  return useQuery({
    queryKey: ['adminUsers', params],
    queryFn: () => apiClient.get<PaginatedResponse<User>>(`/admin/users${queryString ? `?${queryString}` : ''}`),
  })
}

export function useAdminUser(id: number) {
  return useQuery({
    queryKey: ['adminUser', id],
    queryFn: () => apiClient.get<User>(`/admin/users/${id}`),
    enabled: id > 0,
  })
}

export function useAdminUserStats(id: number) {
  return useQuery({
    queryKey: ['adminUserStats', id],
    queryFn: () => apiClient.get<UserStats>(`/admin/users/${id}/stats`),
    enabled: id > 0,
  })
}

export function useAdminCreateUser() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: CreateUserRequest) =>
      apiClient.post<User>('/admin/users', data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['adminUsers'] })
      queryClient.invalidateQueries({ queryKey: ['adminStats'] })
    },
  })
}

export function useAdminUpdateUser() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, ...data }: { id: number } & UpdateUserRequest) =>
      apiClient.patch<User>(`/admin/users/${id}`, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['adminUsers'] })
      queryClient.invalidateQueries({ queryKey: ['adminUser', variables.id] })
      queryClient.invalidateQueries({ queryKey: ['adminStats'] })
    },
  })
}

export function useAdminDeleteUser() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => apiClient.delete(`/admin/users/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['adminUsers'] })
      queryClient.invalidateQueries({ queryKey: ['adminStats'] })
    },
  })
}

export function useAdminFeeds(params: AdminFeedListParams = {}) {
  const searchParams = new URLSearchParams()
  if (params.user_id !== undefined) searchParams.set('user_id', String(params.user_id))
  if (params.disabled !== undefined) searchParams.set('disabled', String(params.disabled))
  if (params.q) searchParams.set('q', params.q)
  if (params.limit) searchParams.set('limit', String(params.limit))
  if (params.offset) searchParams.set('offset', String(params.offset))
  const queryString = searchParams.toString()

  return useQuery({
    queryKey: ['adminFeeds', params],
    queryFn: () => apiClient.get<PaginatedResponse<Feed>>(`/admin/feeds${queryString ? `?${queryString}` : ''}`),
  })
}

export function useAdminCreateFeed() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: AdminCreateFeedRequest) =>
      apiClient.post<Feed>('/admin/feeds', data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['adminFeeds'] })
      queryClient.invalidateQueries({ queryKey: ['adminStats'] })
    },
  })
}

export function useAdminUpdateFeed() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, ...data }: { id: number } & AdminUpdateFeedRequest) =>
      apiClient.patch<Feed>(`/admin/feeds/${id}`, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['adminFeeds'] })
      queryClient.invalidateQueries({ queryKey: ['adminStats'] })
      queryClient.invalidateQueries({ queryKey: ['feed', variables.id] })
    },
  })
}

export function useAdminDeleteFeed() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => apiClient.delete(`/admin/feeds/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['adminFeeds'] })
      queryClient.invalidateQueries({ queryKey: ['adminStats'] })
    },
  })
}

export function useAdminRecommendedFeeds() {
  return useQuery({
    queryKey: ['adminRecommendedFeeds'],
    queryFn: () => apiClient.get<RecommendedFeed[]>('/admin/recommended-feeds'),
  })
}

export function useAdminCreateRecommendedFeed() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: Omit<RecommendedFeed, 'id'>) =>
      apiClient.post<RecommendedFeed>('/admin/recommended-feeds', data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['adminRecommendedFeeds'] })
      queryClient.invalidateQueries({ queryKey: ['recommendedFeeds'] })
    },
  })
}

export function useAdminUpdateRecommendedFeed() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: RecommendedFeed) =>
      apiClient.patch<RecommendedFeed>(`/admin/recommended-feeds/${data.id}`, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['adminRecommendedFeeds'] })
      queryClient.invalidateQueries({ queryKey: ['recommendedFeeds'] })
    },
  })
}

export function useAdminDeleteRecommendedFeed() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => apiClient.delete(`/admin/recommended-feeds/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['adminRecommendedFeeds'] })
      queryClient.invalidateQueries({ queryKey: ['recommendedFeeds'] })
    },
  })
}

export function useAdminSettings() {
  return useQuery({
    queryKey: ['adminSettings'],
    queryFn: () => apiClient.get<AdminSystemSettings>('/admin/settings'),
  })
}

export function useAdminUpdateSettings() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: Partial<AdminSystemSettings>) =>
      apiClient.patch<AdminSystemSettings>('/admin/settings', data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['adminSettings'] })
    },
  })
}
