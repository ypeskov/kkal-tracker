import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState, useMemo, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { calorieService } from '../api/calories'
import { ingredientService, Ingredient } from '../api/ingredients'
import FoodAutocomplete from './FoodAutocomplete'
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
  const { t, i18n } = useTranslation()
  const [foodName, setFoodName] = useState('')
  const [weight, setWeight] = useState('')
  const [kcalPer100g, setKcalPer100g] = useState('')
  const [fats, setFats] = useState('')
  const [carbs, setCarbs] = useState('')
  const [proteins, setProteins] = useState('')
  const [filterType, setFilterType] = useState<FilterType>('today')
  const [customDateFrom, setCustomDateFrom] = useState('')
  const [customDateTo, setCustomDateTo] = useState('')
  const [editingEntry, setEditingEntry] = useState<any>(null)
  const [editFoodName, setEditFoodName] = useState('')
  const [editWeight, setEditWeight] = useState('')
  const [editKcalPer100g, setEditKcalPer100g] = useState('')
  const [editFats, setEditFats] = useState('')
  const [editCarbs, setEditCarbs] = useState('')
  const [editProteins, setEditProteins] = useState('')
  const [editDate, setEditDate] = useState('')
  const [editTime, setEditTime] = useState('')
  const queryClient = useQueryClient()

  // Load ingredients on component mount
  useEffect(() => {
    const loadIngredients = async () => {
      if (!ingredientService.hasCachedIngredients()) {
        try {
          await ingredientService.loadAndCacheIngredients()
        } catch (error) {
          console.error('Failed to load ingredients:', error)
        }
      }
    }
    
    loadIngredients()
  }, [])

  // Auto-calculate total calories
  const totalCalories = useMemo(() => {
    const weightNum = parseFloat(weight) || 0
    const kcalNum = parseFloat(kcalPer100g) || 0
    return Math.round((weightNum * kcalNum) / 100)
  }, [weight, kcalPer100g])

  // Auto-calculate total calories for editing
  const editTotalCalories = useMemo(() => {
    const weightNum = parseFloat(editWeight) || 0
    const kcalNum = parseFloat(editKcalPer100g) || 0
    return Math.round((weightNum * kcalNum) / 100)
  }, [editWeight, editKcalPer100g])

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

  // Group entries by date
  const entriesByDate = useMemo(() => {
    if (!entries || entries.length === 0) {
      return {}
    }
    
    const grouped: { [date: string]: any[] } = {}
    
    entries.forEach(entry => {
      const date = new Date(entry.meal_datetime).toISOString().split('T')[0]
      if (!grouped[date]) {
        grouped[date] = []
      }
      grouped[date].push(entry)
    })
    
    // Sort entries within each date by time
    Object.keys(grouped).forEach(date => {
      grouped[date].sort((a, b) => new Date(a.meal_datetime).getTime() - new Date(b.meal_datetime).getTime())
    })
    
    return grouped
  }, [entries])

  // Calculate totals for filtered results
  const nutritionTotals = useMemo(() => {
    if (!entries || entries.length === 0) {
      return { calories: 0, fats: 0, carbs: 0, proteins: 0 }
    }
    
    return entries.reduce((totals, entry) => ({
      calories: totals.calories + entry.calories,
      fats: totals.fats + (entry.fats || 0),
      carbs: totals.carbs + (entry.carbs || 0),
      proteins: totals.proteins + (entry.proteins || 0)
    }), { calories: 0, fats: 0, carbs: 0, proteins: 0 })
  }, [entries])

  // Format date for display
  const formatDateHeader = (dateString: string) => {
    const date = new Date(dateString)
    const today = new Date()
    const yesterday = new Date(today)
    yesterday.setDate(yesterday.getDate() - 1)
    
    const dateOnly = date.toISOString().split('T')[0]
    const todayOnly = today.toISOString().split('T')[0]
    const yesterdayOnly = yesterday.toISOString().split('T')[0]
    
    // Get current language and map to locale
    const currentLang = i18n.language
    const locale = currentLang === 'uk-UA' ? 'uk-UA' : currentLang === 'ru-UA' ? 'ru-RU' : 'en-US'
    
    if (dateOnly === todayOnly) {
      return t('dashboard.today') + ' (' + date.toLocaleDateString(locale) + ')'
    } else if (dateOnly === yesterdayOnly) {
      return t('dashboard.yesterday') + ' (' + date.toLocaleDateString(locale) + ')'
    } else {
      return date.toLocaleDateString(locale)
    }
  }

  // Format time for display
  const formatTime = (datetimeString: string) => {
    const date = new Date(datetimeString)
    
    // Get current language and map to locale
    const currentLang = i18n.language
    const locale = currentLang === 'uk-UA' ? 'uk-UA' : currentLang === 'ru-UA' ? 'ru-RU' : 'en-US'
    
    return date.toLocaleTimeString(locale, { hour: '2-digit', minute: '2-digit' })
  }

  // Calculate daily totals for a given date's entries
  const calculateDailyTotals = (dateEntries: any[]) => {
    return dateEntries.reduce((totals, entry) => ({
      calories: totals.calories + (entry.calories || 0),
      fats: totals.fats + (entry.fats || 0),
      carbs: totals.carbs + (entry.carbs || 0),
      proteins: totals.proteins + (entry.proteins || 0)
    }), { calories: 0, fats: 0, carbs: 0, proteins: 0 })
  }

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

  const updateEntryMutation = useMutation({
    mutationFn: ({ id, entry }: { id: number, entry: any }) => calorieService.updateEntry(id, entry),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['calories'] })
      setEditingEntry(null)
      setEditFoodName('')
      setEditWeight('')
      setEditKcalPer100g('')
      setEditFats('')
      setEditCarbs('')
      setEditProteins('')
      setEditDate('')
      setEditTime('')
    },
  })

  const [showDeleteConfirmation, setShowDeleteConfirmation] = useState(false)

  const deleteEntryMutation = useMutation({
    mutationFn: (id: number) => calorieService.deleteEntry(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['calories'] })
      setEditingEntry(null)
      setEditFoodName('')
      setEditWeight('')
      setEditKcalPer100g('')
      setEditFats('')
      setEditCarbs('')
      setEditProteins('')
      setEditDate('')
      setEditTime('')
      setShowDeleteConfirmation(false)
    },
  })

  const handleShowDeleteConfirmation = () => {
    setShowDeleteConfirmation(true)
  }

  const handleConfirmDelete = () => {
    if (editingEntry) {
      deleteEntryMutation.mutate(editingEntry.id)
    }
  }

  const handleCancelDelete = () => {
    setShowDeleteConfirmation(false)
  }

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

  const handleEdit = (entry: any) => {
    setEditingEntry(entry)
    setEditFoodName(entry.food)
    setEditWeight(entry.weight.toString())
    setEditKcalPer100g(entry.kcalPer100g.toString())
    setEditFats(entry.fats ? entry.fats.toString() : '')
    setEditCarbs(entry.carbs ? entry.carbs.toString() : '')
    setEditProteins(entry.proteins ? entry.proteins.toString() : '')
    
    // Parse the meal_datetime to separate date and time
    const mealDate = new Date(entry.meal_datetime)
    const dateString = mealDate.toISOString().split('T')[0] // YYYY-MM-DD
    const timeString = mealDate.toTimeString().split(' ')[0].substring(0, 5) // HH:MM
    setEditDate(dateString)
    setEditTime(timeString)
  }

  const handleIngredientSelect = (ingredient: Ingredient) => {
    setKcalPer100g(ingredient.kcalPer100g.toString())
    setFats(ingredient.fats ? ingredient.fats.toString() : '')
    setCarbs(ingredient.carbs ? ingredient.carbs.toString() : '')
    setProteins(ingredient.proteins ? ingredient.proteins.toString() : '')
  }

  const handleLogout = () => {
    ingredientService.clearCache()
    onLogout()
  }

  const handleUpdateSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!editFoodName || !editWeight || !editKcalPer100g || !editingEntry || !editDate || !editTime) return
    
    // Combine date and time to create the meal_datetime
    const mealDateTime = new Date(`${editDate}T${editTime}:00`).toISOString()
    
    const entryData: any = {
      food: editFoodName,
      weight: parseFloat(editWeight),
      kcalPer100g: parseFloat(editKcalPer100g),
      calories: editTotalCalories,
      meal_datetime: mealDateTime,
    }
    
    // Only include nutritional fields if they have values
    if (editFats && parseFloat(editFats) > 0) entryData.fats = parseFloat(editFats)
    if (editCarbs && parseFloat(editCarbs) > 0) entryData.carbs = parseFloat(editCarbs)
    if (editProteins && parseFloat(editProteins) > 0) entryData.proteins = parseFloat(editProteins)
    
    updateEntryMutation.mutate({ id: editingEntry.id, entry: entryData })
  }

  const handleCancelEdit = () => {
    setEditingEntry(null)
    setEditFoodName('')
    setEditWeight('')
    setEditKcalPer100g('')
    setEditFats('')
    setEditCarbs('')
    setEditProteins('')
    setEditDate('')
    setEditTime('')
  }

  return (
    <div className="dashboard-container">
      <header className="dashboard-header">
        <div>
          <h1>{t('dashboard.title')}</h1>
          <p>{t('auth.welcome')}, {user?.email}!</p>
        </div>
        <div className="header-controls">
          <LanguageSwitcher />
          <button 
            onClick={handleLogout}
            className="btn"
            style={{ backgroundColor: '#dc3545' }}
          >
            {t('auth.logout')}
          </button>
        </div>
      </header>

      <section style={{ marginBottom: '2rem' }}>
        <h2>{t('dashboard.addFoodEntry')}</h2>
        <form onSubmit={handleSubmit} className="food-form">
          <div className="form-row">
            <div className="form-group">
              <label htmlFor="foodName">{t('dashboard.foodName')}:</label>
              <FoodAutocomplete
                id="foodName"
                value={foodName}
                onChange={setFoodName}
                onSelect={handleIngredientSelect}
                placeholder={t('dashboard.foodName')}
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
          
          {/* Nutrition totals display */}
          <div style={{ marginTop: '1rem', padding: '0.75rem', backgroundColor: '#e8f5e8', borderRadius: '4px', fontWeight: 'bold', color: '#2d5016' }}>
            <div style={{ marginBottom: '0.5rem' }}>
              {t('dashboard.totalCaloriesFiltered')} {nutritionTotals.calories} {t('dashboard.kcal')}
            </div>
            {(nutritionTotals.fats > 0 || nutritionTotals.carbs > 0 || nutritionTotals.proteins > 0) && (
              <div style={{ display: 'flex', gap: '1rem', fontSize: '0.9em', flexWrap: 'wrap' }}>
                {nutritionTotals.fats > 0 && (
                  <span>{t('dashboard.fats')}: {nutritionTotals.fats.toFixed(1)}g</span>
                )}
                {nutritionTotals.carbs > 0 && (
                  <span>{t('dashboard.carbs')}: {nutritionTotals.carbs.toFixed(1)}g</span>
                )}
                {nutritionTotals.proteins > 0 && (
                  <span>{t('dashboard.proteins')}: {nutritionTotals.proteins.toFixed(1)}g</span>
                )}
              </div>
            )}
          </div>
        </div>
        {isLoading ? (
          <p>{t('common.loading')}</p>
        ) : (
          <div>
            {entries?.length === 0 ? (
              <p>{t('dashboard.noEntries')}</p>
            ) : (
              <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
                {Object.keys(entriesByDate)
                  .sort((a, b) => new Date(b).getTime() - new Date(a).getTime()) // Sort dates desc (newest first)
                  .map((dateKey) => (
                    <div key={dateKey}>
                      {/* Date header - only show if period > 1 day or when filtering by range */}
                      {(filterType !== 'today' || Object.keys(entriesByDate).length > 1) && (
                        <h3 style={{ 
                          margin: '0 0 1rem 0', 
                          padding: '0.5rem 0', 
                          fontSize: '1.1rem',
                          color: '#495057',
                          borderBottom: '2px solid #e9ecef'
                        }}>
                          {formatDateHeader(dateKey)}
                        </h3>
                      )}
                      
                      {/* Entries for this date */}
                      <ul style={{ listStyle: 'none', padding: 0, margin: 0 }}>
                        {entriesByDate[dateKey].map((entry: any) => (
                          <li key={entry.id} className="entry-item" onClick={() => handleEdit(entry)} style={{ cursor: 'pointer' }}>
                            <div className="entry-header">
                              <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                                <span style={{ 
                                  fontSize: '0.85rem', 
                                  color: '#6c757d',
                                  minWidth: '55px',
                                  fontWeight: 'normal'
                                }}>
                                  {formatTime(entry.meal_datetime)}
                                </span>
                                <strong>{entry.food}</strong>
                              </div>
                              <strong>{entry.calories} {t('dashboard.kcal')}</strong>
                            </div>
                            <div className="entry-details" style={{ marginLeft: '60px' }}>
                              <span>{t('dashboard.weight')}: {entry.weight}g</span>
                              <span>{t('dashboard.kcalPer100g')}: {entry.kcalPer100g}</span>
                              {entry.fats && <span>{t('dashboard.fats')}: {entry.fats}g</span>}
                              {entry.carbs && <span>{t('dashboard.carbs')}: {entry.carbs}g</span>}
                              {entry.proteins && <span>{t('dashboard.proteins')}: {entry.proteins}g</span>}
                            </div>
                          </li>
                        ))}
                      </ul>

                      {/* Daily totals - only show when there are multiple days */}
                      {Object.keys(entriesByDate).length > 1 && (() => {
                        const dailyTotals = calculateDailyTotals(entriesByDate[dateKey])
                        return (
                          <div style={{ 
                            marginTop: '1rem', 
                            padding: '0.75rem', 
                            backgroundColor: '#f8f9fa', 
                            borderRadius: '5px',
                            border: '1px solid #dee2e6'
                          }}>
                            <div style={{ 
                              display: 'flex', 
                              alignItems: 'center', 
                              gap: '1rem',
                              fontWeight: 'bold',
                              fontSize: '0.9rem'
                            }}>
                              <span style={{ color: '#495057' }}>{t('dashboard.total')}:</span>
                              <span>{dailyTotals.calories} {t('dashboard.kcal')}</span>
                              {dailyTotals.fats > 0 && (
                                <span>{t('dashboard.fats')}: {dailyTotals.fats.toFixed(1)}g</span>
                              )}
                              {dailyTotals.carbs > 0 && (
                                <span>{t('dashboard.carbs')}: {dailyTotals.carbs.toFixed(1)}g</span>
                              )}
                              {dailyTotals.proteins > 0 && (
                                <span>{t('dashboard.proteins')}: {dailyTotals.proteins.toFixed(1)}g</span>
                              )}
                            </div>
                          </div>
                        )
                      })()}
                    </div>
                  ))}
              </div>
            )}
          </div>
        )}
      </section>

      {/* Edit Modal */}
      {editingEntry && (
        <div style={{ 
          position: 'fixed', 
          top: 0, 
          left: 0, 
          right: 0, 
          bottom: 0, 
          backgroundColor: 'rgba(0,0,0,0.5)', 
          zIndex: 1000,
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
          padding: '1rem'
        }}>
          <div style={{ 
            backgroundColor: 'white', 
            borderRadius: '8px', 
            padding: '2rem', 
            maxWidth: '500px', 
            width: '100%',
            maxHeight: '90vh',
            overflowY: 'auto'
          }}>
            <h2 style={{ marginTop: 0 }}>{t('dashboard.editEntry')}</h2>
            
            <form onSubmit={handleUpdateSubmit}>
              <div className="form-row">
                <div className="form-group">
                  <label htmlFor="editFoodName">{t('dashboard.foodName')}:</label>
                  <input
                    type="text"
                    id="editFoodName"
                    value={editFoodName}
                    onChange={(e) => setEditFoodName(e.target.value)}
                    required
                  />
                </div>
              </div>

              <div className="form-row">
                <div className="form-group">
                  <label htmlFor="editDate">{t('dashboard.date')}:</label>
                  <input
                    type="date"
                    id="editDate"
                    value={editDate}
                    onChange={(e) => setEditDate(e.target.value)}
                    required
                  />
                </div>

                <div className="form-group">
                  <label htmlFor="editTime">{t('dashboard.time')}:</label>
                  <input
                    type="time"
                    id="editTime"
                    value={editTime}
                    onChange={(e) => setEditTime(e.target.value)}
                    required
                  />
                </div>
              </div>
              
              <div className="form-row">
                <div className="form-group">
                  <label htmlFor="editWeight">{t('dashboard.weight')}:</label>
                  <input
                    type="number"
                    id="editWeight"
                    value={editWeight}
                    onChange={(e) => setEditWeight(e.target.value)}
                    required
                    min="0"
                    step="0.1"
                  />
                </div>

                <div className="form-group">
                  <label htmlFor="editKcalPer100g">{t('dashboard.kcalPer100g')}:</label>
                  <input
                    type="number"
                    id="editKcalPer100g"
                    value={editKcalPer100g}
                    onChange={(e) => setEditKcalPer100g(e.target.value)}
                    required
                    min="0"
                    step="0.1"
                  />
                </div>
              </div>

              <div className="form-row">
                <div className="form-group">
                  <label htmlFor="editFats">{t('dashboard.fats')}:</label>
                  <input
                    type="number"
                    id="editFats"
                    value={editFats}
                    onChange={(e) => setEditFats(e.target.value)}
                    min="0"
                    step="0.1"
                  />
                </div>

                <div className="form-group">
                  <label htmlFor="editCarbs">{t('dashboard.carbs')}:</label>
                  <input
                    type="number"
                    id="editCarbs"
                    value={editCarbs}
                    onChange={(e) => setEditCarbs(e.target.value)}
                    min="0"
                    step="0.1"
                  />
                </div>

                <div className="form-group">
                  <label htmlFor="editProteins">{t('dashboard.proteins')}:</label>
                  <input
                    type="number"
                    id="editProteins"
                    value={editProteins}
                    onChange={(e) => setEditProteins(e.target.value)}
                    min="0"
                    step="0.1"
                  />
                </div>
              </div>

              <div className="form-row">
                <div className="form-group">
                  <label htmlFor="editTotalCalories">{t('dashboard.totalCalories')}:</label>
                  <input
                    type="number"
                    id="editTotalCalories"
                    value={editTotalCalories}
                    readOnly
                    className="readonly-field"
                  />
                </div>
              </div>

              <div className="form-row" style={{ marginTop: '2rem', justifyContent: 'space-between', gap: '1rem' }}>
                <button 
                  type="button" 
                  onClick={handleShowDeleteConfirmation}
                  className="btn"
                  style={{ backgroundColor: '#dc3545' }}
                  disabled={updateEntryMutation.isPending || deleteEntryMutation.isPending}
                >
                  {t('dashboard.delete')}
                </button>
                
                <div style={{ display: 'flex', gap: '1rem' }}>
                  <button 
                    type="button" 
                    onClick={handleCancelEdit}
                    className="btn"
                    style={{ backgroundColor: '#6c757d' }}
                    disabled={updateEntryMutation.isPending || deleteEntryMutation.isPending}
                  >
                    {t('dashboard.cancel')}
                  </button>
                  <button 
                    type="submit" 
                    className="btn"
                    disabled={updateEntryMutation.isPending || deleteEntryMutation.isPending || !editFoodName || !editWeight || !editKcalPer100g}
                    style={{ backgroundColor: '#28a745' }}
                  >
                    {updateEntryMutation.isPending ? t('common.loading') + '...' : t('dashboard.updateEntry')}
                  </button>
                </div>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Delete Confirmation Dialog */}
      {showDeleteConfirmation && (
        <div style={{ 
          position: 'fixed', 
          top: 0, 
          left: 0, 
          right: 0, 
          bottom: 0, 
          backgroundColor: 'rgba(0,0,0,0.5)', 
          zIndex: 1001,
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
          padding: '1rem'
        }}>
          <div style={{ 
            backgroundColor: 'white', 
            borderRadius: '8px', 
            padding: '2rem', 
            maxWidth: '400px', 
            width: '100%',
            textAlign: 'center'
          }}>
            <h3 style={{ marginTop: 0, color: '#dc3545' }}>{t('dashboard.confirmDelete')}</h3>
            <p style={{ marginBottom: '2rem' }}>
              {t('dashboard.deleteWarning', { foodName: editingEntry?.food })}
            </p>
            
            <div style={{ display: 'flex', justifyContent: 'center', gap: '1rem' }}>
              <button 
                type="button" 
                onClick={handleCancelDelete}
                className="btn"
                style={{ backgroundColor: '#6c757d' }}
                disabled={deleteEntryMutation.isPending}
              >
                {t('dashboard.cancel')}
              </button>
              <button 
                type="button" 
                onClick={handleConfirmDelete}
                className="btn"
                style={{ backgroundColor: '#dc3545' }}
                disabled={deleteEntryMutation.isPending}
              >
                {deleteEntryMutation.isPending ? t('common.loading') + '...' : t('dashboard.confirmDeleteButton')}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}