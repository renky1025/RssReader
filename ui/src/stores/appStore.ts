import { create } from 'zustand'
import type { ViewState, Article } from '../types'

interface AppState {
  // Auth
  isAuthenticated: boolean
  setAuthenticated: (value: boolean) => void
  
  // View
  currentView: ViewState
  setCurrentView: (view: ViewState) => void
  
  // Selected article
  selectedArticle: Article | null
  setSelectedArticle: (article: Article | null) => void
  
  // UI state
  sidebarCollapsed: boolean
  toggleSidebar: () => void
  
  // Mobile state
  mobileView: 'sidebar' | 'list' | 'reader'
  setMobileView: (view: 'sidebar' | 'list' | 'reader') => void
  isMobile: boolean
  setIsMobile: (value: boolean) => void
  
  // Theme
  theme: 'dark' | 'light'
  setTheme: (theme: 'dark' | 'light') => void
}

export const useAppStore = create<AppState>((set) => ({
  // Auth
  isAuthenticated: !!localStorage.getItem('token'),
  setAuthenticated: (value) => set({ isAuthenticated: value }),
  
  // View
  currentView: { type: 'all', title: 'All Feeds' },
  setCurrentView: (view) => set({ currentView: view, selectedArticle: null }),
  
  // Selected article
  selectedArticle: null,
  setSelectedArticle: (article) => set({ selectedArticle: article }),
  
  // UI state
  sidebarCollapsed: false,
  toggleSidebar: () => set((state) => ({ sidebarCollapsed: !state.sidebarCollapsed })),
  
  // Mobile state
  mobileView: 'list',
  setMobileView: (view) => set({ mobileView: view }),
  isMobile: window.innerWidth < 768,
  setIsMobile: (value) => set({ isMobile: value }),
  
  // Theme
  theme: (localStorage.getItem('theme') as 'dark' | 'light') || 'dark',
  setTheme: (theme) => {
    localStorage.setItem('theme', theme)
    if (theme === 'dark') {
      document.documentElement.classList.add('dark')
    } else {
      document.documentElement.classList.remove('dark')
    }
    set({ theme })
  },
}))
