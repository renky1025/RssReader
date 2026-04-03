-- +migrate Up

-- users
CREATE TABLE IF NOT EXISTS users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  username TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  is_admin INTEGER DEFAULT 0,
  created_at INTEGER DEFAULT (strftime('%s','now'))
);

-- folders
CREATE TABLE IF NOT EXISTS folders (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL,
  name TEXT NOT NULL,
  parent_id INTEGER,
  created_at INTEGER DEFAULT (strftime('%s','now')),
  FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
  FOREIGN KEY(parent_id) REFERENCES folders(id) ON DELETE SET NULL
);

-- feeds
CREATE TABLE IF NOT EXISTS feeds (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL,
  folder_id INTEGER,
  url TEXT NOT NULL,
  title TEXT,
  site_url TEXT,
  description TEXT,
  last_fetched INTEGER,
  etag TEXT,
  last_modified TEXT,
  error_count INTEGER DEFAULT 0,
  last_error TEXT,
  disabled INTEGER DEFAULT 0,
  created_at INTEGER DEFAULT (strftime('%s','now')),
  FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
  FOREIGN KEY(folder_id) REFERENCES folders(id) ON DELETE SET NULL,
  UNIQUE(user_id, url)
);

-- articles
CREATE TABLE IF NOT EXISTS articles (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  feed_id INTEGER NOT NULL,
  guid TEXT,
  url TEXT,
  title TEXT,
  author TEXT,
  content TEXT,
  summary TEXT,
  image_url TEXT,
  published_at INTEGER,
  is_read INTEGER DEFAULT 0,
  is_starred INTEGER DEFAULT 0,
  is_read_later INTEGER DEFAULT 0,
  created_at INTEGER DEFAULT (strftime('%s','now')),
  FOREIGN KEY(feed_id) REFERENCES feeds(id) ON DELETE CASCADE,
  UNIQUE(feed_id, guid)
);

CREATE INDEX IF NOT EXISTS idx_articles_feed_id ON articles(feed_id);
CREATE INDEX IF NOT EXISTS idx_articles_published_at ON articles(published_at DESC);
CREATE INDEX IF NOT EXISTS idx_articles_is_read ON articles(is_read);
CREATE INDEX IF NOT EXISTS idx_articles_is_starred ON articles(is_starred);
CREATE INDEX IF NOT EXISTS idx_articles_is_read_later ON articles(is_read_later);

-- FTS for search (title + content)
CREATE VIRTUAL TABLE IF NOT EXISTS articles_fts USING fts5(
  title, 
  content, 
  content=articles, 
  content_rowid=id
);

-- Triggers to keep FTS in sync
CREATE TRIGGER IF NOT EXISTS articles_ai AFTER INSERT ON articles BEGIN
  INSERT INTO articles_fts(rowid, title, content) VALUES (new.id, new.title, new.content);
END;

CREATE TRIGGER IF NOT EXISTS articles_ad AFTER DELETE ON articles BEGIN
  INSERT INTO articles_fts(articles_fts, rowid, title, content) VALUES('delete', old.id, old.title, old.content);
END;

CREATE TRIGGER IF NOT EXISTS articles_au AFTER UPDATE ON articles BEGIN
  INSERT INTO articles_fts(articles_fts, rowid, title, content) VALUES('delete', old.id, old.title, old.content);
  INSERT INTO articles_fts(rowid, title, content) VALUES (new.id, new.title, new.content);
END;

-- tags
CREATE TABLE IF NOT EXISTS tags (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL,
  name TEXT NOT NULL,
  created_at INTEGER DEFAULT (strftime('%s','now')),
  FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
  UNIQUE(user_id, name)
);

-- article_tags
CREATE TABLE IF NOT EXISTS article_tags (
  article_id INTEGER NOT NULL,
  tag_id INTEGER NOT NULL,
  PRIMARY KEY(article_id, tag_id),
  FOREIGN KEY(article_id) REFERENCES articles(id) ON DELETE CASCADE,
  FOREIGN KEY(tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

-- saved_links (external bookmarks)
CREATE TABLE IF NOT EXISTS saved_links (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL,
  url TEXT NOT NULL,
  title TEXT,
  description TEXT,
  created_at INTEGER DEFAULT (strftime('%s','now')),
  FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Insert default admin user (password: admin)
INSERT INTO users (username, password_hash, is_admin) 
VALUES ('admin', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZRGdjGj/n3.QGpzgKLBqFMfGqUzHy', 1);

-- +migrate Down
DROP TABLE IF EXISTS saved_links;
DROP TABLE IF EXISTS article_tags;
DROP TABLE IF EXISTS tags;
DROP TRIGGER IF EXISTS articles_au;
DROP TRIGGER IF EXISTS articles_ad;
DROP TRIGGER IF EXISTS articles_ai;
DROP TABLE IF EXISTS articles_fts;
DROP TABLE IF EXISTS articles;
DROP TABLE IF EXISTS feeds;
DROP TABLE IF EXISTS folders;
DROP TABLE IF EXISTS users;
