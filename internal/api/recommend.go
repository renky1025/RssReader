package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"rssreader/internal/models"
)

// handleGetRecommendedFeeds returns a list of recommended feeds
func (s *Server) handleGetRecommendedFeeds(w http.ResponseWriter, r *http.Request) {
	feeds, err := s.recommendMgr.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load recommended feeds")
		return
	}
	writeJSON(w, http.StatusOK, feeds)
}

// handleBatchCreateFeeds creates multiple feeds at once
func (s *Server) handleBatchCreateFeeds(w http.ResponseWriter, r *http.Request) {
	claims := getUserFromContext(r)

	var req struct {
		URLs []string `json:"urls"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.URLs) == 0 {
		writeError(w, http.StatusBadRequest, "urls is required")
		return
	}

	var created []interface{}
	var errors []string

	for _, url := range req.URLs {
		// Check if feed already exists
		existing, _ := s.db.GetFeedByURL(claims.UserID, url)
		if existing != nil {
			continue // Skip existing feeds
		}

		// Create the feed
		feed, err := s.db.CreateFeed(claims.UserID, url, "", "", "", nil)
		if err != nil {
			errors = append(errors, url+": "+err.Error())
			continue
		}

		// Fetch feed content in background
		go s.fetcher.FetchFeed(feed)
		created = append(created, feed)
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"created": created,
		"errors":  errors,
	})
}

func (s *Server) handleAdminGetRecommendedFeeds(w http.ResponseWriter, r *http.Request) {
	feeds, err := s.recommendMgr.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load recommended feeds")
		return
	}
	writeJSON(w, http.StatusOK, feeds)
}

func (s *Server) handleAdminCreateRecommendedFeed(w http.ResponseWriter, r *http.Request) {
	var req models.RecommendedFeed
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	created, err := s.recommendMgr.Create(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) handleAdminUpdateRecommendedFeed(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req models.RecommendedFeed
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	updated, err := s.recommendMgr.Update(id, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) handleAdminDeleteRecommendedFeed(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := s.recommendMgr.Delete(id); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
