interface CalorieEntry {
  id?: number
  food: string
  calories: number
  date: string
}

class CalorieService {
  private getAuthHeaders() {
    const token = localStorage.getItem('token')
    return {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    }
  }

  async getEntries(): Promise<CalorieEntry[]> {
    const response = await fetch('/api/calories', {
      headers: this.getAuthHeaders(),
    })

    if (!response.ok) {
      throw new Error('Failed to fetch entries')
    }

    return response.json()
  }

  async addEntry(entry: CalorieEntry): Promise<CalorieEntry> {
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

  async deleteEntry(id: number): Promise<void> {
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