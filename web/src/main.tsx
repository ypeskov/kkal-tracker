import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import App from './App';

// Import Tailwind CSS with comprehensive design system
import './styles/tailwind.css';   // Tailwind base + complete design system theme

// Import essential global styles
import './styles/global.css';     // Global resets, typography, accessibility

// Complete CSS migration accomplished:
// - variables.css: consolidated into tailwind.css @theme
// - utilities.css: removed (264 lines of Tailwind duplicates)
// - components.css: removed (unused form/button classes)
// - layout.css: removed (unused layout classes)
// - App.css: removed (unused app classes)
// - Component-specific files: removed (fully migrated to Tailwind)
// Result: Clean, optimized CSS architecture with 100% Tailwind adoption

// Import i18n
import './i18n';

const queryClient = new QueryClient();

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <App />
    </QueryClientProvider>
  </StrictMode>,
);