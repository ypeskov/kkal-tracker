import { useState, useEffect, useRef, forwardRef } from 'react'
import { Ingredient, ingredientService } from '../api/ingredients'
// FoodAutocomplete.css imports removed - using Tailwind CSS

interface FoodAutocompleteProps {
  value: string
  onChange: (value: string) => void
  onSelect: (ingredient: Ingredient) => void
  placeholder?: string
  required?: boolean
  id?: string
}

const FoodAutocomplete = forwardRef<HTMLInputElement, FoodAutocompleteProps>(({
  value,
  onChange,
  onSelect,
  placeholder,
  required,
  id
}, ref) => {
  const [suggestions, setSuggestions] = useState<Ingredient[]>([])
  const [showSuggestions, setShowSuggestions] = useState(false)
  const [selectedIndex, setSelectedIndex] = useState(-1)
  const listRef = useRef<HTMLDivElement>(null)
  const lastSelectedValueRef = useRef<string>('')

  useEffect(() => {
    // Skip searching if this value was just selected from dropdown
    if (value && value === lastSelectedValueRef.current) {
      lastSelectedValueRef.current = '' // Reset for next time
      return
    }

    const searchIngredients = async () => {
      if (value.length < 2) {
        setSuggestions([])
        setShowSuggestions(false)
        return
      }

      try {
        const results = await ingredientService.searchIngredients(value)
        setSuggestions(results)
        setShowSuggestions(results.length > 0)
        setSelectedIndex(-1)
      } catch (error) {
        console.error('Failed to search ingredients:', error)
        setSuggestions([])
        setShowSuggestions(false)
      }
    }

    const timeoutId = setTimeout(searchIngredients, 300) // Debounce 300ms
    return () => clearTimeout(timeoutId)
  }, [value])

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    onChange(e.target.value)
  }

  const handleSuggestionClick = (ingredient: Ingredient) => {
    lastSelectedValueRef.current = ingredient.name // Remember what we selected
    onChange(ingredient.name)
    onSelect(ingredient)
    setShowSuggestions(false)
    setSuggestions([]) // Clear suggestions to prevent them from reappearing
    setSelectedIndex(-1)
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (!showSuggestions || suggestions.length === 0) return

    switch (e.key) {
      case 'ArrowDown':
        e.preventDefault()
        setSelectedIndex(prev => 
          prev < suggestions.length - 1 ? prev + 1 : prev
        )
        break
      
      case 'ArrowUp':
        e.preventDefault()
        setSelectedIndex(prev => prev > 0 ? prev - 1 : prev)
        break
      
      case 'Enter':
        e.preventDefault()
        if (selectedIndex >= 0 && selectedIndex < suggestions.length) {
          handleSuggestionClick(suggestions[selectedIndex])
        }
        break
      
      case 'Escape':
        setShowSuggestions(false)
        setSelectedIndex(-1)
        break
    }
  }

  const handleBlur = () => {
    // Delay hiding suggestions to allow for click events
    setTimeout(() => {
      setShowSuggestions(false)
      setSelectedIndex(-1)
    }, 150)
  }

  return (
    <div className="relative w-full">
      <input
        ref={ref}
        type="text"
        id={id}
        value={value}
        onChange={handleInputChange}
        onKeyDown={handleKeyDown}
        onBlur={handleBlur}
        placeholder={placeholder}
        required={required}
        autoComplete="off"
        className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent text-sm transition-colors placeholder-gray-400"
      />
      
      {showSuggestions && suggestions.length > 0 && (
        <div ref={listRef} className="absolute top-full left-0 right-0 bg-white border border-gray-300 border-t-0 rounded-b shadow-lg max-h-48 overflow-y-auto z-50">
          {suggestions.map((ingredient, index) => (
            <div
              key={ingredient.id}
              className={`p-3 cursor-pointer border-b border-gray-100 last:border-b-0 transition-colors duration-100 ${
                index === selectedIndex ? 'bg-gray-100' : 'hover:bg-gray-50'
              }`}
              onClick={() => handleSuggestionClick(ingredient)}
              onMouseEnter={() => setSelectedIndex(index)}
            >
              <div className="font-medium text-gray-800 mb-1">{ingredient.name}</div>
              <div className="text-sm text-gray-600 leading-tight">
                {ingredient.kcalPer100g} kcal/100g
                {ingredient.proteins && ` • ${ingredient.proteins}g protein`}
                {ingredient.carbs && ` • ${ingredient.carbs}g carbs`}
                {ingredient.fats && ` • ${ingredient.fats}g fats`}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
})

FoodAutocomplete.displayName = 'FoodAutocomplete'

export default FoodAutocomplete