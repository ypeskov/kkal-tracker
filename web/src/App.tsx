import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import Login from './components/Login'
import Dashboard from './components/Dashboard'
import { authService } from './api/auth'
import './App.css'

function App() {
  const [isAuthenticated, setIsAuthenticated] = useState(false)

  const { data: user, isLoading } = useQuery({
    queryKey: ['user'],
    queryFn: authService.getCurrentUser,
    retry: false,
    enabled: isAuthenticated,
  })

  if (isLoading) {
    return <div>Loading...</div>
  }

  return (
    <div className="app">
      {!isAuthenticated ? (
        <Login onLogin={() => setIsAuthenticated(true)} />
      ) : (
        <Dashboard user={user} />
      )}
    </div>
  )
}

export default App