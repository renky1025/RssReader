import { useEffect } from 'react'
import { useAppStore } from '../stores/appStore'
import { useUpdateArticle, useArticles } from '../api/hooks'

export function useKeyboardShortcuts() {
  const { currentView, selectedArticle, setSelectedArticle } = useAppStore()
  const updateArticle = useUpdateArticle()

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

  const { data } = useArticles(params)
  const articles = data?.items ?? []

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Ignore if typing in an input
      if (
        e.target instanceof HTMLInputElement ||
        e.target instanceof HTMLTextAreaElement
      ) {
        return
      }

      const currentIndex = selectedArticle
        ? articles.findIndex((a) => a.id === selectedArticle.id)
        : -1

      switch (e.key) {
        case 'j': // Next article
          e.preventDefault()
          if (currentIndex < articles.length - 1) {
            setSelectedArticle(articles[currentIndex + 1])
          } else if (articles.length > 0 && currentIndex === -1) {
            setSelectedArticle(articles[0])
          }
          break

        case 'k': // Previous article
          e.preventDefault()
          if (currentIndex > 0) {
            setSelectedArticle(articles[currentIndex - 1])
          }
          break

        case 'm': // Toggle read
          e.preventDefault()
          if (selectedArticle) {
            updateArticle.mutate({
              id: selectedArticle.id,
              is_read: !selectedArticle.is_read,
            })
          }
          break

        case 's': // Toggle star
          e.preventDefault()
          if (selectedArticle) {
            updateArticle.mutate({
              id: selectedArticle.id,
              is_starred: !selectedArticle.is_starred,
            })
          }
          break

        case 'b': // Toggle read later
          e.preventDefault()
          if (selectedArticle) {
            updateArticle.mutate({
              id: selectedArticle.id,
              is_read_later: !selectedArticle.is_read_later,
            })
          }
          break

        case 'o': // Open in new tab
        case 'Enter':
          e.preventDefault()
          if (selectedArticle?.url) {
            window.open(selectedArticle.url, '_blank')
          }
          break

        case 'Escape':
          e.preventDefault()
          setSelectedArticle(null)
          break
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [selectedArticle, articles, setSelectedArticle, updateArticle])
}
