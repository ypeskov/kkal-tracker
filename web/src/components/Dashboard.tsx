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

type FilterType = 'today' | 'lastWeek' | 'lastMonth' | 'customRange'

export default function Dashboard({ user, onLogout }: DashboardProps) {
  const { t } = useTranslation()
  const [foodName, setFoodName] = useState('')
  const [weight, setWeight] = useState('')
  const [kcalPer100g, setKcalPer100g] = useState('')
  const [fats, setFats] = useState('')
  const [carbs, setCarbs] = useState('')
  const [proteins, setProteins] = useState('')
  const [filterType, setFilterType] = useState<FilterType>('today')
  const [customDateFrom, setCustomDateFrom] = useState('')
  const [customDateTo, setCustomDateTo] = useState('')
  const queryClient = useQueryClient()

  // Auto-calculate total calories
  const totalCalories = useMemo(() => {
    const weightNum = parseFloat(weight) || 0
    const kcalNum = parseFloat(kcalPer100g) || 0
    return Math.round((weightNum * kcalNum) / 100)
  }, [weight, kcalPer100g])

  // Calculate date range based on filter type
  const getDateParams = useMemo(() => {
    const today = new Date().toISOString().split('T')[0]
    
    switch (filterType) {
      case 'today':
        return { date: today }
      case 'lastWeek':
        const weekAgo = new Date()
        weekAgo.setDate(weekAgo.getDate() - 7)
        return { 
          dateFrom: weekAgo.toISOString().split('T')[0], 
          dateTo: today 
        }
      case 'lastMonth':
        const monthAgo = new Date()
        monthAgo.setDate(monthAgo.getDate() - 30)
        return { 
          dateFrom: monthAgo.toISOString().split('T')[0], 
          dateTo: today 
        }
      case 'customRange':
        if (customDateFrom && customDateTo) {
          return { dateFrom: customDateFrom, dateTo: customDateTo }
        }
        return { date: today }
      default:
        return { date: today }
    }
  }, [filterType, customDateFrom, customDateTo])

  const { data: entries, isLoading } = useQuery({
    queryKey: ['calories', getDateParams],
    queryFn: () => calorieService.getEntries(getDateParams),
  })

  // Calculate total calories for filtered results
  const totalFilteredCalories = useMemo(() => {
    return entries?.reduce((sum, entry) => sum + entry.calories, 0) || 0
  }, [entries])

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
        
        {/* Filter Section */}
        <div className="filter-section" style={{ marginBottom: '1rem', padding: '1rem', backgroundColor: '#f8f9fa', borderRadius: '8px' }}>
          <div style={{ marginBottom: '1rem' }}>
            <label htmlFor="filterType" style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 'bold' }}>
              {t('dashboard.filterBy')}:
            </label>
            <select 
              id="filterType"
              value={filterType} 
              onChange={(e) => setFilterType(e.target.value as FilterType)}
              style={{ padding: '0.5rem', borderRadius: '4px', border: '1px solid #ddd', marginRight: '1rem' }}
            >
              <option value="today">{t('dashboard.today')}</option>
              <option value="lastWeek">{t('dashboard.lastWeek')}</option>
              <option value="lastMonth">{t('dashboard.lastMonth')}</option>
              <option value="customRange">{t('dashboard.customRange')}</option>
            </select>
          </div>
          
          {filterType === 'customRange' && (
            <div className="custom-date-range" style={{ display: 'flex', gap: '1rem', alignItems: 'center', flexWrap: 'wrap' }}>
              <div>
                <label htmlFor="dateFrom" style={{ display: 'block', marginBottom: '0.25rem' }}>
                  {t('dashboard.dateFrom')}:
                </label>
                <input
                  type="date"
                  id="dateFrom"
                  value={customDateFrom}
                  onChange={(e) => setCustomDateFrom(e.target.value)}
                  style={{ padding: '0.5rem', borderRadius: '4px', border: '1px solid #ddd' }}
                />
              </div>
              <div>
                <label htmlFor="dateTo" style={{ display: 'block', marginBottom: '0.25rem' }}>
                  {t('dashboard.dateTo')}:
                </label>
                <input
                  type="date"
                  id="dateTo"
                  value={customDateTo}
                  onChange={(e) => setCustomDateTo(e.target.value)}
                  style={{ padding: '0.5rem', borderRadius: '4px', border: '1px solid #ddd' }}
                />
              </div>
            </div>
          )}
          
          {/* Total calories display */}
          <div style={{ marginTop: '1rem', padding: '0.75rem', backgroundColor: '#e8f5e8', borderRadius: '4px', fontWeight: 'bold', color: '#2d5016' }}>
            {t('dashboard.totalCaloriesFiltered')} {totalFilteredCalories} {t('dashboard.kcal')}
          </div>
        </div>
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
          </div>
        )}
      </section>
    </div>
  )
}