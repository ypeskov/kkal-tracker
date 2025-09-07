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
    if (!ingredients || !Array.isArray(ingredients) || ingredients.length === 0) return []
    
    const queryLower = query.toLowerCase()
    
    return ingredients
      .filter(ingredient => ingredient && ingredient.name && ingredient.name.toLowerCase().includes(queryLower))
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

      const result = await response.json()
      
      // Ensure we return a valid array
      if (!result || !Array.isArray(result)) {
        console.warn('API returned invalid data:', result)
        return []
      }
      
      return result
    } catch (error) {
      console.error('Failed to search ingredients:', error)
      return []
    }
  }

  // Smart search: use cache first, fallback to API
  searchIngredients = async (query: string, limit: number = 10): Promise<Ingredient[]> => {
    try {
      if (!query || query.length < 2) {
        return []
      }
      
      // First try cached search
      const cachedResults = this.searchCachedIngredients(query, limit)
      
      if (cachedResults && Array.isArray(cachedResults) && cachedResults.length > 0) {
        return cachedResults
      }
      
      // If cache is empty or no results, try API
      const apiResults = await this.searchIngredientsAPI(query, limit)
      return Array.isArray(apiResults) ? apiResults : []
      
    } catch (error) {
      console.error('Failed to search ingredients:', error)
      return []
    }
  }

  // Clear cached ingredients (e.g., on logout)
  clearCache = (): void => {
    sessionStorage.removeItem(this.STORAGE_KEY)
  }

  // Check if ingredients are cached
  hasCachedIngredients = (): boolean => {
    try {
      const cached = sessionStorage.getItem(this.STORAGE_KEY)
      if (!cached) return false
      const parsed = JSON.parse(cached)
      return Array.isArray(parsed) && parsed.length > 0
    } catch (error) {
      return false
    }
  }
}

export const ingredientService = new IngredientService()