import { useState, useMemo } from 'react'
import { useNavigate } from 'react-router-dom'
import { Rss, Check, Loader2, ChevronRight } from 'lucide-react'
import { useRecommendedFeeds, useBatchCreateFeeds, useCompleteOnboarding } from '../api/hooks'
import { showToast } from '../components/Toast'
import type { RecommendedFeed } from '../types'

export default function Onboarding() {
  const navigate = useNavigate()
  const { data: feeds = [], isLoading } = useRecommendedFeeds()
  const batchCreate = useBatchCreateFeeds()
  const completeOnboarding = useCompleteOnboarding()
  const [selectedFeeds, setSelectedFeeds] = useState<Set<string>>(new Set())

  // Group feeds by category
  const feedsByCategory = useMemo(() => {
    const grouped: Record<string, RecommendedFeed[]> = {}
    feeds.forEach((feed) => {
      if (!grouped[feed.category]) {
        grouped[feed.category] = []
      }
      grouped[feed.category].push(feed)
    })
    return grouped
  }, [feeds])

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
    setSelectedFeeds(new Set(feeds.map((f) => f.id)))
  }

  const selectNone = () => {
    setSelectedFeeds(new Set())
  }

  const handleSubmit = async () => {
    try {
      if (selectedFeeds.size > 0) {
        const urls = feeds
          .filter((f) => selectedFeeds.has(f.id))
          .map((f) => f.url)
        await batchCreate.mutateAsync(urls)
        showToast('success', `Added ${selectedFeeds.size} feeds`)
      }
      // Mark onboarding as complete in backend
      await completeOnboarding.mutateAsync()
      navigate('/')
    } catch (err) {
      showToast('error', 'Failed to complete setup')
    }
  }

  const handleSkip = async () => {
    try {
      await completeOnboarding.mutateAsync()
      navigate('/')
    } catch (err) {
      showToast('error', 'Failed to skip onboarding')
    }
  }

  if (isLoading) {
    return (
      <div className="min-h-screen bg-[var(--bg-primary)] flex items-center justify-center">
        <Loader2 className="w-8 h-8 animate-spin text-[var(--accent)]" />
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-[var(--bg-primary)] py-8 px-4">
      <div className="max-w-4xl mx-auto">
        {/* Header */}
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center w-16 h-16 bg-[var(--accent)] rounded-2xl mb-4">
            <Rss className="w-8 h-8 text-white" />
          </div>
          <h1 className="text-3xl font-bold text-[var(--text-primary)] mb-2">
            Welcome to Fread
          </h1>
          <p className="text-[var(--text-secondary)] max-w-md mx-auto">
            Get started by selecting some feeds to follow. You can always add more later.
          </p>
        </div>

        {/* Selection controls */}
        <div className="flex items-center justify-between mb-6">
          <div className="flex items-center gap-4">
            <button
              onClick={selectAll}
              className="text-sm text-[var(--accent)] hover:underline"
            >
              Select All
            </button>
            <button
              onClick={selectNone}
              className="text-sm text-[var(--text-secondary)] hover:underline"
            >
              Clear
            </button>
          </div>
          <span className="text-sm text-[var(--text-muted)]">
            {selectedFeeds.size} selected
          </span>
        </div>

        {/* Feed categories */}
        <div className="space-y-8 mb-8">
          {Object.entries(feedsByCategory).map(([category, categoryFeeds]) => (
            <div key={category}>
              <h2 className="text-lg font-semibold text-[var(--text-primary)] mb-4">
                {category}
              </h2>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                {categoryFeeds.map((feed) => (
                  <FeedCard
                    key={feed.id}
                    feed={feed}
                    selected={selectedFeeds.has(feed.id)}
                    onToggle={() => toggleFeed(feed.id)}
                  />
                ))}
              </div>
            </div>
          ))}
        </div>

        {/* Actions */}
        <div className="flex items-center justify-between pt-6 border-t border-[var(--border-color)]">
          <button
            onClick={handleSkip}
            className="text-[var(--text-secondary)] hover:text-[var(--text-primary)] transition-colors"
          >
            Skip for now
          </button>
          <button
            onClick={handleSubmit}
            disabled={batchCreate.isPending}
            className="flex items-center gap-2 px-6 py-3 bg-[var(--accent)] hover:opacity-90 text-white font-medium rounded-lg transition-colors disabled:opacity-50"
          >
            {batchCreate.isPending ? (
              <>
                <Loader2 className="w-4 h-4 animate-spin" />
                Adding...
              </>
            ) : (
              <>
                {selectedFeeds.size > 0 ? `Add ${selectedFeeds.size} Feeds` : 'Continue'}
                <ChevronRight className="w-4 h-4" />
              </>
            )}
          </button>
        </div>
      </div>
    </div>
  )
}

function FeedCard({
  feed,
  selected,
  onToggle,
}: {
  feed: RecommendedFeed
  selected: boolean
  onToggle: () => void
}) {
  return (
    <button
      onClick={onToggle}
      className={`flex items-start gap-3 p-4 rounded-xl border transition-all text-left ${
        selected
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

      {/* Checkbox */}
      <div
        className={`w-5 h-5 rounded-full border-2 flex items-center justify-center flex-shrink-0 transition-colors ${
          selected
            ? 'border-[var(--accent)] bg-[var(--accent)]'
            : 'border-[var(--text-muted)]'
        }`}
      >
        {selected && <Check className="w-3 h-3 text-white" />}
      </div>
    </button>
  )
}
