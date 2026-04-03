package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"rssreader/internal/models"
)

func (db *DB) GetArticles(userID int64, params models.ArticleListParams) ([]*models.Article, int, error) {
	// 优化：针对特定feed的查询，可以跳过JOIN feeds表的权限检查
	if params.FeedID != nil {
		return db.getArticlesByFeed(userID, *params.FeedID, params)
	}
	
	var conditions []string
	var args []interface{}

	// Base condition: articles from user's feeds
	conditions = append(conditions, "f.user_id = ?")
	args = append(args, userID)

	if params.FolderID != nil {
		conditions = append(conditions, "f.folder_id = ?")
		args = append(args, *params.FolderID)
	}

	if params.IsRead != nil {
		conditions = append(conditions, "a.is_read = ?")
		args = append(args, *params.IsRead)
	}

	if params.IsStarred != nil {
		conditions = append(conditions, "a.is_starred = ?")
		args = append(args, *params.IsStarred)
	}

	if params.IsReadLater != nil {
		conditions = append(conditions, "a.is_read_later = ?")
		args = append(args, *params.IsReadLater)
	}

	whereClause := strings.Join(conditions, " AND ")

	// Handle full-text search
	var fromClause string
	if params.Query != "" {
		fromClause = `
			FROM articles a
			JOIN feeds f ON a.feed_id = f.id
			JOIN articles_fts fts ON fts.rowid = a.id
			WHERE ` + whereClause + ` AND articles_fts MATCH ?`
		args = append(args, params.Query)
	} else {
		fromClause = `
			FROM articles a
			JOIN feeds f ON a.feed_id = f.id
			WHERE ` + whereClause
	}

	// Get total count (优化：只在需要时计算总数)
	var total int
	if params.Offset == 0 {
		countQuery := "SELECT COUNT(*) " + fromClause
		countArgs := make([]interface{}, len(args))
		copy(countArgs, args)
		err := db.QueryRow(countQuery, countArgs...).Scan(&total)
		if err != nil {
			return nil, 0, err
		}
	} else {
		// 对于分页查询，设置一个估算值，避免每次都计算总数
		total = -1
	}

	// Get articles
	if params.Limit <= 0 {
		params.Limit = 50
	}
	if params.Limit > 100 {
		params.Limit = 100
	}

	query := fmt.Sprintf(`
		SELECT a.id, a.feed_id, a.guid, a.url, a.title, a.author, a.content, a.summary,
			   a.image_url, a.published_at, a.is_read, a.is_starred, a.is_read_later, a.created_at,
			   f.title as feed_title
		%s
		ORDER BY a.published_at DESC
		LIMIT ? OFFSET ?
	`, fromClause)
	args = append(args, params.Limit, params.Offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var articles []*models.Article
	for rows.Next() {
		var a models.Article
		var createdAt int64
		err := rows.Scan(&a.ID, &a.FeedID, &a.GUID, &a.URL, &a.Title, &a.Author,
			&a.Content, &a.Summary, &a.ImageURL, &a.PublishedAt, &a.IsRead,
			&a.IsStarred, &a.IsReadLater, &createdAt, &a.FeedTitle)
		if err != nil {
			return nil, 0, err
		}
		a.CreatedAt = time.Unix(createdAt, 0)
		articles = append(articles, &a)
	}
	return articles, total, nil
}

// 优化的单feed查询方法
func (db *DB) getArticlesByFeed(userID int64, feedID int64, params models.ArticleListParams) ([]*models.Article, int, error) {
	// 首先验证feed属于该用户
	var feedTitle string
	err := db.QueryRow("SELECT title FROM feeds WHERE id = ? AND user_id = ?", feedID, userID).Scan(&feedTitle)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, 0, fmt.Errorf("feed not found")
		}
		return nil, 0, err
	}

	var conditions []string
	var args []interface{}

	conditions = append(conditions, "feed_id = ?")
	args = append(args, feedID)

	if params.IsRead != nil {
		conditions = append(conditions, "is_read = ?")
		args = append(args, *params.IsRead)
	}

	if params.IsStarred != nil {
		conditions = append(conditions, "is_starred = ?")
		args = append(args, *params.IsStarred)
	}

	if params.IsReadLater != nil {
		conditions = append(conditions, "is_read_later = ?")
		args = append(args, *params.IsReadLater)
	}

	whereClause := strings.Join(conditions, " AND ")

	// Handle full-text search
	var fromClause string
	if params.Query != "" {
		fromClause = `
			FROM articles a
			JOIN articles_fts fts ON fts.rowid = a.id
			WHERE ` + whereClause + ` AND articles_fts MATCH ?`
		args = append(args, params.Query)
	} else {
		fromClause = `FROM articles WHERE ` + whereClause
	}

	// Get total count (优化：只在首页计算)
	var total int
	if params.Offset == 0 {
		countQuery := "SELECT COUNT(*) " + fromClause
		countArgs := make([]interface{}, len(args))
		copy(countArgs, args)
		err := db.QueryRow(countQuery, countArgs...).Scan(&total)
		if err != nil {
			return nil, 0, err
		}
	} else {
		total = -1
	}

	// Get articles
	if params.Limit <= 0 {
		params.Limit = 50
	}
	if params.Limit > 100 {
		params.Limit = 100
	}

	query := fmt.Sprintf(`
		SELECT id, feed_id, guid, url, title, author, content, summary,
			   image_url, published_at, is_read, is_starred, is_read_later, created_at
		%s
		ORDER BY published_at DESC
		LIMIT ? OFFSET ?
	`, fromClause)
	args = append(args, params.Limit, params.Offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var articles []*models.Article
	for rows.Next() {
		var a models.Article
		var createdAt int64
		err := rows.Scan(&a.ID, &a.FeedID, &a.GUID, &a.URL, &a.Title, &a.Author,
			&a.Content, &a.Summary, &a.ImageURL, &a.PublishedAt, &a.IsRead,
			&a.IsStarred, &a.IsReadLater, &createdAt)
		if err != nil {
			return nil, 0, err
		}
		a.CreatedAt = time.Unix(createdAt, 0)
		a.FeedTitle = feedTitle // 直接设置，避免JOIN
		articles = append(articles, &a)
	}
	return articles, total, nil
}

func (db *DB) GetArticleByID(id int64) (*models.Article, error) {
	var a models.Article
	var createdAt int64
	err := db.QueryRow(`
		SELECT a.id, a.feed_id, a.guid, a.url, a.title, a.author, a.content, a.summary,
			   a.image_url, a.published_at, a.is_read, a.is_starred, a.is_read_later, a.created_at,
			   f.title as feed_title
		FROM articles a
		JOIN feeds f ON a.feed_id = f.id
		WHERE a.id = ?
	`, id).Scan(&a.ID, &a.FeedID, &a.GUID, &a.URL, &a.Title, &a.Author,
		&a.Content, &a.Summary, &a.ImageURL, &a.PublishedAt, &a.IsRead,
		&a.IsStarred, &a.IsReadLater, &createdAt, &a.FeedTitle)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	a.CreatedAt = time.Unix(createdAt, 0)
	return &a, nil
}

func (db *DB) CreateArticle(feedID int64, guid, url, title, author, content, summary, imageURL string, publishedAt int64) (*models.Article, error) {
	result, err := db.Exec(`
		INSERT OR IGNORE INTO articles (feed_id, guid, url, title, author, content, summary, image_url, published_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, feedID, guid, url, title, author, content, summary, imageURL, publishedAt)
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	if id == 0 {
		// Article already exists
		return nil, nil
	}
	return db.GetArticleByID(id)
}

func (db *DB) UpdateArticle(id int64, req models.UpdateArticleRequest) error {
	var updates []string
	var args []interface{}

	if req.IsRead != nil {
		updates = append(updates, "is_read = ?")
		args = append(args, *req.IsRead)
	}
	if req.IsStarred != nil {
		updates = append(updates, "is_starred = ?")
		args = append(args, *req.IsStarred)
	}
	if req.IsReadLater != nil {
		updates = append(updates, "is_read_later = ?")
		args = append(args, *req.IsReadLater)
	}

	if len(updates) == 0 {
		return nil
	}

	args = append(args, id)
	query := fmt.Sprintf("UPDATE articles SET %s WHERE id = ?", strings.Join(updates, ", "))
	_, err := db.Exec(query, args...)
	return err
}

func (db *DB) MarkAllAsRead(userID int64, feedID *int64, folderID *int64) error {
	var query string
	var args []interface{}

	if feedID != nil {
		query = `UPDATE articles SET is_read = 1 WHERE feed_id = ?`
		args = append(args, *feedID)
	} else if folderID != nil {
		query = `UPDATE articles SET is_read = 1 WHERE feed_id IN (SELECT id FROM feeds WHERE folder_id = ? AND user_id = ?)`
		args = append(args, *folderID, userID)
	} else {
		query = `UPDATE articles SET is_read = 1 WHERE feed_id IN (SELECT id FROM feeds WHERE user_id = ?)`
		args = append(args, userID)
	}

	_, err := db.Exec(query, args...)
	return err
}

func (db *DB) GetStats(userID int64) (*models.Stats, error) {
	var stats models.Stats

	err := db.QueryRow(`
		SELECT 
			(SELECT COUNT(*) FROM feeds WHERE user_id = ?) as total_feeds,
			(SELECT COUNT(*) FROM articles a JOIN feeds f ON a.feed_id = f.id WHERE f.user_id = ?) as total_articles,
			(SELECT COUNT(*) FROM articles a JOIN feeds f ON a.feed_id = f.id WHERE f.user_id = ? AND a.is_read = 0) as unread_articles,
			(SELECT COUNT(*) FROM articles a JOIN feeds f ON a.feed_id = f.id WHERE f.user_id = ? AND a.is_starred = 1) as starred_articles,
			(SELECT COUNT(*) FROM articles a JOIN feeds f ON a.feed_id = f.id WHERE f.user_id = ? AND a.is_read_later = 1) as read_later_count,
			(SELECT COUNT(*) FROM folders WHERE user_id = ?) as total_folders,
			(SELECT COUNT(*) FROM feeds WHERE user_id = ? AND disabled = 1) as disabled_feeds,
			(SELECT COUNT(*) FROM feeds WHERE user_id = ? AND error_count > 0) as error_feeds
	`, userID, userID, userID, userID, userID, userID, userID, userID).Scan(
		&stats.TotalFeeds, &stats.TotalArticles, &stats.UnreadArticles,
		&stats.StarredArticles, &stats.ReadLaterCount, &stats.TotalFolders,
		&stats.DisabledFeeds, &stats.ErrorFeeds)

	return &stats, err
}
