import { useState, useEffect, useRef } from 'react'
import { Ingredient, ingredientService } from '../api/ingredients'
import './FoodAutocomplete.css'

interface FoodAutocompleteProps {
  value: string
  onChange: (value: string) => void
  onSelect: (ingredient: Ingredient) => void
  placeholder?: string
  required?: boolean
  id?: string
}

export default function FoodAutocomplete({
  value,
  onChange,
  onSelect,
  placeholder,
  required,
  id
}: FoodAutocompleteProps) {
  const [suggestions, setSuggestions] = useState<Ingredient[]>([])
  const [showSuggestions, setShowSuggestions] = useState(false)
  const [selectedIndex, setSelectedIndex] = useState(-1)
  const inputRef = useRef<HTMLInputElement>(null)
  const listRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
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
    onChange(ingredient.name)
    onSelect(ingredient)
    setShowSuggestions(false)
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

  const handleFocus = () => {
    if (value.length >= 2 && suggestions.length > 0) {
      setShowSuggestions(true)
    }
  }

  return (
    <div className="food-autocomplete">
      <input
        ref={inputRef}
        type="text"
        id={id}
        value={value}
        onChange={handleInputChange}
        onKeyDown={handleKeyDown}
        onBlur={handleBlur}
        onFocus={handleFocus}
        placeholder={placeholder}
        required={required}
        autoComplete="off"
        className="form-input"
      />
      
      {showSuggestions && suggestions.length > 0 && (
        <div ref={listRef} className="autocomplete-suggestions">
          {suggestions.map((ingredient, index) => (
            <div
              key={ingredient.id}
              className={`suggestion-item ${index === selectedIndex ? 'selected' : ''}`}
              onClick={() => handleSuggestionClick(ingredient)}
              onMouseEnter={() => setSelectedIndex(index)}
            >
              <div className="suggestion-name">{ingredient.name}</div>
              <div className="suggestion-details">
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
}