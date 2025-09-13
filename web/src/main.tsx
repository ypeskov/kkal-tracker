import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import App from './App';

// Import CSS files in correct order for optimal loading
// 1. CSS Variables (foundation)
import './styles/variables.css';
// 2. Global styles and resets
import './styles/global.css';
// 3. Layout system
import './styles/layout.css';
// 4. Global components
import './styles/components.css';
// 5. Utility classes
import './styles/utilities.css';
// 6. Specific component styles
import './styles/components/entries-list.css';
import './styles/components/entry-item.css';
import './styles/components/food-form.css';
import './styles/components/modal-forms.css';

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