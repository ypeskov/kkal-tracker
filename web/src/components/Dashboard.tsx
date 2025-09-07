import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { calorieService } from '../api/calories'
import LanguageSwitcher from './LanguageSwitcher'

interface User {
  id: number
  email: string
}

interface DashboardProps {
  user?: User
  onLogout: () => void
}

export default function Dashboard({ user, onLogout }: DashboardProps) {
  const { t } = useTranslation()
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
      <LanguageSwitcher />
      <header style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div>
          <h1>{t('dashboard.title')}</h1>
          <p>{t('auth.welcome')}, {user?.email}!</p>
        </div>
        <button 
          onClick={onLogout}
          className="btn"
          style={{ backgroundColor: '#dc3545' }}
        >
          {t('auth.logout')}
        </button>
      </header>

      <section style={{ marginBottom: '2rem' }}>
        <h2>{t('dashboard.addFoodEntry')}</h2>
        <form onSubmit={handleSubmit} style={{ display: 'flex', gap: '1rem', alignItems: 'end' }}>
          <div className="form-group">
            <label htmlFor="food">{t('dashboard.food')}:</label>
            <input
              type="text"
              id="food"
              value={food}
              onChange={(e) => setFood(e.target.value)}
              required
            />
          </div>

          <div className="form-group">
            <label htmlFor="calories">{t('dashboard.calories')}:</label>
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
            {t('dashboard.addEntry')}
          </button>
        </form>
      </section>

      <section>
        <h2>{t('dashboard.todayEntries')}</h2>
        {isLoading ? (
          <p>{t('common.loading')}</p>
        ) : (
          <div>
            {entries?.length === 0 ? (
              <p>{t('dashboard.noEntries')}</p>
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
                    <span>{entry.calories} {t('dashboard.kcal')}</span>
                  </li>
                ))}
              </ul>
            )}
            {entries && entries.length > 0 && (
              <div style={{ marginTop: '1rem', fontWeight: 'bold' }}>
                {t('dashboard.total')}: {entries.reduce((sum: number, entry: any) => sum + entry.calories, 0)} {t('dashboard.kcal')}
              </div>
            )}
          </div>
        )}
      </section>
    </div>
  )
}