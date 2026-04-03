-- +migrate Up

-- 优化文章查询性能的复合索引
-- 针对按feed_id查询的场景
CREATE INDEX IF NOT EXISTS idx_articles_feed_published ON articles(feed_id, published_at DESC);

-- 针对按feed_id和is_read状态查询的场景
CREATE INDEX IF NOT EXISTS idx_articles_feed_read ON articles(feed_id, is_read, published_at DESC);

-- 针对按folder查询的场景，需要通过feeds表关联
CREATE INDEX IF NOT EXISTS idx_feeds_folder_user ON feeds(folder_id, user_id);

-- 针对用户所有文章的查询优化
CREATE INDEX IF NOT EXISTS idx_feeds_user ON feeds(user_id);

-- 针对starred文章的查询
CREATE INDEX IF NOT EXISTS idx_articles_starred_published ON articles(is_starred, published_at DESC);

-- 针对read_later文章的查询
CREATE INDEX IF NOT EXISTS idx_articles_read_later_published ON articles(is_read_later, published_at DESC);

-- 针对未读文章的查询
CREATE INDEX IF NOT EXISTS idx_articles_unread_published ON articles(is_read, published_at DESC);

-- 复合索引用于统计查询
CREATE INDEX IF NOT EXISTS idx_articles_feed_status ON articles(feed_id, is_read, is_starred, is_read_later);

-- +migrate Down
DROP INDEX IF EXISTS idx_articles_feed_published;
DROP INDEX IF EXISTS idx_articles_feed_read;
DROP INDEX IF EXISTS idx_feeds_folder_user;
DROP INDEX IF EXISTS idx_feeds_user;
DROP INDEX IF EXISTS idx_articles_starred_published;
DROP INDEX IF EXISTS idx_articles_read_later_published;
DROP INDEX IF EXISTS idx_articles_unread_published;
DROP INDEX IF EXISTS idx_articles_feed_status;