import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState, useMemo } from 'react'
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
  const [foodName, setFoodName] = useState('')
  const [weight, setWeight] = useState('')
  const [kcalPer100g, setKcalPer100g] = useState('')
  const queryClient = useQueryClient()

  // Auto-calculate total calories
  const totalCalories = useMemo(() => {
    const weightNum = parseFloat(weight) || 0
    const kcalNum = parseFloat(kcalPer100g) || 0
    return Math.round((weightNum * kcalNum) / 100)
  }, [weight, kcalPer100g])

  const { data: entries, isLoading } = useQuery({
    queryKey: ['calories'],
    queryFn: calorieService.getEntries,
  })

  const addEntryMutation = useMutation({
    mutationFn: calorieService.addEntry,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['calories'] })
      setFoodName('')
      setWeight('')
      setKcalPer100g('')
    },
  })

  const isButtonDisabled = addEntryMutation.isPending || !foodName || !weight || !kcalPer100g

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!foodName || !weight || !kcalPer100g) return
    
    addEntryMutation.mutate({
      food: foodName,
      weight: parseFloat(weight),
      kcalPer100g: parseFloat(kcalPer100g),
      calories: totalCalories,
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
        <form onSubmit={handleSubmit} style={{ display: 'flex', gap: '1rem', alignItems: 'end', flexWrap: 'wrap' }}>
          <div className="form-group">
            <label htmlFor="foodName">{t('dashboard.foodName')}:</label>
            <input
              type="text"
              id="foodName"
              value={foodName}
              onChange={(e) => setFoodName(e.target.value)}
              required
              style={{ minWidth: '150px' }}
            />
          </div>

          <div className="form-group">
            <label htmlFor="weight">{t('dashboard.weight')}:</label>
            <input
              type="number"
              id="weight"
              value={weight}
              onChange={(e) => setWeight(e.target.value)}
              required
              min="0"
              step="0.1"
              style={{ width: '100px' }}
            />
          </div>

          <div className="form-group">
            <label htmlFor="kcalPer100g">{t('dashboard.kcalPer100g')}:</label>
            <input
              type="number"
              id="kcalPer100g"
              value={kcalPer100g}
              onChange={(e) => setKcalPer100g(e.target.value)}
              required
              min="0"
              step="0.1"
              style={{ width: '100px' }}
            />
          </div>

          <div className="form-group">
            <label htmlFor="totalCalories">{t('dashboard.totalCalories')}:</label>
            <input
              type="number"
              id="totalCalories"
              value={totalCalories}
              readOnly
              style={{ 
                width: '100px', 
                backgroundColor: '#f5f5f5',
                border: '1px solid #ddd',
                cursor: 'not-allowed'
              }}
            />
          </div>

          <button 
            type="submit" 
            className="btn"
            disabled={isButtonDisabled}
            style={{ 
              minWidth: '80px',
              alignSelf: 'flex-end'
            }}
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
                      border: '1px solid #eee'
                    }}
                  >
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.5rem' }}>
                      <strong>{entry.food}</strong>
                      <strong>{entry.calories} {t('dashboard.kcal')}</strong>
                    </div>
                    <div style={{ display: 'flex', gap: '1rem', fontSize: '0.9em', color: '#666' }}>
                      <span>{t('dashboard.weight')}: {entry.weight}g</span>
                      <span>{t('dashboard.kcalPer100g')}: {entry.kcalPer100g}</span>
                    </div>
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