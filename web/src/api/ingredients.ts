export interface Ingredient {
  id: number
  name: string
  kcalPer100g: number
  fats?: number
  carbs?: number
  proteins?: number
}

class IngredientService {
  private getAuthHeaders = () => {
    const token = sessionStorage.getItem('token')
    return {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    }
  }

  private STORAGE_KEY = 'user_ingredients'

  // Load all user ingredients and cache in sessionStorage
  loadAndCacheIngredients = async (): Promise<Ingredient[]> => {
    try {
      const response = await fetch('/api/ingredients', {
        headers: this.getAuthHeaders(),
      })

      if (!response.ok) {
        throw new Error('Failed to fetch ingredients')
      }

      const ingredients: Ingredient[] = await response.json()
      
      // Cache in sessionStorage for fast access
      sessionStorage.setItem(this.STORAGE_KEY, JSON.stringify(ingredients))
      
      return ingredients
    } catch (error) {
      console.error('Failed to load ingredients:', error)
      return this.getCachedIngredients()
    }
  }

  // Get cached ingredients from sessionStorage
  getCachedIngredients = (): Ingredient[] => {
    try {
      const cached = sessionStorage.getItem(this.STORAGE_KEY)
      return cached ? JSON.parse(cached) : []
    } catch (error) {
      console.error('Failed to parse cached ingredients:', error)
      return []
    }
  }

  // Search ingredients locally from cache
  searchCachedIngredients = (query: string, limit: number = 10): Ingredient[] => {
    if (query.length < 2) return []
    
    const ingredients = this.getCachedIngredients()
    const queryLower = query.toLowerCase()
    
    return ingredients
      .filter(ingredient => ingredient.name.toLowerCase().startsWith(queryLower))
      .slice(0, limit)
      .sort((a, b) => a.name.localeCompare(b.name))
  }

  // Fallback: search ingredients via API if cache is empty
  searchIngredientsAPI = async (query: string, limit: number = 10): Promise<Ingredient[]> => {
    if (query.length < 2) return []
    
    try {
      const response = await fetch(`/api/ingredients/search?q=${encodeURIComponent(query)}&limit=${limit}`, {
        headers: this.getAuthHeaders(),
      })

      if (!response.ok) {
        throw new Error('Failed to search ingredients')
      }

      return response.json()
    } catch (error) {
      console.error('Failed to search ingredients:', error)
      return []
    }
  }

  // Smart search: use cache first, fallback to API
  searchIngredients = async (query: string, limit: number = 10): Promise<Ingredient[]> => {
    // First try cached search
    const cachedResults = this.searchCachedIngredients(query, limit)
    
    if (cachedResults.length > 0) {
      return cachedResults
    }
    
    // If cache is empty or no results, try API
    return this.searchIngredientsAPI(query, limit)
  }

  // Clear cached ingredients (e.g., on logout)
  clearCache = (): void => {
    sessionStorage.removeItem(this.STORAGE_KEY)
  }

  // Check if ingredients are cached
  hasCachedIngredients = (): boolean => {
    const cached = sessionStorage.getItem(this.STORAGE_KEY)
    return !!(cached && JSON.parse(cached).length > 0)
  }
}

export const ingredientService = new IngredientService()