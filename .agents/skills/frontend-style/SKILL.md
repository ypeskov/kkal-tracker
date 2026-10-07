---
name: frontend-style
description: Kkal Tracker frontend conventions - Tailwind page layouts, cards, headers, form inputs, CalculatorInput, buttons, i18n and units. Use when creating or changing pages and components in web/.
---

# Frontend Style Standards

Strict Tailwind CSS utility-first approach with standardized layout patterns. Reference pages: `DashboardPage.tsx`,
`FoodList.tsx`, `Profile.tsx`, `Report.tsx`.

## Critical rules
1. **Never use undefined CSS classes**: semantic classes like `.page`, `.card`, `.page-header` were removed during the
   Tailwind migration
2. **Only Tailwind utilities** for styling
3. **Follow existing patterns** of the reference pages
4. **Responsive**: always include breakpoints (`md:`, `lg:`) for padding and layout
5. **Consistency over customization**: standard patterns rather than a unique layout per page
6. **Every user-facing string** through `t('key')` (`useTranslation()`), added to all 4 locale files in
   `web/src/i18n/locales/` (`en_US`, `uk_UA`, `ru_UA`, `bg_BG`). Units too: `common.kg`, `common.g`,
   `common.kcalPerDay`, `common.kcalPer100g` — no Latin "kg"/"kcal" in non-Latin locales
7. Functional components with hooks, TanStack Query for server state, TanStack Router for routing, `lucide-react` icons

## Page layout
Standard content pages (Dashboard, Food List, Profile, ...) — constrained width:
```tsx
<div className="max-w-screen-xl mx-auto px-4 py-2 md:px-6 lg:px-8">
```
Data visualization pages (Reports, charts) — full width, same padding:
```tsx
<div className="px-4 py-2 md:px-6 lg:px-8">
```

## Card / section
```tsx
<div className="bg-white rounded-lg shadow-md p-4">
```

## Page header
```tsx
<div className="mb-6">
  <h2 className="text-3xl font-semibold text-gray-800">{t('page.title')}</h2>
</div>
```

## Form inputs
```tsx
<input
  type="text"
  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-100 disabled:text-gray-500"
/>
```
Numeric inputs (weight, calories, fats, carbs, proteins) use the integrated calculator:
```tsx
import CalculatorInput from '@/components/CalculatorInput';

<CalculatorInput
  id="weight"
  value={weight}
  onChange={setWeight}
  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
  required
/>
```
The calculator icon on the right toggles a keypad below the input; the input itself is the display. Supports keyboard
and buttons, `+ - * / ( )`, evaluates on `=`, Enter or blur, `C` clears, `←` deletes. Focus stays on the input. Used
in AddFoodEntryForm, EditEntryModal, AddIngredientModal, EditIngredientModal.

## Buttons
Primary:
```tsx
<button className="px-6 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors">
```
Secondary / outline:
```tsx
<button className="px-4 py-2 border border-gray-300 rounded-md hover:bg-gray-50 transition-colors">
```

## Replacing old CSS classes
| Old class | Replacement |
|-----------|-------------|
| `.page` | `max-w-screen-xl mx-auto px-4 py-2 md:px-6 lg:px-8` (or the full-width variant) |
| `.card` | `bg-white rounded-lg shadow-md` |
| `.page-header` | `mb-6` |
| `.page-title` | `text-3xl font-semibold text-gray-800` |
