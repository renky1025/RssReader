package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"rssreader/internal/models"
)

func scanUser(row interface{ Scan(...interface{}) error }) (*models.User, error) {
	var user models.User
	var createdAt int64
	var email sql.NullString
	var status sql.NullInt64
	var lastLoginAt sql.NullInt64
	var onboardingComplete sql.NullInt64
	var lastLoginIP sql.NullString
	var lastLoginDevice sql.NullString

	err := row.Scan(
		&user.ID, &user.Username, &email, &user.PasswordHash,
		&user.IsAdmin, &status, &lastLoginAt, &onboardingComplete,
		&lastLoginIP, &lastLoginDevice, &createdAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	user.Email = email.String
	user.Status = int(status.Int64)
	if !status.Valid {
		user.Status = models.UserStatusActive // default
	}
	user.CreatedAt = time.Unix(createdAt, 0)
	if lastLoginAt.Valid {
		t := time.Unix(lastLoginAt.Int64, 0)
		user.LastLoginAt = &t
	}
	user.OnboardingComplete = onboardingComplete.Valid && onboardingComplete.Int64 == 1
	user.LastLoginIP = lastLoginIP.String
	user.LastLoginDevice = lastLoginDevice.String
	return &user, nil
}

func (db *DB) GetUserByUsername(username string) (*models.User, error) {
	row := db.QueryRow(`
		SELECT id, username, email, password_hash, is_admin, status, last_login_at, onboarding_complete, last_login_ip, last_login_device, created_at 
		FROM users WHERE username = ?
	`, username)
	return scanUser(row)
}

func (db *DB) GetUserByEmail(email string) (*models.User, error) {
	row := db.QueryRow(`
		SELECT id, username, email, password_hash, is_admin, status, last_login_at, onboarding_complete, last_login_ip, last_login_device, created_at 
		FROM users WHERE email = ?
	`, email)
	return scanUser(row)
}

func (db *DB) GetUserByID(id int64) (*models.User, error) {
	row := db.QueryRow(`
		SELECT id, username, email, password_hash, is_admin, status, last_login_at, onboarding_complete, last_login_ip, last_login_device, created_at 
		FROM users WHERE id = ?
	`, id)
	return scanUser(row)
}

func (db *DB) CreateUser(username, email, passwordHash string, isAdmin bool) (*models.User, error) {
	result, err := db.Exec(`
		INSERT INTO users (username, email, password_hash, is_admin, status) VALUES (?, ?, ?, ?, ?)
	`, username, email, passwordHash, isAdmin, models.UserStatusActive)
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	return db.GetUserByID(id)
}

func (db *DB) UpdateUserLastLogin(userID int64, ip, device string) error {
	_, err := db.Exec(`UPDATE users SET last_login_at = strftime('%s','now'), last_login_ip = ?, last_login_device = ? WHERE id = ?`, ip, device, userID)
	return err
}

func (db *DB) CompleteOnboarding(userID int64) error {
	_, err := db.Exec(`UPDATE users SET onboarding_complete = 1 WHERE id = ?`, userID)
	return err
}

func (db *DB) UpdateUser(userID int64, req models.UpdateUserRequest) error {
	var updates []string
	var args []interface{}

	if req.Email != nil {
		updates = append(updates, "email = ?")
		args = append(args, *req.Email)
	}
	if req.Password != nil {
		updates = append(updates, "password_hash = ?")
		args = append(args, *req.Password)
	}
	if req.Status != nil {
		updates = append(updates, "status = ?")
		args = append(args, *req.Status)
	}
	if req.IsAdmin != nil {
		updates = append(updates, "is_admin = ?")
		args = append(args, *req.IsAdmin)
	}

	if len(updates) == 0 {
		return nil
	}

	args = append(args, userID)
	query := fmt.Sprintf("UPDATE users SET %s WHERE id = ?", strings.Join(updates, ", "))
	_, err := db.Exec(query, args...)
	return err
}

func (db *DB) DeleteUser(userID int64) error {
	_, err := db.Exec(`DELETE FROM users WHERE id = ?`, userID)
	return err
}

// GetAllUsers returns all users with optional filtering
func (db *DB) GetAllUsers(params models.AdminUserListParams) ([]*models.User, int, error) {
	var conditions []string
	var args []interface{}

	if params.Status != nil {
		conditions = append(conditions, "status = ?")
		args = append(args, *params.Status)
	}
	if params.Query != "" {
		conditions = append(conditions, "(username LIKE ? OR email LIKE ?)")
		q := "%" + params.Query + "%"
		args = append(args, q, q)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Get total count
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM users %s", whereClause)
	if err := db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get users with pagination
	if params.Limit <= 0 {
		params.Limit = 50
	}
	query := fmt.Sprintf(`
		SELECT id, username, email, password_hash, is_admin, status, last_login_at, onboarding_complete, last_login_ip, last_login_device, created_at 
		FROM users %s 
		ORDER BY created_at DESC 
		LIMIT ? OFFSET ?
	`, whereClause)
	args = append(args, params.Limit, params.Offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, user)
	}

	return users, total, nil
}

// GetAdminStats returns global statistics for admin dashboard
func (db *DB) GetAdminStats() (*models.AdminStats, error) {
	stats := &models.AdminStats{}

	// User stats
	db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&stats.TotalUsers)
	db.QueryRow(`SELECT COUNT(*) FROM users WHERE status = 1`).Scan(&stats.ActiveUsers)
	db.QueryRow(`SELECT COUNT(*) FROM users WHERE status = 0`).Scan(&stats.DisabledUsers)
	db.QueryRow(`SELECT COUNT(*) FROM users WHERE is_admin = 1`).Scan(&stats.AdminUsers)

	// Content stats
	db.QueryRow(`SELECT COUNT(*) FROM feeds`).Scan(&stats.TotalFeeds)
	db.QueryRow(`SELECT COUNT(*) FROM articles`).Scan(&stats.TotalArticles)
	db.QueryRow(`SELECT COUNT(*) FROM folders`).Scan(&stats.TotalFolders)

	// Time-based article stats
	now := time.Now().Unix()
	dayAgo := now - 86400
	weekAgo := now - 86400*7
	monthAgo := now - 86400*30

	db.QueryRow(`SELECT COUNT(*) FROM articles WHERE created_at >= ?`, dayAgo).Scan(&stats.ArticlesToday)
	db.QueryRow(`SELECT COUNT(*) FROM articles WHERE created_at >= ?`, weekAgo).Scan(&stats.ArticlesThisWeek)
	db.QueryRow(`SELECT COUNT(*) FROM articles WHERE created_at >= ?`, monthAgo).Scan(&stats.ArticlesThisMonth)

	return stats, nil
}

// GetUserStats returns per-user statistics for admin view
func (db *DB) GetUserStats(userID int64) (*models.UserStats, error) {
	stats := &models.UserStats{UserID: userID}

	db.QueryRow(`SELECT username FROM users WHERE id = ?`, userID).Scan(&stats.Username)
	db.QueryRow(`SELECT COUNT(*) FROM feeds WHERE user_id = ?`, userID).Scan(&stats.FeedCount)
	db.QueryRow(`
		SELECT COUNT(*) FROM articles a 
		JOIN feeds f ON a.feed_id = f.id 
		WHERE f.user_id = ?
	`, userID).Scan(&stats.ArticleCount)
	db.QueryRow(`
		SELECT COUNT(*) FROM articles a 
		JOIN feeds f ON a.feed_id = f.id 
		WHERE f.user_id = ? AND a.is_read = 0
	`, userID).Scan(&stats.UnreadCount)
	db.QueryRow(`
		SELECT COUNT(*) FROM articles a 
		JOIN feeds f ON a.feed_id = f.id 
		WHERE f.user_id = ? AND a.is_starred = 1
	`, userID).Scan(&stats.StarredCount)

	return stats, nil
}
