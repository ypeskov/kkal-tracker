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

class CalorieService {
  private getAuthHeaders = () => {
    const token = sessionStorage.getItem('token')
    return {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    }
  }

  getEntries = async (): Promise<CalorieEntry[]> => {
    const response = await fetch('/api/calories', {
      headers: this.getAuthHeaders(),
    })

    if (!response.ok) {
      throw new Error('Failed to fetch entries')
    }

    return response.json()
  }

  addEntry = async (entry: CalorieEntry): Promise<CalorieEntry> => {
    const response = await fetch('/api/calories', {
      method: 'POST',
      headers: this.getAuthHeaders(),
      body: JSON.stringify(entry),
    })

    if (!response.ok) {
      throw new Error('Failed to add entry')
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