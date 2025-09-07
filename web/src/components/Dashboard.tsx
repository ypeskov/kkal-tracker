import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { calorieService } from '../api/calories'

interface User {
  id: number
  email: string
}

interface DashboardProps {
  user?: User
}

export default function Dashboard({ user }: DashboardProps) {
  const [food, setFood] = useState('')
  const [calories, setCalories] = useState('')
  const queryClient = useQueryClient()

  const { data: entries, isLoading } = useQuery({
    queryKey: ['calories'],
    queryFn: calorieService.getEntries,
  })

  const addEntryMutation = useMutation({
    mutationFn: calorieService.addEntry,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['calories'] })
      setFood('')
      setCalories('')
    },
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    addEntryMutation.mutate({
      food,
      calories: parseInt(calories),
      date: new Date().toISOString().split('T')[0],
    })
  }

  return (
    <div className="dashboard-container">
      <header>
        <h1>Calorie Tracker</h1>
        <p>Welcome, {user?.email}!</p>
      </header>

      <section style={{ marginBottom: '2rem' }}>
        <h2>Add Food Entry</h2>
        <form onSubmit={handleSubmit} style={{ display: 'flex', gap: '1rem', alignItems: 'end' }}>
          <div className="form-group">
            <label htmlFor="food">Food:</label>
            <input
              type="text"
              id="food"
              value={food}
              onChange={(e) => setFood(e.target.value)}
              required
            />
          </div>

          <div className="form-group">
            <label htmlFor="calories">Calories:</label>
            <input
              type="number"
              id="calories"
              value={calories}
              onChange={(e) => setCalories(e.target.value)}
              required
            />
          </div>

          <button 
            type="submit" 
            className="btn"
            disabled={addEntryMutation.isPending}
          >
            Add Entry
          </button>
        </form>
      </section>

      <section>
        <h2>Today's Entries</h2>
        {isLoading ? (
          <p>Loading entries...</p>
        ) : (
          <div>
            {entries?.length === 0 ? (
              <p>No entries yet today.</p>
            ) : (
              <ul style={{ listStyle: 'none', padding: 0 }}>
                {entries?.map((entry: any) => (
                  <li 
                    key={entry.id} 
                    style={{ 
                      background: 'white', 
                      padding: '1rem', 
                      marginBottom: '0.5rem', 
                      borderRadius: '4px',
                      display: 'flex',
                      justifyContent: 'space-between'
                    }}
                  >
                    <span>{entry.food}</span>
                    <span>{entry.calories} kcal</span>
                  </li>
                ))}
              </ul>
            )}
            {entries && entries.length > 0 && (
              <div style={{ marginTop: '1rem', fontWeight: 'bold' }}>
                Total: {entries.reduce((sum: number, entry: any) => sum + entry.calories, 0)} kcal
              </div>
            )}
          </div>
        )}
      </section>
    </div>
  )
}