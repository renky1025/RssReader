import { useEffect } from 'react'
import { useAppStore } from '../stores/appStore'
import Sidebar from '../components/Sidebar'
import ArticleList from '../components/ArticleList'
import ArticleReader from '../components/ArticleReader'
import Discover from '../pages/Discover'
import { useKeyboardShortcuts } from '../hooks/useKeyboardShortcuts'
import { Menu, ArrowLeft } from 'lucide-react'

export default function MainLayout() {
  const { 
    selectedArticle, 
    sidebarCollapsed, 
    currentView, 
    isMobile, 
    setIsMobile, 
    mobileView, 
    setMobileView,
    setSelectedArticle 
  } = useAppStore()
  
  // Enable keyboard shortcuts
  useKeyboardShortcuts()

  // Handle window resize for mobile detection
  useEffect(() => {
    const handleResize = () => {
      const mobile = window.innerWidth < 768
      setIsMobile(mobile)
      // Reset to list view when switching from mobile to desktop
      if (!mobile && mobileView !== 'list') {
        setMobileView('list')
      }
    }
    
    window.addEventListener('resize', handleResize)
    handleResize() // Initial check
    return () => window.removeEventListener('resize', handleResize)
  }, [setIsMobile, setMobileView, mobileView])

  // Auto-switch to reader view when article is selected on mobile
  useEffect(() => {
    if (isMobile && selectedArticle) {
      setMobileView('reader')
    }
  }, [selectedArticle, isMobile, setMobileView])

  // Mobile Layout
  if (isMobile) {
    return (
      <div className="h-screen bg-[var(--bg-primary)] text-[var(--text-primary)] overflow-hidden">
        {/* Mobile Header */}
        <div className="h-14 flex items-center justify-between px-4 border-b border-[var(--border-color)] bg-[var(--bg-secondary)]">
          {mobileView === 'list' && (
            <>
              <button
                onClick={() => setMobileView('sidebar')}
                className="p-2 hover:bg-[var(--bg-hover)] rounded-lg -ml-2"
              >
                <Menu className="w-5 h-5" />
              </button>
              <h1 className="text-lg font-semibold truncate flex-1 text-center">{currentView.title}</h1>
              <div className="w-9" /> {/* Spacer for centering */}
            </>
          )}
          {mobileView === 'reader' && (
            <>
              <button
                onClick={() => {
                  setMobileView('list')
                  setSelectedArticle(null)
                }}
                className="p-2 hover:bg-[var(--bg-hover)] rounded-lg -ml-2 flex items-center gap-1"
              >
                <ArrowLeft className="w-5 h-5" />
                <span className="text-sm">Back</span>
              </button>
              <div className="flex-1" />
            </>
          )}
          {mobileView === 'sidebar' && (
            <>
              <button
                onClick={() => setMobileView('list')}
                className="p-2 hover:bg-[var(--bg-hover)] rounded-lg -ml-2"
              >
                <ArrowLeft className="w-5 h-5" />
              </button>
              <h1 className="text-lg font-semibold flex-1 text-center">Menu</h1>
              <div className="w-9" />
            </>
          )}
        </div>

        {/* Mobile Content */}
        <div className="h-[calc(100vh-3.5rem)] overflow-hidden">
          {mobileView === 'sidebar' && (
            <div className="h-full">
              <Sidebar onNavigate={() => setMobileView('list')} />
            </div>
          )}
          {mobileView === 'list' && (
            currentView.type === 'discover' ? (
              <Discover />
            ) : (
              <ArticleList />
            )
          )}
          {mobileView === 'reader' && selectedArticle && (
            <ArticleReader article={selectedArticle} />
          )}
        </div>
      </div>
    )
  }

  // Desktop Layout - Show Discover page when in discover view
  if (currentView.type === 'discover') {
    return (
      <div className="flex h-screen bg-[var(--bg-primary)] text-[var(--text-primary)] overflow-hidden">
        {/* Sidebar */}
        <div className={`${sidebarCollapsed ? 'w-16' : 'w-64'} flex-shrink-0 transition-all duration-300`}>
          <Sidebar />
        </div>

        {/* Discover Page */}
        <Discover />
      </div>
    )
  }

  // Desktop Layout - Normal view
  return (
    <div className="flex h-screen bg-[var(--bg-primary)] text-[var(--text-primary)] overflow-hidden">
      {/* Sidebar */}
      <div className={`${sidebarCollapsed ? 'w-16' : 'w-64'} flex-shrink-0 transition-all duration-300`}>
        <Sidebar />
      </div>

      {/* Article List */}
      <div className="w-96 flex-shrink-0 border-l border-[var(--border-color)]">
        <ArticleList />
      </div>

      {/* Article Reader */}
      <div className="flex-1 border-l border-[var(--border-color)] overflow-hidden">
        {selectedArticle ? (
          <ArticleReader article={selectedArticle} />
        ) : (
          <div className="h-full flex items-center justify-center text-[var(--text-muted)]">
            <p>Select an article to read</p>
          </div>
        )}
      </div>
    </div>
  )
}
