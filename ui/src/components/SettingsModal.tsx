import { useState, useRef } from 'react'
import { X, Upload, Download, Trash2 } from 'lucide-react'
import { useImportOPML, useDeleteFeed, useFeeds } from '../api/hooks'
import { showToast } from './Toast'

interface SettingsModalProps {
  isOpen: boolean
  onClose: () => void
}

export default function SettingsModal({ isOpen, onClose }: SettingsModalProps) {
  const [activeTab, setActiveTab] = useState<'general' | 'feeds' | 'import'>('general')
  const fileInputRef = useRef<HTMLInputElement>(null)
  const importOPML = useImportOPML()
  const deleteFeed = useDeleteFeed()
  const { data: feeds = [] } = useFeeds()

  if (!isOpen) return null

  const handleImportOPML = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return

    try {
      await importOPML.mutateAsync(file)
      showToast('success', 'OPML imported successfully')
      if (fileInputRef.current) {
        fileInputRef.current.value = ''
      }
    } catch (err) {
      showToast('error', 'Failed to import OPML')
    }
  }

  const handleExportOPML = () => {
    const token = localStorage.getItem('token')
    window.open(`/api/v1/opml/export?token=${token}`, '_blank')
  }

  const handleDeleteFeed = async (id: number, title: string) => {
    if (!confirm(`Delete feed "${title}"?`)) return
    try {
      await deleteFeed.mutateAsync(id)
      showToast('success', 'Feed deleted')
    } catch (err) {
      showToast('error', 'Failed to delete feed')
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4 sm:p-6">
      <div className="w-full max-w-2xl max-h-[85vh] overflow-hidden rounded-2xl border border-slate-200 bg-slate-50 p-2 text-slate-900 shadow-2xl dark:border-[#30363d] dark:bg-[#1a1d21] dark:text-slate-100">
        {/* Header */}
        <div className="flex items-center justify-between border-b border-slate-200 px-4 py-4 sm:px-5 dark:border-[#30363d]">
          <h2 className="text-xl font-semibold text-slate-900 dark:text-slate-100">Settings</h2>
          <button
            onClick={onClose}
            className="rounded-lg p-2 !text-slate-600 transition-colors hover:bg-slate-200 hover:!text-slate-900 dark:!text-slate-300 dark:hover:bg-[#2d333b] dark:hover:!text-slate-100"
          >
            <X className="w-5 h-5 !text-inherit" />
          </button>
        </div>

        {/* Tabs */}
        <div className="flex border-b border-slate-200 px-1 dark:border-[#30363d]">
          <button
            onClick={() => setActiveTab('general')}
            className={`px-4 py-3 text-sm font-medium transition-colors ${
              activeTab === 'general'
                ? '!text-[var(--accent)] border-b-2 border-[var(--accent)] dark:!text-[#f7931e]'
                : '!text-slate-600 hover:!text-slate-900 dark:!text-slate-300 dark:hover:!text-slate-100'
            }`}
          >
            General
          </button>
          <button
            onClick={() => setActiveTab('feeds')}
            className={`px-4 py-3 text-sm font-medium transition-colors ${
              activeTab === 'feeds'
                ? '!text-[var(--accent)] border-b-2 border-[var(--accent)] dark:!text-[#f7931e]'
                : '!text-slate-600 hover:!text-slate-900 dark:!text-slate-300 dark:hover:!text-slate-100'
            }`}
          >
            Manage Feeds
          </button>
          <button
            onClick={() => setActiveTab('import')}
            className={`px-4 py-3 text-sm font-medium transition-colors ${
              activeTab === 'import'
                ? '!text-[var(--accent)] border-b-2 border-[var(--accent)] dark:!text-[#f7931e]'
                : '!text-slate-600 hover:!text-slate-900 dark:!text-slate-300 dark:hover:!text-slate-100'
            }`}
          >
            Import / Export
          </button>
        </div>

        {/* Content */}
        <div className="max-h-[62vh] overflow-y-auto px-4 py-4 sm:px-5 sm:py-5">
          {activeTab === 'general' && (
            <div className="space-y-4">
              <div>
                <h3 className="mb-2 text-sm font-medium text-slate-700 dark:text-slate-200">About</h3>
                <p className="text-sm text-slate-700 dark:text-slate-200">
                  Fread v1.0.0
                </p>
                <p className="mt-1 text-xs text-slate-500 dark:text-slate-400">
                  A self-hosted RSS reader built with Go and React
                </p>
              </div>

              <div>
                <h3 className="mb-2 text-sm font-medium text-slate-700 dark:text-slate-200">Sync Settings</h3>
                <p className="text-sm text-slate-700 dark:text-slate-200">
                  Feeds are automatically synced every 15 minutes.
                </p>
              </div>

              <div>
                <h3 className="mb-2 text-sm font-medium text-slate-700 dark:text-slate-200">Keyboard Shortcuts</h3>
                <div className="grid grid-cols-2 gap-2 text-sm">
                  <div className="flex justify-between text-slate-700 dark:text-slate-200">
                    <span>Next article</span>
                    <kbd className="rounded bg-slate-200 px-2 py-0.5 text-xs text-slate-700 dark:bg-[#2d333b] dark:text-slate-200">j</kbd>
                  </div>
                  <div className="flex justify-between text-slate-700 dark:text-slate-200">
                    <span>Previous article</span>
                    <kbd className="rounded bg-slate-200 px-2 py-0.5 text-xs text-slate-700 dark:bg-[#2d333b] dark:text-slate-200">k</kbd>
                  </div>
                  <div className="flex justify-between text-slate-700 dark:text-slate-200">
                    <span>Toggle read</span>
                    <kbd className="rounded bg-slate-200 px-2 py-0.5 text-xs text-slate-700 dark:bg-[#2d333b] dark:text-slate-200">m</kbd>
                  </div>
                  <div className="flex justify-between text-slate-700 dark:text-slate-200">
                    <span>Toggle star</span>
                    <kbd className="rounded bg-slate-200 px-2 py-0.5 text-xs text-slate-700 dark:bg-[#2d333b] dark:text-slate-200">s</kbd>
                  </div>
                </div>
              </div>
            </div>
          )}

          {activeTab === 'feeds' && (
            <div className="space-y-2">
              {feeds.length === 0 ? (
                <p className="py-8 text-center text-slate-600 dark:text-slate-300">No feeds yet</p>
              ) : (
                feeds.map((feed) => (
                  <div
                    key={feed.id}
                    className="flex items-center justify-between rounded-lg bg-white p-3 dark:bg-[#1e2227]"
                  >
                    <div className="flex-1 min-w-0">
                      <p className="truncate text-sm text-slate-900 dark:text-slate-100">{feed.title || feed.url}</p>
                      <p className="truncate text-xs text-slate-500 dark:text-slate-400">{feed.url}</p>
                    </div>
                    <button
                      onClick={() => handleDeleteFeed(feed.id, feed.title || feed.url)}
                      className="rounded-lg p-2 !text-slate-600 transition-colors hover:bg-slate-200 hover:!text-red-500 dark:!text-slate-300 dark:hover:bg-[#2d333b]"
                    >
                      <Trash2 className="w-4 h-4 !text-inherit" />
                    </button>
                  </div>
                ))
              )}
            </div>
          )}

          {activeTab === 'import' && (
            <div className="space-y-6">
              <div>
                <h3 className="mb-3 text-sm font-medium text-slate-700 dark:text-slate-200">Import OPML</h3>
                <p className="mb-3 text-sm text-slate-700 dark:text-slate-200">
                  Import your subscriptions from another RSS reader.
                </p>
                <label className="flex w-fit cursor-pointer items-center gap-2 rounded-lg bg-[var(--accent)] px-4 py-2 !text-white transition-colors hover:opacity-90">
                  <Upload className="w-4 h-4 !text-inherit" />
                  <span className="text-sm">Choose OPML File</span>
                  <input
                    ref={fileInputRef}
                    type="file"
                    accept=".opml,.xml"
                    onChange={handleImportOPML}
                    className="hidden"
                  />
                </label>
              </div>

              <div>
                <h3 className="mb-3 text-sm font-medium text-slate-700 dark:text-slate-200">Export OPML</h3>
                <p className="mb-3 text-sm text-slate-700 dark:text-slate-200">
                  Export your subscriptions to use in another RSS reader.
                </p>
                <button
                  onClick={handleExportOPML}
                  className="flex items-center gap-2 rounded-lg bg-slate-200 px-4 py-2 !text-slate-800 transition-colors hover:bg-slate-300 dark:bg-[#2d333b] dark:!text-slate-100 dark:hover:bg-[#38404b]"
                >
                  <Download className="w-4 h-4 !text-inherit" />
                  <span className="text-sm">Download OPML</span>
                </button>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
