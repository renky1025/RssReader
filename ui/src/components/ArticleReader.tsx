import { useEffect, useMemo } from 'react'
import { format } from 'date-fns'
import DOMPurify from 'dompurify'
import {
  ExternalLink,
  Check,
  Star,
  Clock,
  ThumbsUp,
  ThumbsDown,
  Archive,
} from 'lucide-react'
import { useUpdateArticle } from '../api/hooks'
import type { Article } from '../types'
import { getProxyImageUrl, processContentImages } from '../utils/imageProxy'

interface ArticleReaderProps {
  article: Article
}

export default function ArticleReader({ article }: ArticleReaderProps) {
  const updateArticle = useUpdateArticle()

  // Mark as read when viewing
  useEffect(() => {
    if (!article.is_read) {
      updateArticle.mutate({ id: article.id, is_read: true })
    }
  }, [article.id])

  const handleToggleStar = () => {
    updateArticle.mutate({ id: article.id, is_starred: !article.is_starred })
  }

  const handleToggleReadLater = () => {
    updateArticle.mutate({ id: article.id, is_read_later: !article.is_read_later })
  }

  const handleMarkUnread = () => {
    updateArticle.mutate({ id: article.id, is_read: false })
  }

  const publishedDate = article.published_at
    ? format(new Date(article.published_at * 1000), 'MMMM d, yyyy')
    : ''

  // Process images through proxy to avoid 403 errors from hotlink protection
  const sanitizedContent = useMemo(() => {
    const rawContent = article.content || article.summary || ''
    // First process images to use proxy, then sanitize
    let proxiedContent = processContentImages(rawContent)
    
    // Remove duplicate images that match the featured image
    if (article.image_url) {
      const imageUrlPattern = article.image_url.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
      const regex = new RegExp(`<img[^>]*src=["']([^"']*${imageUrlPattern}[^"']*)["'][^>]*>`, 'gi')
      proxiedContent = proxiedContent.replace(regex, '')
    }
    
    return DOMPurify.sanitize(proxiedContent, {
      ADD_TAGS: ['iframe'],
      ADD_ATTR: ['allow', 'allowfullscreen', 'frameborder', 'scrolling'],
    })
  }, [article.content, article.summary, article.image_url])

  // Proxy the featured image URL
  const proxiedImageUrl = useMemo(() => {
    return article.image_url ? getProxyImageUrl(article.image_url) : undefined
  }, [article.image_url])

  return (
    <div className="h-full flex flex-col bg-[var(--bg-primary)]">
      {/* Header */}
      <div className="p-3 md:p-4 border-b border-[var(--border-color)]">
        <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-2">
          <div className="flex items-center gap-2 text-sm text-[var(--text-secondary)] truncate">
            <span className="truncate">{article.feed_title}</span>
            <span>•</span>
            <span className="whitespace-nowrap">{publishedDate}</span>
          </div>

          {/* Actions - scrollable on mobile */}
          <div className="flex items-center gap-1 overflow-x-auto pb-1 md:pb-0 -mx-1 px-1">
            <a
              href={article.url}
              target="_blank"
              rel="noopener noreferrer"
              className="p-2 hover:bg-[var(--bg-hover)] rounded-lg transition-colors flex-shrink-0"
              title="Open Original"
            >
              <ExternalLink className="w-4 h-4" />
            </a>
            <button
              onClick={handleMarkUnread}
              className="p-2 hover:bg-[var(--bg-hover)] rounded-lg transition-colors flex-shrink-0"
              title="Mark Unread"
            >
              <Check className="w-4 h-4" />
            </button>
            <button
              onClick={() => { }}
              className="p-2 hover:bg-[var(--bg-hover)] rounded-lg transition-colors flex-shrink-0"
              title="Like"
            >
              <ThumbsUp className="w-4 h-4" />
            </button>
            <button
              onClick={() => { }}
              className="p-2 hover:bg-[var(--bg-hover)] rounded-lg transition-colors flex-shrink-0"
              title="Dislike"
            >
              <ThumbsDown className="w-4 h-4" />
            </button>
            <button
              onClick={handleToggleReadLater}
              className={`p-2 hover:bg-[var(--bg-hover)] rounded-lg transition-colors flex-shrink-0 ${article.is_read_later ? 'text-[var(--accent)]' : ''
                }`}
              title="Read Later"
            >
              <Clock className="w-4 h-4" />
            </button>
            <button
              onClick={() => { }}
              className="p-2 hover:bg-[var(--bg-hover)] rounded-lg transition-colors flex-shrink-0"
              title="Archive"
            >
              <Archive className="w-4 h-4" />
            </button>
          </div>
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-y-auto p-4 md:p-6">
        <article className="max-w-3xl mx-auto">
          {/* Title */}
          <h1 className="text-3xl font-bold mb-6 leading-tight dark:text-white text-[var(--text-primary)]">{article.title}</h1>

          {/* Star button */}
          <button
            onClick={handleToggleStar}
            className={`flex items-center gap-2 mb-6 px-3 py-1.5 rounded-full border transition-colors ${article.is_starred
                ? 'border-yellow-500 text-yellow-500'
                : 'border-[var(--border-color)] text-[var(--text-secondary)] hover:border-[var(--text-muted)]'
              }`}
          >
            <Star className={`w-4 h-4 ${article.is_starred ? 'fill-current' : ''}`} />
            <span className="text-sm">{article.is_starred ? 'Starred' : 'Star'}</span>
          </button>

          {/* Featured image */}
          {proxiedImageUrl && (
            <div className="mb-8 rounded-xl overflow-hidden shadow-lg">
              <img
                src={proxiedImageUrl}
                alt={article.title}
                className="w-full h-auto max-h-96 object-cover"
                loading="lazy"
              />
            </div>
          )}

          {/* Article content */}
          <div
            className="article-content prose prose-lg max-w-none"
            dangerouslySetInnerHTML={{ __html: sanitizedContent }}
          />

          {/* Footer */}
          <div className="mt-8 pt-6 border-t border-[var(--border-color)]">
            <a
              href={article.url}
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-2 text-[var(--accent)] hover:underline"
            >
              <ExternalLink className="w-4 h-4" />
              Read original article
            </a>
          </div>
        </article>
      </div>
    </div>
  )
}
