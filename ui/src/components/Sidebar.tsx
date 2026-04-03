import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  Rss,
  Inbox,
  Star,
  Clock,
  FolderOpen,
  ChevronDown,
  ChevronRight,
  Plus,
  Settings,
  LogOut,
  Sun,
  Moon,
  Shield,
  Compass,
} from 'lucide-react'
import { useFeeds, useFolders, useStats, useCreateFeed, useCreateFolder, useCurrentUser, useUpdateFeed } from '../api/hooks'
import { useAppStore } from '../stores/appStore'
import { apiClient } from '../api/client'
import { showToast } from './Toast'
import SettingsModal from './SettingsModal'
import type { Feed, Folder } from '../types'

interface SidebarProps {
  onNavigate?: () => void
}

export default function Sidebar({ onNavigate }: SidebarProps) {
  const navigate = useNavigate()
  const { currentView, setCurrentView, sidebarCollapsed, toggleSidebar, theme, setTheme, setAuthenticated } = useAppStore()
  const { data: feedsData } = useFeeds()
  const { data: foldersData } = useFolders()
  const { data: stats } = useStats()
  const { data: currentUser } = useCurrentUser()
  
  // Ensure arrays are never null
  const feeds = feedsData ?? []
  const folders = foldersData ?? []
  const createFeed = useCreateFeed()
  const createFolder = useCreateFolder()
  const updateFeed = useUpdateFeed()

  const [expandedFolders, setExpandedFolders] = useState<Set<number>>(new Set())
  const [showAddFeed, setShowAddFeed] = useState(false)
  const [newFeedUrl, setNewFeedUrl] = useState('')
  const [showAddFolder, setShowAddFolder] = useState(false)
  const [newFolderName, setNewFolderName] = useState('')
  const [showSettings, setShowSettings] = useState(false)
  const [draggedFeed, setDraggedFeed] = useState<Feed | null>(null)
  const [dragOverFolder, setDragOverFolder] = useState<number | null>(null)

  const toggleFolder = (id: number) => {
    const newExpanded = new Set(expandedFolders)
    if (newExpanded.has(id)) {
      newExpanded.delete(id)
    } else {
      newExpanded.add(id)
    }
    setExpandedFolders(newExpanded)
  }

  const handleAddFeed = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!newFeedUrl.trim()) return
    try {
      await createFeed.mutateAsync({ url: newFeedUrl })
      setNewFeedUrl('')
      setShowAddFeed(false)
      showToast('success', 'Feed added successfully')
    } catch (err) {
      showToast('error', 'Failed to add feed')
    }
  }

  const handleAddFolder = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!newFolderName.trim()) return
    try {
      await createFolder.mutateAsync({ name: newFolderName })
      setNewFolderName('')
      setShowAddFolder(false)
      showToast('success', 'Folder created')
    } catch (err) {
      showToast('error', 'Failed to create folder')
    }
  }

  const handleLogout = () => {
    apiClient.setToken(null)
    setAuthenticated(false)
    window.location.href = '/login'
  }

  // 拖拽处理函数
  const handleDragStart = (e: React.DragEvent, feed: Feed) => {
    setDraggedFeed(feed)
    e.dataTransfer.effectAllowed = 'move'
    e.dataTransfer.setData('text/plain', feed.id.toString())
  }

  const handleDragOver = (e: React.DragEvent, folderId?: number) => {
    e.preventDefault()
    e.dataTransfer.dropEffect = 'move'
    setDragOverFolder(folderId || null)
  }

  const handleDragLeave = (e: React.DragEvent) => {
    // 只有当离开整个容器时才清除高亮
    if (!e.currentTarget.contains(e.relatedTarget as Node)) {
      setDragOverFolder(null)
    }
  }

  const handleDrop = async (e: React.DragEvent, folderId?: number) => {
    e.preventDefault()
    setDragOverFolder(null)
    
    if (!draggedFeed) return

    try {
      await updateFeed.mutateAsync({
        id: draggedFeed.id,
        folder_id: folderId || null
      })
      showToast('success', `Feed moved to ${folderId ? 'folder' : 'root'}`)
    } catch (err) {
      showToast('error', 'Failed to move feed')
    } finally {
      setDraggedFeed(null)
    }
  }

  const handleDragEnd = () => {
    setDraggedFeed(null)
    setDragOverFolder(null)
  }

  const getFeedsInFolder = (folderId: number | null): Feed[] => {
    return feeds.filter((f) => f.folder_id === folderId)
  }

  const unassignedFeeds = feeds.filter((f) => !f.folder_id)
  const totalUnread = stats?.unread_articles || 0

  if (sidebarCollapsed) {
    return (
      <div className="h-full bg-[var(--bg-secondary)] flex flex-col items-center py-4">
        <button
          onClick={toggleSidebar}
          className="mb-4 rounded-lg p-2 hover:bg-[var(--bg-hover)]"
        >
          <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-slate-900 text-white ring-1 ring-slate-700/20 dark:bg-white dark:text-[#f7931e] dark:ring-white/20">
            <Rss className="w-5 h-5" />
          </div>
        </button>
        <button
          onClick={() => setCurrentView({ type: 'all', title: 'All Feeds' })}
          className={`p-2 rounded-lg mb-2 ${currentView.type === 'all' ? 'bg-[var(--accent)] text-white [&_svg]:!text-white' : 'text-slate-600 hover:bg-[var(--bg-hover)] hover:text-slate-900 dark:text-slate-300 dark:hover:text-slate-100 [&_svg]:!text-slate-500 dark:[&_svg]:!text-slate-300'}`}
        >
          <Inbox className="w-5 h-5" />
        </button>
        <button
          onClick={() => setCurrentView({ type: 'unread', title: 'Unread' })}
          className={`p-2 rounded-lg mb-2 ${currentView.type === 'unread' ? 'bg-[var(--accent)] text-white [&_svg]:!text-white' : 'text-slate-600 hover:bg-[var(--bg-hover)] hover:text-slate-900 dark:text-slate-300 dark:hover:text-slate-100 [&_svg]:!text-slate-500 dark:[&_svg]:!text-slate-300'}`}
        >
          <Rss className="w-5 h-5" />
        </button>
        <button
          onClick={() => setCurrentView({ type: 'starred', title: 'Starred' })}
          className={`p-2 rounded-lg mb-2 ${currentView.type === 'starred' ? 'bg-[var(--accent)] text-white [&_svg]:!text-white' : 'text-slate-600 hover:bg-[var(--bg-hover)] hover:text-slate-900 dark:text-slate-300 dark:hover:text-slate-100 [&_svg]:!text-slate-500 dark:[&_svg]:!text-slate-300'}`}
        >
          <Star className="w-5 h-5" />
        </button>
        <button
          onClick={() => setCurrentView({ type: 'read-later', title: 'Read Later' })}
          className={`p-2 rounded-lg ${currentView.type === 'read-later' ? 'bg-[var(--accent)] text-white [&_svg]:!text-white' : 'text-slate-600 hover:bg-[var(--bg-hover)] hover:text-slate-900 dark:text-slate-300 dark:hover:text-slate-100 [&_svg]:!text-slate-500 dark:[&_svg]:!text-slate-300'}`}
        >
          <Clock className="w-5 h-5" />
        </button>
      </div>
    )
  }

  return (
    <div className="h-full bg-[var(--bg-secondary)] flex flex-col">
      {/* Header */}
      <div className="p-4 flex items-center justify-between border-b border-[var(--border-color)]">
        <button onClick={toggleSidebar} className="flex items-center gap-2">
          <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-slate-900 text-white ring-1 ring-slate-700/20 dark:bg-white dark:text-[#f7931e] dark:ring-white/20">
            <Rss className="w-5 h-5" />
          </div>
          <span className="text-xl font-bold text-[var(--text-primary)]">Fread</span>
        </button>
      </div>

      {/* Navigation */}
      <div className="flex-1 overflow-y-auto p-2">
        {/* Quick filters */}
        <div className="mb-4">
          <NavItem
            icon={<Inbox className="w-4 h-4" />}
            label="All"
            count={totalUnread}
            active={currentView.type === 'all'}
            onClick={() => { setCurrentView({ type: 'all', title: 'All Feeds' }); onNavigate?.() }}
          />
          <NavItem
            icon={<Rss className="w-4 h-4" />}
            label="Unread"
            active={currentView.type === 'unread'}
            onClick={() => { setCurrentView({ type: 'unread', title: 'Unread' }); onNavigate?.() }}
          />
          <NavItem
            icon={<Star className="w-4 h-4" />}
            label="Starred"
            count={stats?.starred_articles}
            active={currentView.type === 'starred'}
            onClick={() => { setCurrentView({ type: 'starred', title: 'Starred' }); onNavigate?.() }}
          />
          <NavItem
            icon={<Clock className="w-4 h-4" />}
            label="Read Later"
            count={stats?.read_later_count}
            active={currentView.type === 'read-later'}
            onClick={() => { setCurrentView({ type: 'read-later', title: 'Read Later' }); onNavigate?.() }}
          />
          <NavItem
            icon={<Compass className="w-4 h-4" />}
            label="Discover"
            active={currentView.type === 'discover'}
            onClick={() => { setCurrentView({ type: 'discover', title: 'Discover Feeds' }); onNavigate?.() }}
          />
        </div>

        {/* Feeds section */}
        <div className="mb-2 flex items-center justify-between px-2">
          <span className="text-xs font-semibold text-gray-400 uppercase">Feeds</span>
          <button
            onClick={() => setShowAddFeed(true)}
            className="p-1 hover:bg-[var(--bg-hover)] rounded"
          >
            <Plus className="w-4 h-4 text-gray-400" />
          </button>
        </div>

        {/* Add feed form */}
        {showAddFeed && (
          <form onSubmit={handleAddFeed} className="px-2 mb-2">
            <input
              type="url"
              value={newFeedUrl}
              onChange={(e) => setNewFeedUrl(e.target.value)}
              placeholder="Feed URL..."
              className="w-full px-2 py-1 bg-[var(--bg-primary)] border border-[var(--border-color)] rounded text-sm text-[var(--text-primary)] focus:outline-none focus:border-[var(--accent)]"
              autoFocus
            />
          </form>
        )}

        {/* Folders */}
        {folders.map((folder: Folder) => (
          <div key={folder.id}>
            <div
              className={`flex items-center gap-2 px-2 py-1.5 rounded-lg cursor-pointer hover:bg-[var(--bg-hover)] transition-colors ${
                currentView.type === 'folder' && currentView.id === folder.id ? 'bg-[var(--bg-hover)]' : ''
              } ${dragOverFolder === folder.id ? 'bg-blue-100 dark:bg-blue-900 border-2 border-blue-400 border-dashed' : ''}`}
              onDragOver={(e) => handleDragOver(e, folder.id)}
              onDragLeave={handleDragLeave}
              onDrop={(e) => handleDrop(e, folder.id)}
            >
              <button onClick={() => toggleFolder(folder.id)} className="p-0.5">
                {expandedFolders.has(folder.id) ? (
                  <ChevronDown className="w-4 h-4 text-gray-400" />
                ) : (
                  <ChevronRight className="w-4 h-4 text-gray-400" />
                )}
              </button>
              <FolderOpen className="w-4 h-4 text-amber-600 dark:!text-amber-300" />
              <span
                className="flex-1 text-sm text-[var(--text-secondary)] truncate"
                onClick={() => { setCurrentView({ type: 'folder', id: folder.id, title: folder.name }); onNavigate?.() }}
              >
                {folder.name}
              </span>
            </div>
            {expandedFolders.has(folder.id) && (
              <div className="ml-6">
                {getFeedsInFolder(folder.id).map((feed: Feed) => (
                  <FeedItem
                    key={feed.id}
                    feed={feed}
                    active={currentView.type === 'feed' && currentView.id === feed.id}
                    onClick={() => { setCurrentView({ type: 'feed', id: feed.id, title: feed.title }); onNavigate?.() }}
                    onDragStart={handleDragStart}
                    onDragEnd={handleDragEnd}
                    isDragging={draggedFeed?.id === feed.id}
                  />
                ))}
              </div>
            )}
          </div>
        ))}

        {/* Unassigned feeds */}
        <div
          className={`min-h-[2rem] rounded-lg transition-colors ${
            dragOverFolder === null && draggedFeed ? 'bg-blue-100 dark:bg-blue-900 border-2 border-blue-400 border-dashed' : ''
          }`}
          onDragOver={(e) => handleDragOver(e)}
          onDragLeave={handleDragLeave}
          onDrop={(e) => handleDrop(e)}
        >
          {unassignedFeeds.length === 0 && draggedFeed && (
            <div className="px-2 py-1 text-xs text-gray-400 text-center">
              Drop here to move to root
            </div>
          )}
          {unassignedFeeds.map((feed: Feed) => (
            <FeedItem
              key={feed.id}
              feed={feed}
              active={currentView.type === 'feed' && currentView.id === feed.id}
              onClick={() => { setCurrentView({ type: 'feed', id: feed.id, title: feed.title }); onNavigate?.() }}
              onDragStart={handleDragStart}
              onDragEnd={handleDragEnd}
              isDragging={draggedFeed?.id === feed.id}
            />
          ))}
        </div>

        {/* Add folder */}
        <div className="mt-4 mb-2 flex items-center justify-between px-2">
          <span className="text-xs font-semibold text-gray-400 uppercase">Folders</span>
          <button
            onClick={() => setShowAddFolder(true)}
            className="p-1 hover:bg-[var(--bg-hover)] rounded"
          >
            <Plus className="w-4 h-4 text-gray-400" />
          </button>
        </div>

        {showAddFolder && (
          <form onSubmit={handleAddFolder} className="px-2 mb-2">
            <input
              type="text"
              value={newFolderName}
              onChange={(e) => setNewFolderName(e.target.value)}
              placeholder="Folder name..."
              className="w-full px-2 py-1 bg-[var(--bg-primary)] border border-[var(--border-color)] rounded text-sm text-[var(--text-primary)] focus:outline-none focus:border-[var(--accent)]"
              autoFocus
            />
          </form>
        )}
      </div>

      {/* Footer */}
      <div className="p-2 border-t border-[var(--border-color)]">
        <div className="flex items-center justify-between">
          <button
            onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}
            className="p-2 hover:bg-[var(--bg-hover)] rounded-lg text-slate-600 hover:text-slate-900 dark:text-slate-300 dark:hover:text-slate-100 [&_svg]:!text-slate-500 dark:[&_svg]:!text-slate-300"
            title={theme === 'dark' ? 'Light mode' : 'Dark mode'}
          >
            {theme === 'dark' ? <Sun className="w-5 h-5" /> : <Moon className="w-5 h-5" />}
          </button>
          {currentUser?.is_admin && (
            <button 
              onClick={() => navigate('/admin')}
              className="p-2 hover:bg-[var(--bg-hover)] rounded-lg text-violet-500 dark:text-violet-400 [&_svg]:!text-violet-500 dark:[&_svg]:!text-violet-400"
              title="Admin Panel"
            >
              <Shield className="w-5 h-5" />
            </button>
          )}
          <button 
            onClick={() => setShowSettings(true)}
            className="p-2 hover:bg-[var(--bg-hover)] rounded-lg text-slate-600 hover:text-slate-900 dark:text-slate-300 dark:hover:text-slate-100 [&_svg]:!text-slate-500 dark:[&_svg]:!text-slate-300"
            title="Settings"
          >
            <Settings className="w-5 h-5" />
          </button>
          <button 
            onClick={handleLogout} 
            className="p-2 hover:bg-[var(--bg-hover)] rounded-lg text-slate-600 hover:text-slate-900 dark:text-slate-300 dark:hover:text-slate-100 [&_svg]:!text-slate-500 dark:[&_svg]:!text-slate-300"
            title="Logout"
          >
            <LogOut className="w-5 h-5" />
          </button>
        </div>
      </div>

      {/* Settings Modal */}
      <SettingsModal isOpen={showSettings} onClose={() => setShowSettings(false)} />
    </div>
  )
}

function NavItem({
  icon,
  label,
  count,
  active,
  onClick,
}: {
  icon: React.ReactNode
  label: string
  count?: number
  active: boolean
  onClick: () => void
}) {
  return (
    <button
      onClick={onClick}
      className={`w-full flex items-center gap-3 px-3 py-2 rounded-lg transition-colors ${
        active 
          ? 'sidebar-nav-item-active' 
          : 'text-[var(--text-secondary)] hover:bg-[var(--bg-hover)] hover:text-[var(--text-primary)] [&_svg]:!text-slate-500 dark:[&_svg]:!text-slate-300'
      }`}
    >
      {icon}
      <span className="flex-1 text-left text-sm">{label}</span>
      {count !== undefined && count > 0 && (
        <span className={`text-xs px-2 py-0.5 rounded-full ${active ? 'bg-white/20 text-white' : 'bg-gray-600 text-white'}`}>
          {count}
        </span>
      )}
    </button>
  )
}

function FeedItem({
  feed,
  active,
  onClick,
  onDragStart,
  onDragEnd,
  isDragging,
}: {
  feed: Feed
  active: boolean
  onClick: () => void
  onDragStart: (e: React.DragEvent, feed: Feed) => void
  onDragEnd: () => void
  isDragging: boolean
}) {
  return (
    <button
      onClick={onClick}
      draggable
      onDragStart={(e) => onDragStart(e, feed)}
      onDragEnd={onDragEnd}
      className={`w-full flex items-center gap-2 px-2 py-1.5 rounded-lg transition-colors cursor-move ${
        active ? 'bg-[var(--bg-hover)]' : 'hover:bg-[var(--bg-hover)]'
      } ${isDragging ? 'opacity-50' : ''}`}
      title="Drag to move to folder"
    >
      <Rss className="w-4 h-4 text-[#f7931e]" />
      <span className="flex-1 text-left text-sm text-[var(--text-secondary)] truncate">{feed.title || feed.url}</span>
      {feed.unread_count !== undefined && feed.unread_count > 0 && (
        <span className="text-xs px-1.5 py-0.5 bg-[var(--accent)] text-white rounded-full">
          {feed.unread_count}
        </span>
      )}
    </button>
  )
}
