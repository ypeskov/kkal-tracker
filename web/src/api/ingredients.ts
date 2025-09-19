export interface Ingredient {
  id: number
  name: string
  kcalPer100g: number
  fats?: number
  carbs?: number
  proteins?: number
  user_id?: number
  created_at?: string
  updated_at?: string
}

export interface CreateIngredientData {
  name: string
  kcalPer100g: number
  fats?: number
  carbs?: number
  proteins?: number
}

export interface UpdateIngredientData {
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

  // Load ingredients if cache is empty
  private ensureIngredientsLoaded = async (): Promise<void> => {
    if (!this.hasCachedIngredients()) {
      await this.loadAndCacheIngredients()
    }
  }

  // Search ingredients from cache (load if needed)
  searchIngredients = async (query: string, limit: number = 10): Promise<Ingredient[]> => {
    try {
      if (!query || query.length < 2) {
        return []
      }

      // Ensure ingredients are loaded
      await this.ensureIngredientsLoaded()

      // Search from cache
      const cachedResults = this.searchCachedIngredients(query, limit)
      return Array.isArray(cachedResults) ? cachedResults : []

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

  // Get ingredient by ID
  getIngredientById = async (id: number): Promise<Ingredient | null> => {
    try {
      const response = await fetch(`/api/ingredients/${id}`, {
        headers: this.getAuthHeaders(),
      })

      if (!response.ok) {
        if (response.status === 404) {
          return null
        }
        throw new Error('Failed to fetch ingredient')
      }

      return await response.json()
    } catch (error) {
      console.error('Failed to get ingredient:', error)
      throw error
    }
  }

  // Create a new ingredient
  createIngredient = async (data: CreateIngredientData): Promise<Ingredient> => {
    try {
      const response = await fetch('/api/ingredients', {
        method: 'POST',
        headers: this.getAuthHeaders(),
        body: JSON.stringify(data),
      })

      if (!response.ok) {
        const error = await response.text()
        throw new Error(error || 'Failed to create ingredient')
      }

      const newIngredient = await response.json()
      
      // Update cache
      await this.loadAndCacheIngredients()
      
      return newIngredient
    } catch (error) {
      console.error('Failed to create ingredient:', error)
      throw error
    }
  }

  // Update an existing ingredient
  updateIngredient = async (id: number, data: UpdateIngredientData): Promise<Ingredient> => {
    try {
      const response = await fetch(`/api/ingredients/${id}`, {
        method: 'PUT',
        headers: this.getAuthHeaders(),
        body: JSON.stringify(data),
      })

      if (!response.ok) {
        const error = await response.text()
        throw new Error(error || 'Failed to update ingredient')
      }

      const updatedIngredient = await response.json()
      
      // Update cache
      await this.loadAndCacheIngredients()
      
      return updatedIngredient
    } catch (error) {
      console.error('Failed to update ingredient:', error)
      throw error
    }
  }

  // Delete an ingredient
  deleteIngredient = async (id: number): Promise<void> => {
    try {
      const response = await fetch(`/api/ingredients/${id}`, {
        method: 'DELETE',
        headers: this.getAuthHeaders(),
      })

      if (!response.ok) {
        if (response.status === 404) {
          throw new Error('Ingredient not found')
        }
        throw new Error('Failed to delete ingredient')
      }

      // Update cache
      await this.loadAndCacheIngredients()
    } catch (error) {
      console.error('Failed to delete ingredient:', error)
      throw error
    }
  }
}

export const ingredientService = new IngredientService()