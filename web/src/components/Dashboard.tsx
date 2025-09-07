import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState, useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { calorieService } from '../api/calories'
import LanguageSwitcher from './LanguageSwitcher'
import './Dashboard.css'

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
  const [fats, setFats] = useState('')
  const [carbs, setCarbs] = useState('')
  const [proteins, setProteins] = useState('')
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
      setFats('')
      setCarbs('')
      setProteins('')
    },
  })

  const isButtonDisabled = addEntryMutation.isPending || !foodName || !weight || !kcalPer100g

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!foodName || !weight || !kcalPer100g) return
    
    const entryData: any = {
      food: foodName,
      weight: parseFloat(weight),
      kcalPer100g: parseFloat(kcalPer100g),
      calories: totalCalories,
      meal_datetime: new Date().toISOString(),
    }
    
    // Only include nutritional fields if they have values
    if (fats && parseFloat(fats) > 0) entryData.fats = parseFloat(fats)
    if (carbs && parseFloat(carbs) > 0) entryData.carbs = parseFloat(carbs)
    if (proteins && parseFloat(proteins) > 0) entryData.proteins = parseFloat(proteins)
    
    addEntryMutation.mutate(entryData)
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
        <form onSubmit={handleSubmit} className="food-form">
          <div className="form-row">
            <div className="form-group">
              <label htmlFor="foodName">{t('dashboard.foodName')}:</label>
              <input
                type="text"
                id="foodName"
                value={foodName}
                onChange={(e) => setFoodName(e.target.value)}
                required
              />
            </div>
          </div>
          
          <div className="form-row">
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
              />
            </div>
          </div>

          <div className="form-row">
            <div className="form-group">
              <label htmlFor="fats">{t('dashboard.fats')}:</label>
              <input
                type="number"
                id="fats"
                value={fats}
                onChange={(e) => setFats(e.target.value)}
                min="0"
                step="0.1"
              />
            </div>

            <div className="form-group">
              <label htmlFor="carbs">{t('dashboard.carbs')}:</label>
              <input
                type="number"
                id="carbs"
                value={carbs}
                onChange={(e) => setCarbs(e.target.value)}
                min="0"
                step="0.1"
              />
            </div>

            <div className="form-group">
              <label htmlFor="proteins">{t('dashboard.proteins')}:</label>
              <input
                type="number"
                id="proteins"
                value={proteins}
                onChange={(e) => setProteins(e.target.value)}
                min="0"
                step="0.1"
              />
            </div>
          </div>

          <div className="form-row">
            <div className="form-group">
              <label htmlFor="totalCalories">{t('dashboard.totalCalories')}:</label>
              <input
                type="number"
                id="totalCalories"
                value={totalCalories}
                readOnly
                className="readonly-field"
              />
            </div>

            <button 
              type="submit" 
              className="btn submit-btn"
              disabled={isButtonDisabled}
            >
              {t('dashboard.addEntry')}
            </button>
          </div>
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
                  <li key={entry.id} className="entry-item">
                    <div className="entry-header">
                      <strong>{entry.food}</strong>
                      <strong>{entry.calories} {t('dashboard.kcal')}</strong>
                    </div>
                    <div className="entry-details">
                      <span>{t('dashboard.weight')}: {entry.weight}g</span>
                      <span>{t('dashboard.kcalPer100g')}: {entry.kcalPer100g}</span>
                      {entry.fats && <span>{t('dashboard.fats')}: {entry.fats}g</span>}
                      {entry.carbs && <span>{t('dashboard.carbs')}: {entry.carbs}g</span>}
                      {entry.proteins && <span>{t('dashboard.proteins')}: {entry.proteins}g</span>}
                    </div>
                  </li>
                ))}
              </ul>
            )}
            {entries && entries.length > 0 && (
              <div className="total-calories">
                {t('dashboard.total')}: {entries.reduce((sum: number, entry: any) => sum + entry.calories, 0)} {t('dashboard.kcal')}
              </div>
            )}
          </div>
        )}
      </section>
    </div>
  )
}