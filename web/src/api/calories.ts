interface CalorieEntry {
  id?: number
  food: string
  weight: number
  kcalPer100g: number
  fats?: number
  carbs?: number
  proteins?: number
  calories: number
  meal_datetime: string
}

interface CreateEntryResult {
  entry: CalorieEntry
  new_ingredient_created: boolean
}

class CalorieService {
  private getAuthHeaders = () => {
    const token = sessionStorage.getItem('token')
    return {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    }
  }

  getEntries = async (params?: { date?: string; dateFrom?: string; dateTo?: string }): Promise<CalorieEntry[]> => {
    const searchParams = new URLSearchParams()
    
    if (params?.date) {
      searchParams.append('date', params.date)
    }
    
    if (params?.dateFrom && params?.dateTo) {
      searchParams.append('dateFrom', params.dateFrom)
      searchParams.append('dateTo', params.dateTo)
    }

    const url = `/api/calories${searchParams.toString() ? `?${searchParams.toString()}` : ''}`
    
    const response = await fetch(url, {
      headers: this.getAuthHeaders(),
    })

    if (!response.ok) {
      throw new Error('Failed to fetch entries')
    }

    return response.json()
  }

  addEntry = async (entry: CalorieEntry): Promise<CreateEntryResult> => {
    const response = await fetch('/api/calories', {
      method: 'POST',
      headers: this.getAuthHeaders(),
      body: JSON.stringify(entry),
    })

    if (!response.ok) {
      throw new Error('Failed to add entry')
    }

    const result: CreateEntryResult = await response.json()
    
    // If a new ingredient was created, invalidate the ingredients cache
    if (result.new_ingredient_created) {
      const { ingredientService } = await import('./ingredients')
      ingredientService.clearCache()
      // Reload ingredients to refresh the cache
      await ingredientService.loadAndCacheIngredients()
    }

    return result
  }

  updateEntry = async (id: number, entry: CalorieEntry): Promise<CalorieEntry> => {
    const response = await fetch(`/api/calories/${id}`, {
      method: 'PUT',
      headers: this.getAuthHeaders(),
      body: JSON.stringify(entry),
    })

    if (!response.ok) {
      throw new Error('Failed to update entry')
    }

    return response.json()
  }

  deleteEntry = async (id: number): Promise<void> => {
    const response = await fetch(`/api/calories/${id}`, {
      method: 'DELETE',
      headers: this.getAuthHeaders(),
    })

    if (!response.ok) {
      throw new Error('Failed to delete entry')
    }
  }
}

export const calorieService = new CalorieService()