import { format } from 'date-fns'
import { Check, CheckCheck, RefreshCw, Search } from 'lucide-react'
import { useState } from 'react'
import { useArticles, useMarkAllRead, useFetchFeed } from '../api/hooks'
import { useAppStore } from '../stores/appStore'
import type { Article } from '../types'
import { getProxyImageUrl } from '../utils/imageProxy'

export default function ArticleList() {
  const { currentView, selectedArticle, setSelectedArticle } = useAppStore()
  const [searchQuery, setSearchQuery] = useState('')
  const markAllRead = useMarkAllRead()
  const fetchFeed = useFetchFeed()

  const params: Record<string, unknown> = { limit: 50 }

  if (currentView.type === 'feed' && currentView.id) {
    params.feed_id = currentView.id
  } else if (currentView.type === 'folder' && currentView.id) {
    params.folder_id = currentView.id
  } else if (currentView.type === 'unread') {
    params.is_read = false
  } else if (currentView.type === 'starred') {
    params.is_starred = true
  } else if (currentView.type === 'read-later') {
    params.is_read_later = true
  }

  if (searchQuery) {
    params.q = searchQuery
  }

  const { data, isLoading, refetch } = useArticles(params)
  const articles = data?.items || []

  const handleMarkAllRead = () => {
    const markParams: { feed_id?: number; folder_id?: number } = {}
    if (currentView.type === 'feed' && currentView.id) {
      markParams.feed_id = currentView.id
    } else if (currentView.type === 'folder' && currentView.id) {
      markParams.folder_id = currentView.id
    }
    markAllRead.mutate(markParams)
  }

  const handleRefresh = () => {
    if (currentView.type === 'feed' && currentView.id) {
      fetchFeed.mutate(currentView.id)
    }
    refetch()
  }

  return (
    <div className="h-full flex flex-col bg-[var(--bg-tertiary)]">
      {/* Header */}
      <div className="p-4 border-b border-[var(--border-color)]">
        <div className="flex items-center justify-between mb-3">
          <h2 className="text-lg font-semibold text-[var(--text-primary)] truncate">{currentView.title}</h2>
          <div className="flex items-center gap-2">
            <button
              onClick={handleRefresh}
              className="p-2 hover:bg-[var(--bg-hover)] rounded-lg transition-colors"
              title="Refresh"
            >
              <RefreshCw className={`w-4 h-4 ${fetchFeed.isPending ? 'animate-spin' : ''}`} />
            </button>
            <button
              onClick={handleMarkAllRead}
              className="p-2 hover:bg-[var(--bg-hover)] rounded-lg transition-colors"
              title="Mark all as read"
            >
              <CheckCheck className="w-4 h-4" />
            </button>
          </div>
        </div>

        {/* Search */}
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="Search articles..."
            className="w-full pl-10 pr-4 py-2 bg-[var(--bg-secondary)] border border-[var(--border-color)] rounded-lg text-[var(--text-primary)] text-sm focus:outline-none focus:border-[var(--accent)]"
          />
        </div>
      </div>

      {/* Article list */}
      <div className="flex-1 overflow-y-auto">
        {isLoading ? (
          <div className="flex items-center justify-center h-32">
            <RefreshCw className="w-6 h-6 animate-spin text-gray-400" />
          </div>
        ) : articles.length === 0 ? (
          <div className="flex items-center justify-center h-32 text-[var(--text-muted)]">
            No articles found
          </div>
        ) : (
          articles.map((article: Article) => (
            <ArticleItem
              key={article.id}
              article={article}
              isSelected={selectedArticle?.id === article.id}
              onClick={() => setSelectedArticle(article)}
            />
          ))
        )}
      </div>

      {/* Footer */}
      {data && (
        <div className="p-2 border-t border-[var(--border-color)] text-center text-xs text-[var(--text-muted)]">
          {data.total} articles
        </div>
      )}
    </div>
  )
}

function ArticleItem({
  article,
  isSelected,
  onClick,
}: {
  article: Article
  isSelected: boolean
  onClick: () => void
}) {
  const publishedDate = article.published_at
    ? format(new Date(article.published_at * 1000), 'MMM d')
    : ''

  return (
    <button
      onClick={onClick}
      className={`w-full text-left p-4 border-b border-[var(--border-color)]/50 transition-colors ${isSelected ? 'bg-[var(--bg-hover)] border-l-4 border-l-[var(--accent)]' : 'hover:bg-[var(--bg-hover)]'
        } ${article.is_read ? 'opacity-60' : ''}`}
    >
      <div className="flex items-start gap-3">
        {/* Thumbnail */}
        {article.image_url && (
          <div className="w-20 h-16 flex-shrink-0 rounded-xl overflow-hidden bg-[var(--bg-tertiary)] shadow-sm">
            <img
              src={getProxyImageUrl(article.image_url)}
              alt=""
              className="w-full h-full object-cover"
              loading="lazy"
            />
          </div>
        )}

        <div className="flex-1 min-w-0">
          {/* Title */}
          <h3 className={`text-sm font-medium mb-1 line-clamp-2 ${article.is_read ? 'text-[var(--text-muted)]' : 'text-[var(--text-primary)]'}`}>
            {article.title}
          </h3>

          {/* Meta */}
          <div className="flex items-center gap-2 text-xs text-[var(--text-muted)]">
            <span className="truncate">{article.feed_title}</span>
            <span>•</span>
            <span>{publishedDate}</span>
            {article.is_read && <Check className="w-3 h-3" />}
          </div>

          {/* Summary */}
          {article.summary && (
            <p className="text-xs text-[var(--text-secondary)] mt-1 line-clamp-2">
              {article.summary.replace(/<[^>]*>/g, '')}
            </p>
          )}
        </div>
      </div>
    </button>
  )
}
