export interface User {
  id: number
  username: string
  email?: string
  is_admin: boolean
  status: number // 0=disabled, 1=active
  last_login_at?: string
  last_login_ip?: string
  last_login_device?: string
  onboarding_complete: boolean
  created_at: string
}

export interface RegisterRequest {
  username: string
  email: string
  password: string
}

export interface AdminStats {
  total_users: number
  active_users: number
  disabled_users: number
  admin_users: number
  total_feeds: number
  total_articles: number
  total_folders: number
  articles_today: number
  articles_this_week: number
  articles_this_month: number
}

export interface UserStats {
  user_id: number
  username: string
  feed_count: number
  article_count: number
  unread_count: number
  starred_count: number
}

export interface UpdateUserRequest {
  email?: string
  password?: string
  status?: number
  is_admin?: boolean
}

export interface CreateUserRequest {
  username: string
  email: string
  password: string
  is_admin: boolean
}

export interface Feed {
  id: number
  user_id: number
  username?: string
  folder_id?: number
  url: string
  title: string
  site_url?: string
  description?: string
  last_fetched?: number
  error_count: number
  last_error?: string
  disabled: boolean
  created_at: string
  unread_count?: number
}

export interface Article {
  id: number
  feed_id: number
  guid?: string
  url: string
  title: string
  author?: string
  content?: string
  summary?: string
  image_url?: string
  published_at: number
  is_read: boolean
  is_starred: boolean
  is_read_later: boolean
  created_at: string
  feed_title?: string
}

export interface Folder {
  id: number
  user_id: number
  name: string
  parent_id?: number
  created_at: string
  feed_count?: number
}

export interface Stats {
  total_feeds: number
  total_articles: number
  unread_articles: number
  starred_articles: number
  read_later_count: number
  total_folders: number
  disabled_feeds: number
  error_feeds: number
}

export interface PaginatedResponse<T> {
  items: T[]
  total: number
  limit: number
  offset: number
  has_more: boolean
}

export interface LoginRequest {
  username: string
  password: string
}

export interface SecureLoginRequest {
  username: string
  password: string
  encrypted?: boolean
  key_id?: string
  nonce?: string
  timestamp?: number
  captcha_verified?: boolean
}

export interface PublicKeyResponse {
  key_id: string
  public_key: string
  expires_at: number
}

export interface CaptchaStatusResponse {
  required: boolean
  failed_attempts: number
}

export interface CaptchaChallenge {
  token: string
  image_index: number
  target_x: number
}

export interface LoginResponse {
  token: string
  user: User
}

export interface RecommendedFeed {
  id: string
  name: string
  url: string
  description: string
  category: string
  icon: string
}

export interface AdminFeedListParams {
  user_id?: number
  disabled?: boolean
  q?: string
  limit?: number
  offset?: number
}

export interface AdminUpdateFeedRequest {
  user_id?: number
  folder_id?: number
  url?: string
  title?: string
  site_url?: string
  description?: string
  disabled?: boolean
  error_count?: number
  last_error?: string
}

export interface AdminCreateFeedRequest {
  user_id: number
  folder_id?: number
  url: string
  title?: string
  site_url?: string
  description?: string
}

export interface AdminSystemSettings {
  fetch_interval_minutes: number
}

export type ViewType = 'all' | 'unread' | 'starred' | 'read-later' | 'feed' | 'folder' | 'discover'

export interface ViewState {
  type: ViewType
  id?: number
  title: string
}
