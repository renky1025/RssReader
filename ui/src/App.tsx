import { useEffect } from 'react'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { useAppStore } from './stores/appStore'
import Login from './pages/Login'
import Register from './pages/Register'
import Onboarding from './pages/Onboarding'
import Admin from './pages/Admin'
import MainLayout from './layouts/MainLayout'
import { ToastContainer } from './components/Toast'
import { useCurrentUser } from './api/hooks'

function PrivateRoute({ children }: { children: React.ReactNode }) {
  const isAuthenticated = useAppStore((state) => state.isAuthenticated)
  return isAuthenticated ? <>{children}</> : <Navigate to="/login" replace />
}

function AdminRoute({ children }: { children: React.ReactNode }) {
  const isAuthenticated = useAppStore((state) => state.isAuthenticated)
  const { data: user } = useCurrentUser()
  
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }
  
  // Wait for user data to load
  if (!user) {
    return <div className="min-h-screen bg-[var(--bg-primary)] flex items-center justify-center">
      <div className="text-[var(--text-secondary)]">Loading...</div>
    </div>
  }
  
  // Check if user is admin
  if (!user.is_admin) {
    return <Navigate to="/" replace />
  }
  
  return <>{children}</>
}

function OnboardingCheck({ children }: { children: React.ReactNode }) {
  const { data: user, isLoading } = useCurrentUser()
  
  if (isLoading) {
    return <div className="min-h-screen bg-[var(--bg-primary)] flex items-center justify-center">
      <div className="text-[var(--text-secondary)]">Loading...</div>
    </div>
  }
  
  // Only redirect to onboarding if user hasn't completed it
  if (user && !user.onboarding_complete) {
    return <Navigate to="/onboarding" replace />
  }
  
  return <>{children}</>
}

export default function App() {
  const theme = useAppStore((state) => state.theme)

  // Apply theme on mount and when theme changes
  useEffect(() => {
    if (theme === 'dark') {
      document.documentElement.classList.add('dark')
    } else {
      document.documentElement.classList.remove('dark')
    }
  }, [theme])

  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />
        <Route
          path="/admin"
          element={
            <AdminRoute>
              <Admin />
            </AdminRoute>
          }
        />
        <Route
          path="/onboarding"
          element={
            <PrivateRoute>
              <Onboarding />
            </PrivateRoute>
          }
        />
        <Route
          path="/*"
          element={
            <PrivateRoute>
              <OnboardingCheck>
                <MainLayout />
              </OnboardingCheck>
            </PrivateRoute>
          }
        />
      </Routes>
      <ToastContainer />
    </BrowserRouter>
  )
}
