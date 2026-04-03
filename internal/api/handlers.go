package api

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"rssreader/internal/auth"
	"rssreader/internal/models"
)

const fetchIntervalSettingKey = "fetch_interval_minutes"

// Auth handlers
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req models.SecureLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	clientIP := getClientIP(r)

	// Check if captcha is required
	if s.loginLimiter.RequiresCaptcha(req.Username) {
		// Verify captcha if required - 前端使用 react-puzzle-captcha 库验证
		// 只需要检查是否提供了 captcha_verified 标记
		if !req.CaptchaVerified {
			writeError(w, http.StatusPreconditionRequired, "captcha_required")
			return
		}
		// Clear attempts after successful captcha
		s.loginLimiter.ClearAttempts(req.Username)
	}

	// Validate timestamp (within 30 seconds)
	if req.Timestamp > 0 && !s.loginLimiter.ValidateTimestamp(req.Timestamp, 30*time.Second) {
		writeError(w, http.StatusBadRequest, "request expired")
		return
	}

	// Validate nonce (prevent replay attacks)
	if req.Nonce != "" && !s.loginLimiter.ValidateNonce(req.Nonce) {
		writeError(w, http.StatusBadRequest, "invalid nonce")
		return
	}

	// Decrypt password if encrypted
	password := req.Password
	if req.KeyID != "" && req.Encrypted {
		decrypted, err := s.rsaManager.DecryptPassword(req.KeyID, req.Password)
		if err != nil {
			log.Printf("Password decryption failed for user %s: %v", req.Username, err)
			writeError(w, http.StatusBadRequest, "decryption failed")
			return
		}
		password = decrypted
	}

	user, err := s.db.GetUserByUsername(req.Username)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	if user == nil || !auth.CheckPassword(password, user.PasswordHash) {
		// Record failed attempt
		s.loginLimiter.RecordAttempt(req.Username, false)

		// Check if captcha is now required
		if s.loginLimiter.RequiresCaptcha(req.Username) {
			writeError(w, http.StatusUnauthorized, "invalid credentials, captcha_required")
		} else {
			writeError(w, http.StatusUnauthorized, "invalid credentials")
		}
		return
	}

	// Check if user is disabled
	if user.Status == models.UserStatusDisabled {
		writeError(w, http.StatusForbidden, "account is disabled")
		return
	}

	// Record successful login
	s.loginLimiter.RecordAttempt(req.Username, true)

	token, err := auth.GenerateToken(user.ID, user.Username, user.IsAdmin, s.config.JWTSecret)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	// Get client IP and device info
	userAgent := r.Header.Get("User-Agent")

	// Update last login time, IP and device
	s.db.UpdateUserLastLogin(user.ID, clientIP, userAgent)

	// Refresh user data to include updated login info
	user, _ = s.db.GetUserByID(user.ID)

	writeJSON(w, http.StatusOK, models.LoginResponse{Token: token, User: user})
}

// handleGetPublicKey returns the current RSA public key for password encryption
func (s *Server) handleGetPublicKey(w http.ResponseWriter, r *http.Request) {
	keyPair, err := s.rsaManager.GetCurrentPublicKey()
	if err != nil {
		// Generate new key if none exists or expired
		keyPair, err = s.rsaManager.GenerateKeyPair()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to generate key")
			return
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"key_id":     keyPair.KeyID,
		"public_key": keyPair.PublicKey,
		"expires_at": keyPair.ExpiresAt.Unix(),
	})
}

// handleGetCaptchaStatus checks if captcha is required for a username
func (s *Server) handleGetCaptchaStatus(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		writeError(w, http.StatusBadRequest, "username required")
		return
	}

	required := s.loginLimiter.RequiresCaptcha(username)
	failedAttempts := s.loginLimiter.GetFailedAttempts(username)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"required":        required,
		"failed_attempts": failedAttempts,
	})
}

// handleGenerateCaptcha generates a new captcha challenge
func (s *Server) handleGenerateCaptcha(w http.ResponseWriter, r *http.Request) {
	challenge, err := s.captchaManager.GenerateChallenge()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate captcha")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"token":       challenge.Token,
		"image_index": challenge.ImageIndex,
		"target_x":    challenge.TargetX,
	})
}

// getClientIP extracts the real client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	// Remove port if present
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate input
	if req.Username == "" || len(req.Username) < 3 {
		writeError(w, http.StatusBadRequest, "username must be at least 3 characters")
		return
	}
	if req.Password == "" || len(req.Password) < 6 {
		writeError(w, http.StatusBadRequest, "password must be at least 6 characters")
		return
	}
	if req.Email == "" {
		writeError(w, http.StatusBadRequest, "email is required")
		return
	}

	// Check if username already exists
	existing, _ := s.db.GetUserByUsername(req.Username)
	if existing != nil {
		writeError(w, http.StatusConflict, "username already exists")
		return
	}

	// Check if email already exists
	existingEmail, _ := s.db.GetUserByEmail(req.Email)
	if existingEmail != nil {
		writeError(w, http.StatusConflict, "email already exists")
		return
	}

	// Hash password
	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	// Create user (not admin by default)
	user, err := s.db.CreateUser(req.Username, req.Email, passwordHash, false)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	// Generate token for auto-login
	token, err := auth.GenerateToken(user.ID, user.Username, user.IsAdmin, s.config.JWTSecret)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	writeJSON(w, http.StatusCreated, models.LoginResponse{Token: token, User: user})
}

func (s *Server) handleGetCurrentUser(w http.ResponseWriter, r *http.Request) {
	claims := getUserFromContext(r)
	user, err := s.db.GetUserByID(claims.UserID)
	if err != nil || user == nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (s *Server) handleCompleteOnboarding(w http.ResponseWriter, r *http.Request) {
	claims := getUserFromContext(r)
	if err := s.db.CompleteOnboarding(claims.UserID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to complete onboarding")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// Check if username or email is available
func (s *Server) handleCheckAvailability(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	username := query.Get("username")
	email := query.Get("email")

	result := map[string]interface{}{}

	if username != "" {
		existing, _ := s.db.GetUserByUsername(username)
		result["username_available"] = existing == nil
	}

	if email != "" {
		existing, _ := s.db.GetUserByEmail(email)
		result["email_available"] = existing == nil
	}

	writeJSON(w, http.StatusOK, result)
}

// Feed handlers
func (s *Server) handleGetFeeds(w http.ResponseWriter, r *http.Request) {
	claims := getUserFromContext(r)
	feeds, err := s.db.GetFeeds(claims.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get feeds")
		return
	}
	if feeds == nil {
		feeds = []*models.Feed{}
	}
	writeJSON(w, http.StatusOK, feeds)
}

func (s *Server) handleCreateFeed(w http.ResponseWriter, r *http.Request) {
	claims := getUserFromContext(r)
	var req models.CreateFeedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.URL == "" {
		writeError(w, http.StatusBadRequest, "url is required")
		return
	}

	// Check if feed already exists
	existing, _ := s.db.GetFeedByURL(claims.UserID, req.URL)
	if existing != nil {
		writeError(w, http.StatusConflict, "feed already exists")
		return
	}

	// Create the feed first
	feed, err := s.db.CreateFeed(claims.UserID, req.URL, "", "", "", req.FolderID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create feed")
		return
	}

	// Immediately fetch feed content to get metadata and articles
	if err := s.fetcher.FetchFeed(feed); err != nil {
		log.Printf("Warning: initial fetch failed for %s: %v", req.URL, err)
	}

	// Get updated feed info after fetch
	updatedFeed, _ := s.db.GetFeedByID(feed.ID)
	if updatedFeed != nil {
		feed = updatedFeed
	}

	writeJSON(w, http.StatusCreated, feed)
}

func (s *Server) handleGetFeed(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid feed id")
		return
	}

	feed, err := s.db.GetFeedByID(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get feed")
		return
	}
	if feed == nil {
		writeError(w, http.StatusNotFound, "feed not found")
		return
	}

	writeJSON(w, http.StatusOK, feed)
}

func (s *Server) handleDeleteFeed(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid feed id")
		return
	}

	if err := s.db.DeleteFeed(id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete feed")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleUpdateFeed(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid feed id")
		return
	}

	var req models.UpdateFeedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := s.db.UpdateFeed(id, req); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update feed")
		return
	}

	feed, _ := s.db.GetFeedByID(id)
	writeJSON(w, http.StatusOK, feed)
}

func (s *Server) handleFetchFeed(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid feed id")
		return
	}

	feed, err := s.db.GetFeedByID(id)
	if err != nil || feed == nil {
		writeError(w, http.StatusNotFound, "feed not found")
		return
	}

	go s.fetcher.FetchFeed(feed)
	writeJSON(w, http.StatusOK, map[string]string{"message": "fetch started"})
}

// Article handlers
func (s *Server) handleGetArticles(w http.ResponseWriter, r *http.Request) {
	claims := getUserFromContext(r)
	query := r.URL.Query()

	params := models.ArticleListParams{
		Query:  query.Get("q"),
		Limit:  50,
		Offset: 0,
	}

	if feedID := query.Get("feed_id"); feedID != "" {
		if id, err := strconv.ParseInt(feedID, 10, 64); err == nil {
			params.FeedID = &id
		}
	}
	if folderID := query.Get("folder_id"); folderID != "" {
		if id, err := strconv.ParseInt(folderID, 10, 64); err == nil {
			params.FolderID = &id
		}
	}
	if isRead := query.Get("is_read"); isRead != "" {
		b := isRead == "true"
		params.IsRead = &b
	}
	if isStarred := query.Get("is_starred"); isStarred != "" {
		b := isStarred == "true"
		params.IsStarred = &b
	}
	if isReadLater := query.Get("is_read_later"); isReadLater != "" {
		b := isReadLater == "true"
		params.IsReadLater = &b
	}
	if limit := query.Get("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 && l <= 100 {
			params.Limit = l
		}
	}
	if offset := query.Get("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil && o >= 0 {
			params.Offset = o
		}
	}

	// 添加缓存头部，对于已读文章可以缓存更长时间
	if params.IsRead != nil && *params.IsRead {
		w.Header().Set("Cache-Control", "public, max-age=300") // 5分钟缓存
	} else {
		w.Header().Set("Cache-Control", "public, max-age=60") // 1分钟缓存
	}

	articles, total, err := s.db.GetArticles(claims.UserID, params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get articles")
		return
	}

	if articles == nil {
		articles = []*models.Article{}
	}

	// 优化：对于分页查询，如果total为-1，计算hasMore的方式
	hasMore := false
	if total == -1 {
		// 如果返回的文章数等于limit，可能还有更多
		hasMore = len(articles) == params.Limit
		total = params.Offset + len(articles) // 设置一个估算值
		if hasMore {
			total++ // 表示至少还有一个
		}
	} else {
		hasMore = params.Offset+len(articles) < total
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Items:   articles,
		Total:   total,
		Limit:   params.Limit,
		Offset:  params.Offset,
		HasMore: hasMore,
	})
}

func (s *Server) handleGetArticle(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid article id")
		return
	}

	article, err := s.db.GetArticleByID(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get article")
		return
	}
	if article == nil {
		writeError(w, http.StatusNotFound, "article not found")
		return
	}

	writeJSON(w, http.StatusOK, article)
}

func (s *Server) handleUpdateArticle(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid article id")
		return
	}

	var req models.UpdateArticleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := s.db.UpdateArticle(id, req); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update article")
		return
	}

	article, _ := s.db.GetArticleByID(id)
	writeJSON(w, http.StatusOK, article)
}

func (s *Server) handleMarkAllRead(w http.ResponseWriter, r *http.Request) {
	claims := getUserFromContext(r)
	query := r.URL.Query()

	var feedID, folderID *int64
	if fid := query.Get("feed_id"); fid != "" {
		if id, err := strconv.ParseInt(fid, 10, 64); err == nil {
			feedID = &id
		}
	}
	if fid := query.Get("folder_id"); fid != "" {
		if id, err := strconv.ParseInt(fid, 10, 64); err == nil {
			folderID = &id
		}
	}

	if err := s.db.MarkAllAsRead(claims.UserID, feedID, folderID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to mark as read")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "marked as read"})
}

// Folder handlers
func (s *Server) handleGetFolders(w http.ResponseWriter, r *http.Request) {
	claims := getUserFromContext(r)
	folders, err := s.db.GetFolders(claims.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get folders")
		return
	}
	if folders == nil {
		folders = []*models.Folder{}
	}
	writeJSON(w, http.StatusOK, folders)
}

func (s *Server) handleCreateFolder(w http.ResponseWriter, r *http.Request) {
	claims := getUserFromContext(r)
	var req models.CreateFolderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	folder, err := s.db.CreateFolder(claims.UserID, req.Name, req.ParentID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create folder")
		return
	}

	writeJSON(w, http.StatusCreated, folder)
}

func (s *Server) handleDeleteFolder(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid folder id")
		return
	}

	if err := s.db.DeleteFolder(id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete folder")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Stats handler
func (s *Server) handleGetStats(w http.ResponseWriter, r *http.Request) {
	claims := getUserFromContext(r)
	stats, err := s.db.GetStats(claims.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get stats")
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

// Helper functions - Unified API response format
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	response := models.APIResponse{
		Success: status >= 200 && status < 300,
		Data:    data,
	}
	json.NewEncoder(w).Encode(response)
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	response := models.APIResponse{
		Success: false,
		Error: &models.APIError{
			Code:    status,
			Message: message,
		},
	}
	json.NewEncoder(w).Encode(response)
}

// Admin handlers

func (s *Server) handleAdminGetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := s.db.GetAdminStats()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get admin stats")
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

func (s *Server) handleAdminGetUsers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	params := models.AdminUserListParams{
		Query:  query.Get("q"),
		Limit:  50,
		Offset: 0,
	}

	if status := query.Get("status"); status != "" {
		s, err := strconv.Atoi(status)
		if err == nil {
			params.Status = &s
		}
	}
	if limit := query.Get("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			params.Limit = l
		}
	}
	if offset := query.Get("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil {
			params.Offset = o
		}
	}

	users, total, err := s.db.GetAllUsers(params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get users")
		return
	}

	if users == nil {
		users = []*models.User{}
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Items:   users,
		Total:   total,
		Limit:   params.Limit,
		Offset:  params.Offset,
		HasMore: params.Offset+len(users) < total,
	})
}

func (s *Server) handleAdminGetUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	user, err := s.db.GetUserByID(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get user")
		return
	}
	if user == nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func (s *Server) handleAdminUpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	var req models.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// If password is being updated, hash it
	if req.Password != nil && *req.Password != "" {
		hash, err := auth.HashPassword(*req.Password)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to hash password")
			return
		}
		req.Password = &hash
	}

	if err := s.db.UpdateUser(id, req); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update user")
		return
	}

	user, _ := s.db.GetUserByID(id)
	writeJSON(w, http.StatusOK, user)
}

func (s *Server) handleAdminDeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	// Prevent deleting self
	claims := getUserFromContext(r)
	if claims.UserID == id {
		writeError(w, http.StatusBadRequest, "cannot delete yourself")
		return
	}

	if err := s.db.DeleteUser(id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete user")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleAdminGetUserStats(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	stats, err := s.db.GetUserStats(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get user stats")
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

func (s *Server) handleAdminCreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
		IsAdmin  bool   `json:"is_admin"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate input
	if req.Username == "" || len(req.Username) < 3 {
		writeError(w, http.StatusBadRequest, "username must be at least 3 characters")
		return
	}
	if req.Password == "" || len(req.Password) < 6 {
		writeError(w, http.StatusBadRequest, "password must be at least 6 characters")
		return
	}

	// Check if username already exists
	existing, _ := s.db.GetUserByUsername(req.Username)
	if existing != nil {
		writeError(w, http.StatusConflict, "username already exists")
		return
	}

	// Hash password
	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	// Create user
	user, err := s.db.CreateUser(req.Username, req.Email, passwordHash, req.IsAdmin)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	writeJSON(w, http.StatusCreated, user)
}

func (s *Server) handleAdminGetSettings(w http.ResponseWriter, r *http.Request) {
	interval := int(s.fetcher.SchedulerInterval().Minutes())
	if interval <= 0 {
		interval = s.config.FetchInterval
	}

	if raw, ok, err := s.db.GetAppSetting(fetchIntervalSettingKey); err == nil && ok {
		if v, convErr := strconv.Atoi(raw); convErr == nil && v > 0 {
			interval = v
		}
	}

	writeJSON(w, http.StatusOK, models.AdminSystemSettings{
		FetchIntervalMinutes: interval,
	})
}

func (s *Server) handleAdminUpdateSettings(w http.ResponseWriter, r *http.Request) {
	var req models.UpdateAdminSystemSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.FetchIntervalMinutes == nil {
		writeError(w, http.StatusBadRequest, "fetch_interval_minutes is required")
		return
	}
	if *req.FetchIntervalMinutes < 1 || *req.FetchIntervalMinutes > 1440 {
		writeError(w, http.StatusBadRequest, "fetch_interval_minutes must be between 1 and 1440")
		return
	}

	if err := s.db.SetAppSetting(fetchIntervalSettingKey, strconv.Itoa(*req.FetchIntervalMinutes)); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save settings")
		return
	}

	s.fetcher.SetSchedulerInterval(time.Duration(*req.FetchIntervalMinutes) * time.Minute)
	writeJSON(w, http.StatusOK, models.AdminSystemSettings{
		FetchIntervalMinutes: *req.FetchIntervalMinutes,
	})
}

func (s *Server) handleAdminGetFeeds(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	params := models.AdminFeedListParams{
		Query:  query.Get("q"),
		Limit:  50,
		Offset: 0,
	}

	if userID := query.Get("user_id"); userID != "" {
		if id, err := strconv.ParseInt(userID, 10, 64); err == nil {
			params.UserID = &id
		}
	}
	if disabled := query.Get("disabled"); disabled != "" {
		if d, err := strconv.ParseBool(disabled); err == nil {
			params.Disabled = &d
		}
	}
	if limit := query.Get("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			params.Limit = l
		}
	}
	if offset := query.Get("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil {
			params.Offset = o
		}
	}

	feeds, total, err := s.db.GetAdminFeeds(params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get feeds")
		return
	}
	if feeds == nil {
		feeds = []*models.Feed{}
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Items:   feeds,
		Total:   total,
		Limit:   params.Limit,
		Offset:  params.Offset,
		HasMore: params.Offset+len(feeds) < total,
	})
}

func (s *Server) handleAdminCreateFeed(w http.ResponseWriter, r *http.Request) {
	var req models.AdminCreateFeedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.UserID <= 0 || strings.TrimSpace(req.URL) == "" {
		writeError(w, http.StatusBadRequest, "user_id and url are required")
		return
	}

	feed, err := s.db.AdminCreateFeed(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to create feed")
		return
	}
	writeJSON(w, http.StatusCreated, feed)
}

func (s *Server) handleAdminUpdateFeed(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid feed id")
		return
	}

	var req models.AdminUpdateFeedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := s.db.AdminUpdateFeed(id, req); err != nil {
		writeError(w, http.StatusBadRequest, "failed to update feed")
		return
	}
	feed, err := s.db.GetFeedByID(id)
	if err != nil || feed == nil {
		writeError(w, http.StatusInternalServerError, "failed to get updated feed")
		return
	}
	writeJSON(w, http.StatusOK, feed)
}

func (s *Server) handleAdminDeleteFeed(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid feed id")
		return
	}

	if err := s.db.AdminDeleteFeed(id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete feed")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Image proxy handler - proxies external images to avoid 403 errors from hotlink protection
func (s *Server) handleImageProxy(w http.ResponseWriter, r *http.Request) {
	imageURL := r.URL.Query().Get("url")
	if imageURL == "" {
		http.Error(w, "url parameter is required", http.StatusBadRequest)
		return
	}

	// Decode HTML entities (e.g., &amp; -> &)
	imageURL = strings.ReplaceAll(imageURL, "&amp;", "&")

	// Validate URL
	parsedURL, err := url.Parse(imageURL)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		http.Error(w, "invalid url", http.StatusBadRequest)
		return
	}

	// Generate possible Referer values to try
	// For CDN subdomains like image.xxx.com, static.xxx.com, cdn.xxx.com,
	// the actual referring site is usually www.xxx.com or xxx.com
	referers := generateReferers(parsedURL)

	// Try each referer until one works
	var resp *http.Response
	var lastErr error
	for _, referer := range referers {
		resp, lastErr = fetchImageWithReferer(imageURL, referer)
		if lastErr == nil && resp != nil && resp.StatusCode == http.StatusOK {
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
	}

	if lastErr != nil {
		log.Printf("Image proxy error for %s: %v", imageURL, lastErr)
		http.Error(w, "failed to fetch image", http.StatusBadGateway)
		return
	}

	if resp == nil || resp.StatusCode != http.StatusOK {
		statusCode := http.StatusBadGateway
		if resp != nil {
			statusCode = resp.StatusCode
			resp.Body.Close()
		}
		log.Printf("Image proxy upstream error for %s: status %d (tried referers: %v)", imageURL, statusCode, referers)
		http.Error(w, "upstream error", statusCode)
		return
	}
	defer resp.Body.Close()

	// Get content type
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Only proxy image content types (also allow application/octet-stream as some servers use it)
	if !strings.HasPrefix(contentType, "image/") && contentType != "application/octet-stream" {
		log.Printf("Image proxy: not an image for %s, got content-type: %s", imageURL, contentType)
		http.Error(w, "not an image", http.StatusBadRequest)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=86400") // Cache for 24 hours
	w.Header().Set("Access-Control-Allow-Origin", "*")       // Allow CORS
	if contentLength := resp.Header.Get("Content-Length"); contentLength != "" {
		w.Header().Set("Content-Length", contentLength)
	}

	// Copy the image data to response
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("Image proxy copy error: %v", err)
	}
}

// generateReferers creates a list of possible Referer values to try
func generateReferers(parsedURL *url.URL) []string {
	host := parsedURL.Host
	scheme := parsedURL.Scheme

	// Extract the base domain (remove common CDN subdomains)
	baseDomain := host
	cdnPrefixes := []string{"image.", "img.", "images.", "static.", "cdn.", "assets.", "media.", "pics.", "i.", "external-preview."}
	for _, prefix := range cdnPrefixes {
		if strings.HasPrefix(host, prefix) {
			baseDomain = strings.TrimPrefix(host, prefix)
			break
		}
	}

	referers := []string{}

	// If we found a CDN subdomain, try the main site first
	if baseDomain != host {
		referers = append(referers, scheme+"://www."+baseDomain+"/")
		referers = append(referers, scheme+"://"+baseDomain+"/")
	}

	// Then try the image host itself
	referers = append(referers, scheme+"://"+host+"/")

	// Finally try with no referer (some servers allow this)
	referers = append(referers, "")

	return referers
}

// fetchImageWithReferer fetches an image with the specified Referer header
func fetchImageWithReferer(imageURL, referer string) (*http.Response, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return http.ErrUseLastResponse
			}
			req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
			return nil
		},
	}

	req, err := http.NewRequest("GET", imageURL, nil)
	if err != nil {
		return nil, err
	}

	// Set comprehensive headers to mimic a real browser request
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "image/avif,image/webp,image/apng,image/svg+xml,image/*,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7")
	req.Header.Set("Accept-Encoding", "identity")
	req.Header.Set("Sec-Ch-Ua", `"Not_A Brand";v="8", "Chromium";v="120", "Google Chrome";v="120"`)
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)
	req.Header.Set("Sec-Fetch-Dest", "image")
	req.Header.Set("Sec-Fetch-Mode", "no-cors")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")

	if referer != "" {
		req.Header.Set("Referer", referer)
		// Extract origin from referer
		if u, err := url.Parse(referer); err == nil {
			req.Header.Set("Origin", u.Scheme+"://"+u.Host)
			req.Header.Set("Sec-Fetch-Site", "cross-site")
		}
	} else {
		// No referer - pretend it's a direct navigation
		req.Header.Set("Sec-Fetch-Site", "none")
	}

	return client.Do(req)
}
