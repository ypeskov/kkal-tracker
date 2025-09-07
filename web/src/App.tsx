import { useState, useEffect } from 'react'
import { useQuery } from '@tanstack/react-query'
import Login from './components/Login'
import Dashboard from './components/Dashboard'
import { authService } from './api/auth'
import './App.css'

interface User {
  id: number
  email: string
}

function App() {
  const [isAuthenticated, setIsAuthenticated] = useState(false)
  const [isInitializing, setIsInitializing] = useState(true)

  useEffect(() => {
    // Check for existing token on app startup
    const token = authService.getToken()
    if (token) {
      setIsAuthenticated(true)
    }
    setIsInitializing(false)
  }, [])

  const { data: user, isLoading, error } = useQuery<User>({
    queryKey: ['user'],
    queryFn: authService.getCurrentUser,
    retry: false,
    enabled: isAuthenticated,
  })

  // Handle token validation errors
  useEffect(() => {
    if (error && isAuthenticated) {
      authService.logout()
      setIsAuthenticated(false)
    }
  }, [error, isAuthenticated])

  const handleLogout = () => {
    authService.logout()
    setIsAuthenticated(false)
  }

  if (isInitializing || isLoading) {
    return <div>Loading...</div>
  }

  return (
    <div className="app">
      {!isAuthenticated ? (
        <Login onLogin={() => setIsAuthenticated(true)} />
      ) : (
        <Dashboard user={user} onLogout={handleLogout} />
      )}
    </div>
  )
}

export default App