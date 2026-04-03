package models

import "time"

type User struct {
	ID                 int64      `json:"id"`
	Username           string     `json:"username"`
	Email              string     `json:"email,omitempty"`
	PasswordHash       string     `json:"-"`
	IsAdmin            bool       `json:"is_admin"`
	Status             int        `json:"status"` // 1=active, 0=disabled
	LastLoginAt        *time.Time `json:"last_login_at,omitempty"`
	LastLoginIP        string     `json:"last_login_ip,omitempty"`
	LastLoginDevice    string     `json:"last_login_device,omitempty"`
	OnboardingComplete bool       `json:"onboarding_complete"`
	CreatedAt          time.Time  `json:"created_at"`
}

// User status constants
const (
	UserStatusDisabled = 0
	UserStatusActive   = 1
)

type Folder struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Name      string    `json:"name"`
	ParentID  *int64    `json:"parent_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	// Computed fields
	FeedCount int `json:"feed_count,omitempty"`
}

type Feed struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	FolderID     *int64    `json:"folder_id,omitempty"`
	URL          string    `json:"url"`
	Title        string    `json:"title"`
	SiteURL      string    `json:"site_url,omitempty"`
	Description  string    `json:"description,omitempty"`
	LastFetched  *int64    `json:"last_fetched,omitempty"`
	ETag         string    `json:"-"`
	LastModified string    `json:"-"`
	ErrorCount   int       `json:"error_count"`
	LastError    string    `json:"last_error,omitempty"`
	Disabled     bool      `json:"disabled"`
	CreatedAt    time.Time `json:"created_at"`
	// Computed fields
	UnreadCount int    `json:"unread_count,omitempty"`
	Username    string `json:"username,omitempty"`
}

type Article struct {
	ID          int64     `json:"id"`
	FeedID      int64     `json:"feed_id"`
	GUID        string    `json:"guid,omitempty"`
	URL         string    `json:"url"`
	Title       string    `json:"title"`
	Author      string    `json:"author,omitempty"`
	Content     string    `json:"content,omitempty"`
	Summary     string    `json:"summary,omitempty"`
	ImageURL    string    `json:"image_url,omitempty"`
	PublishedAt int64     `json:"published_at"`
	IsRead      bool      `json:"is_read"`
	IsStarred   bool      `json:"is_starred"`
	IsReadLater bool      `json:"is_read_later"`
	CreatedAt   time.Time `json:"created_at"`
	// Joined fields
	FeedTitle string `json:"feed_title,omitempty"`
}

type Tag struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type SavedLink struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	URL         string    `json:"url"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// Security-related models
type LoginAttempt struct {
	ID          int64     `json:"id"`
	Username    string    `json:"username"`
	IPAddress   string    `json:"ip_address"`
	Success     bool      `json:"success"`
	AttemptedAt time.Time `json:"attempted_at"`
	UserAgent   string    `json:"user_agent,omitempty"`
}

type RSAKey struct {
	ID         int64     `json:"id"`
	SessionID  string    `json:"session_id"`
	PublicKey  string    `json:"public_key"`
	PrivateKey string    `json:"-"` // Never expose private key in JSON
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	IsActive   bool      `json:"is_active"`
}

// API Request/Response types
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// SecureLoginRequest includes RSA encryption and captcha support
type SecureLoginRequest struct {
	Username        string `json:"username"`
	Password        string `json:"password"`         // Base64 RSA encrypted or plain text
	Encrypted       bool   `json:"encrypted"`        // Whether password is RSA encrypted
	KeyID           string `json:"key_id"`           // RSA key ID used for encryption
	Nonce           string `json:"nonce"`            // Random string for replay protection
	Timestamp       int64  `json:"timestamp"`        // Request timestamp
	CaptchaVerified bool   `json:"captcha_verified"` // Whether captcha was verified on frontend
}

type GetPublicKeyResponse struct {
	PublicKey string `json:"public_key"`
	SessionID string `json:"session_id"`
}

type CaptchaChallenge struct {
	Required   bool   `json:"required"`
	Challenge  string `json:"challenge,omitempty"`
	ImageIndex int    `json:"image_index,omitempty"` // Index for selecting background image
}

type LoginResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}

type CreateFeedRequest struct {
	URL      string `json:"url"`
	FolderID *int64 `json:"folder_id,omitempty"`
}

type UpdateFeedRequest struct {
	Title    *string `json:"title,omitempty"`
	FolderID *int64  `json:"folder_id,omitempty"`
}

type CreateFolderRequest struct {
	Name     string `json:"name"`
	ParentID *int64 `json:"parent_id,omitempty"`
}

type UpdateArticleRequest struct {
	IsRead      *bool `json:"is_read,omitempty"`
	IsStarred   *bool `json:"is_starred,omitempty"`
	IsReadLater *bool `json:"is_read_later,omitempty"`
}

type ArticleListParams struct {
	FeedID      *int64
	FolderID    *int64
	IsRead      *bool
	IsStarred   *bool
	IsReadLater *bool
	Query       string
	Limit       int
	Offset      int
}

type PaginatedResponse struct {
	Items   interface{} `json:"items"`
	Total   int         `json:"total"`
	Limit   int         `json:"limit"`
	Offset  int         `json:"offset"`
	HasMore bool        `json:"has_more"`
}

// Unified API response format
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Deprecated: Use APIResponse instead
type ErrorResponse struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}

type Stats struct {
	TotalFeeds      int `json:"total_feeds"`
	TotalArticles   int `json:"total_articles"`
	UnreadArticles  int `json:"unread_articles"`
	StarredArticles int `json:"starred_articles"`
	ReadLaterCount  int `json:"read_later_count"`
	TotalFolders    int `json:"total_folders"`
	DisabledFeeds   int `json:"disabled_feeds"`
	ErrorFeeds      int `json:"error_feeds"`
}

// Registration request
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Admin: User management
type AdminUserListParams struct {
	Status *int
	Query  string
	Limit  int
	Offset int
}

type UpdateUserRequest struct {
	Email    *string `json:"email,omitempty"`
	Password *string `json:"password,omitempty"`
	Status   *int    `json:"status,omitempty"`
	IsAdmin  *bool   `json:"is_admin,omitempty"`
}

// Admin: Global statistics
type AdminStats struct {
	TotalUsers        int `json:"total_users"`
	ActiveUsers       int `json:"active_users"`
	DisabledUsers     int `json:"disabled_users"`
	AdminUsers        int `json:"admin_users"`
	TotalFeeds        int `json:"total_feeds"`
	TotalArticles     int `json:"total_articles"`
	TotalFolders      int `json:"total_folders"`
	ArticlesToday     int `json:"articles_today"`
	ArticlesThisWeek  int `json:"articles_this_week"`
	ArticlesThisMonth int `json:"articles_this_month"`
}

// User stats for admin view
type UserStats struct {
	UserID       int64  `json:"user_id"`
	Username     string `json:"username"`
	FeedCount    int    `json:"feed_count"`
	ArticleCount int    `json:"article_count"`
	UnreadCount  int    `json:"unread_count"`
	StarredCount int    `json:"starred_count"`
}

// Admin feed list filters
type AdminFeedListParams struct {
	UserID   *int64
	Disabled *bool
	Query    string
	Limit    int
	Offset   int
}

// Admin feed update request
type AdminUpdateFeedRequest struct {
	UserID      *int64  `json:"user_id,omitempty"`
	FolderID    *int64  `json:"folder_id,omitempty"`
	URL         *string `json:"url,omitempty"`
	Title       *string `json:"title,omitempty"`
	SiteURL     *string `json:"site_url,omitempty"`
	Description *string `json:"description,omitempty"`
	Disabled    *bool   `json:"disabled,omitempty"`
	ErrorCount  *int    `json:"error_count,omitempty"`
	LastError   *string `json:"last_error,omitempty"`
}

type AdminCreateFeedRequest struct {
	UserID      int64  `json:"user_id"`
	FolderID    *int64 `json:"folder_id,omitempty"`
	URL         string `json:"url"`
	Title       string `json:"title"`
	SiteURL     string `json:"site_url"`
	Description string `json:"description"`
}

// 推荐订阅源
type RecommendedFeed struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Icon        string `json:"icon"`
}

type AdminSystemSettings struct {
	FetchIntervalMinutes int `json:"fetch_interval_minutes"`
}

type UpdateAdminSystemSettingsRequest struct {
	FetchIntervalMinutes *int `json:"fetch_interval_minutes,omitempty"`
}
