import { useState, useMemo } from 'react'
import { Rss, Check, Loader2, Plus, ArrowLeft } from 'lucide-react'
import { useRecommendedFeeds, useBatchCreateFeeds, useFeeds } from '../api/hooks'
import { showToast } from '../components/Toast'
import { useAppStore } from '../stores/appStore'
import type { RecommendedFeed } from '../types'

export default function Discover() {
  const { data: recommendedFeeds = [], isLoading } = useRecommendedFeeds()
  const { data: userFeeds = [] } = useFeeds()
  const batchCreate = useBatchCreateFeeds()
  const [selectedFeeds, setSelectedFeeds] = useState<Set<string>>(new Set())
  const setCurrentView = useAppStore((state) => state.setCurrentView)

  // Get URLs of feeds user already has
  const existingFeedUrls = useMemo(() => {
    return new Set(userFeeds.map((f) => f.url))
  }, [userFeeds])

  // Group feeds by category
  const feedsByCategory = useMemo(() => {
    const grouped: Record<string, RecommendedFeed[]> = {}
    recommendedFeeds.forEach((feed) => {
      if (!grouped[feed.category]) {
        grouped[feed.category] = []
      }
      grouped[feed.category].push(feed)
    })
    return grouped
  }, [recommendedFeeds])

  const toggleFeed = (feedId: string) => {
    const newSelected = new Set(selectedFeeds)
    if (newSelected.has(feedId)) {
      newSelected.delete(feedId)
    } else {
      newSelected.add(feedId)
    }
    setSelectedFeeds(newSelected)
  }

  const selectAll = () => {
    // Only select feeds that user doesn't already have
    const availableFeeds = recommendedFeeds
      .filter((f) => !existingFeedUrls.has(f.url))
      .map((f) => f.id)
    setSelectedFeeds(new Set(availableFeeds))
  }

  const selectNone = () => {
    setSelectedFeeds(new Set())
  }

  const handleAddFeeds = async () => {
    if (selectedFeeds.size === 0) return

    const urls = recommendedFeeds
      .filter((f) => selectedFeeds.has(f.id))
      .map((f) => f.url)

    try {
      await batchCreate.mutateAsync(urls)
      showToast('success', `Added ${selectedFeeds.size} feeds`)
      setSelectedFeeds(new Set())
    } catch (err) {
      showToast('error', 'Failed to add some feeds')
    }
  }

  if (isLoading) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <Loader2 className="w-8 h-8 animate-spin text-[var(--accent)]" />
      </div>
    )
  }

  return (
    <div className="flex-1 overflow-y-auto bg-[var(--bg-primary)]">
      <div className="max-w-4xl mx-auto py-8 px-4">
        {/* Header */}
        <div className="mb-8">
          <button
            onClick={() => setCurrentView({ type: 'all', title: 'All Feeds' })}
            className="flex items-center gap-2 text-[var(--text-secondary)] hover:text-[var(--text-primary)] mb-4"
          >
            <ArrowLeft className="w-4 h-4" />
            Back to feeds
          </button>
          <h1 className="text-2xl font-bold text-[var(--text-primary)] mb-2">
            Discover Feeds
          </h1>
          <p className="text-[var(--text-secondary)]">
            Browse and add recommended RSS feeds to your collection.
          </p>
        </div>

        {/* Selection controls */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 mb-6">
          <div className="flex items-center gap-4">
            <button
              onClick={selectAll}
              className="text-sm text-[var(--accent)] hover:underline"
            >
              Select All Available
            </button>
            <button
              onClick={selectNone}
              className="text-sm text-[var(--text-secondary)] hover:underline"
            >
              Clear
            </button>
          </div>
          <div className="flex items-center gap-4">
            <span className="text-sm text-[var(--text-muted)]">
              {selectedFeeds.size} selected
            </span>
            {selectedFeeds.size > 0 && (
              <button
                onClick={handleAddFeeds}
                disabled={batchCreate.isPending}
                className="flex items-center gap-2 px-4 py-2 bg-[var(--accent)] hover:opacity-90 text-white text-sm font-medium rounded-lg transition-colors disabled:opacity-50"
              >
                {batchCreate.isPending ? (
                  <>
                    <Loader2 className="w-4 h-4 animate-spin" />
                    Adding...
                  </>
                ) : (
                  <>
                    <Plus className="w-4 h-4" />
                    Add {selectedFeeds.size} Feeds
                  </>
                )}
              </button>
            )}
          </div>
        </div>

        {/* Feed categories */}
        <div className="space-y-8">
          {Object.entries(feedsByCategory).map(([category, categoryFeeds]) => (
            <div key={category}>
              <h2 className="text-lg font-semibold text-[var(--text-primary)] mb-4">
                {category}
              </h2>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                {categoryFeeds.map((feed) => {
                  const alreadyAdded = existingFeedUrls.has(feed.url)
                  return (
                    <FeedCard
                      key={feed.id}
                      feed={feed}
                      selected={selectedFeeds.has(feed.id)}
                      alreadyAdded={alreadyAdded}
                      onToggle={() => !alreadyAdded && toggleFeed(feed.id)}
                    />
                  )
                })}
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}

function FeedCard({
  feed,
  selected,
  alreadyAdded,
  onToggle,
}: {
  feed: RecommendedFeed
  selected: boolean
  alreadyAdded: boolean
  onToggle: () => void
}) {
  return (
    <button
      onClick={onToggle}
      disabled={alreadyAdded}
      className={`flex items-start gap-3 p-4 rounded-xl border transition-all text-left ${
        alreadyAdded
          ? 'border-green-500/30 bg-green-500/5 cursor-default'
          : selected
          ? 'border-[var(--accent)] bg-[var(--accent)]/10'
          : 'border-[var(--border-color)] bg-[var(--bg-secondary)] hover:border-[var(--text-muted)]'
      }`}
    >
      {/* Icon */}
      <div className="w-10 h-10 rounded-lg bg-[var(--bg-tertiary)] flex items-center justify-center flex-shrink-0 overflow-hidden">
        {feed.icon ? (
          <img
            src={feed.icon}
            alt=""
            className="w-6 h-6"
            onError={(e) => {
              e.currentTarget.style.display = 'none'
              e.currentTarget.parentElement!.innerHTML = `<svg class="w-5 h-5 text-[var(--text-muted)]" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M4 11a9 9 0 0 1 9 9"/><path d="M4 4a16 16 0 0 1 16 16"/><circle cx="5" cy="19" r="1"/></svg>`
            }}
          />
        ) : (
          <Rss className="w-5 h-5 text-[var(--text-muted)]" />
        )}
      </div>

      {/* Content */}
      <div className="flex-1 min-w-0">
        <h3 className="font-medium text-[var(--text-primary)] truncate">
          {feed.name}
        </h3>
        <p className="text-sm text-[var(--text-secondary)] line-clamp-2">
          {feed.description}
        </p>
      </div>

      {/* Status indicator */}
      <div
        className={`w-5 h-5 rounded-full border-2 flex items-center justify-center flex-shrink-0 transition-colors ${
          alreadyAdded
            ? 'border-green-500 bg-green-500'
            : selected
            ? 'border-[var(--accent)] bg-[var(--accent)]'
            : 'border-[var(--text-muted)]'
        }`}
      >
        {(alreadyAdded || selected) && <Check className="w-3 h-3 text-white" />}
      </div>
    </button>
  )
}
