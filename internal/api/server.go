package api

import (
	"net/http"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"rssreader/internal/config"
	"rssreader/internal/fetcher"
	"rssreader/internal/recommend"
	"rssreader/internal/security"
	"rssreader/internal/store"
)

type Server struct {
	config         *config.Config
	db             *store.DB
	fetcher        *fetcher.Fetcher
	router         *chi.Mux
	rsaManager     *security.RSAManager
	captchaManager *security.CaptchaManager
	loginLimiter   *security.LoginLimiter
	recommendMgr   *recommend.Manager
}

func NewServer(cfg *config.Config, db *store.DB, f *fetcher.Fetcher) *Server {
	s := &Server{
		config:         cfg,
		db:             db,
		fetcher:        f,
		router:         chi.NewRouter(),
		rsaManager:     security.NewRSAManager(24 * time.Hour),
		captchaManager: security.NewCaptchaManager(),
		loginLimiter:   security.NewLoginLimiter(3, 15*time.Minute), // 3 attempts in 15 minutes
		recommendMgr:   recommend.NewManager(filepath.Join("data", "recommended_feeds.json")),
	}

	// Generate initial RSA key pair
	s.rsaManager.GenerateKeyPair()

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	r := s.router

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Public routes
	r.Post("/api/v1/login", s.handleLogin)
	r.Post("/api/v1/register", s.handleRegister)
	r.Get("/api/v1/check-availability", s.handleCheckAvailability)
	r.Get("/api/v1/recommended-feeds", s.handleGetRecommendedFeeds)
	r.Get("/api/v1/proxy", s.handleImageProxy) // Image proxy for bypassing hotlink protection

	// Security routes (public)
	r.Get("/api/v1/auth/public-key", s.handleGetPublicKey)
	r.Get("/api/v1/auth/captcha-status", s.handleGetCaptchaStatus)
	r.Post("/api/v1/auth/captcha", s.handleGenerateCaptcha)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(s.AuthMiddleware)

		// User
		r.Get("/api/v1/me", s.handleGetCurrentUser)
		r.Post("/api/v1/onboarding/complete", s.handleCompleteOnboarding)

		// Feeds
		r.Get("/api/v1/feeds", s.handleGetFeeds)
		r.Post("/api/v1/feeds", s.handleCreateFeed)
		r.Post("/api/v1/feeds/batch", s.handleBatchCreateFeeds)
		r.Get("/api/v1/feeds/{id}", s.handleGetFeed)
		r.Patch("/api/v1/feeds/{id}", s.handleUpdateFeed)
		r.Delete("/api/v1/feeds/{id}", s.handleDeleteFeed)
		r.Post("/api/v1/feeds/{id}/fetch", s.handleFetchFeed)

		// Articles
		r.Get("/api/v1/articles", s.handleGetArticles)
		r.Get("/api/v1/articles/{id}", s.handleGetArticle)
		r.Patch("/api/v1/articles/{id}", s.handleUpdateArticle)
		r.Post("/api/v1/articles/mark-all-read", s.handleMarkAllRead)

		// Folders
		r.Get("/api/v1/folders", s.handleGetFolders)
		r.Post("/api/v1/folders", s.handleCreateFolder)
		r.Delete("/api/v1/folders/{id}", s.handleDeleteFolder)

		// Stats
		r.Get("/api/v1/stats", s.handleGetStats)

		// OPML
		r.Post("/api/v1/opml/import", s.handleImportOPML)
		r.Get("/api/v1/opml/export", s.handleExportOPML)

		// Admin routes (require admin privileges)
		r.Group(func(r chi.Router) {
			r.Use(s.AdminMiddleware)

			r.Get("/api/v1/admin/stats", s.handleAdminGetStats)
			r.Get("/api/v1/admin/users", s.handleAdminGetUsers)
			r.Post("/api/v1/admin/users", s.handleAdminCreateUser)
			r.Get("/api/v1/admin/users/{id}", s.handleAdminGetUser)
			r.Patch("/api/v1/admin/users/{id}", s.handleAdminUpdateUser)
			r.Delete("/api/v1/admin/users/{id}", s.handleAdminDeleteUser)
			r.Get("/api/v1/admin/users/{id}/stats", s.handleAdminGetUserStats)
			r.Get("/api/v1/admin/feeds", s.handleAdminGetFeeds)
			r.Post("/api/v1/admin/feeds", s.handleAdminCreateFeed)
			r.Patch("/api/v1/admin/feeds/{id}", s.handleAdminUpdateFeed)
			r.Delete("/api/v1/admin/feeds/{id}", s.handleAdminDeleteFeed)
			r.Get("/api/v1/admin/recommended-feeds", s.handleAdminGetRecommendedFeeds)
			r.Post("/api/v1/admin/recommended-feeds", s.handleAdminCreateRecommendedFeed)
			r.Patch("/api/v1/admin/recommended-feeds/{id}", s.handleAdminUpdateRecommendedFeed)
			r.Delete("/api/v1/admin/recommended-feeds/{id}", s.handleAdminDeleteRecommendedFeed)
			r.Get("/api/v1/admin/settings", s.handleAdminGetSettings)
			r.Patch("/api/v1/admin/settings", s.handleAdminUpdateSettings)
		})
	})

	// Serve static files for frontend
	fileServer := http.FileServer(http.Dir("./ui/dist"))
	r.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to serve the file, if not found serve index.html for SPA routing
		http.StripPrefix("/", fileServer).ServeHTTP(w, r)
	}))
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
