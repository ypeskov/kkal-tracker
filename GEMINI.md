
# Project Analysis

This document provides an analysis of the Kkal-tracker project.

## Backend

The backend is a Go application built with the Echo framework. It uses a SQLite database for data storage and `goose` for database migrations. Authentication is handled using JWT.

### Key Components:

*   **Framework:** Echo
*   **Database:** SQLite
*   **Migrations:** `pressly/goose`
*   **Authentication:** `golang-jwt/jwt`
*   **Configuration:** `joho/godotenv`

### API Endpoints:

*   **Auth:**
    *   `POST /api/v1/auth/login`: User login
    *   `POST /api/v1/auth/register`: User registration
    *   `GET /api/v1/auth/me`: Get current user details (requires authentication)
*   **Calories:** (all require authentication)
    *   `GET /api/v1/calories`: Get calorie entries by date or date range
    *   `POST /api/v1/calories`: Create a new calorie entry
    *   `PUT /api/v1/calories/:id`: Update an existing calorie entry
    *   `DELETE /api/v1/calories/:id`: Delete a calorie entry
*   **Ingredients:** (all require authentication)
    *   `GET /api/v1/ingredients`: Get all ingredients for the user
    *   `GET /api/v1/ingredients/search`: Search for ingredients

### Database Schema:

The database consists of the following tables:

*   `users`: Stores user information (id, email, password_hash).
*   `calorie_entries`: Stores individual calorie consumption entries.
*   `global_ingredients`: Stores a master list of ingredients with their nutritional information.
*   `global_ingredient_names`: Stores the names of global ingredients in different languages.
*   `user_ingredients`: Stores user-specific ingredients, which can be from the global list or custom.

## Frontend

The frontend is a single-page application built with React and TypeScript. It uses Vite for development and building.

### Key Libraries:

*   **Framework:** React
*   **Language:** TypeScript
*   **Build Tool:** Vite
*   **Routing:** `@tanstack/react-router`
*   **Data Fetching:** `@tanstack/react-query`
*   **Internationalization (i18n):** `i18next`

The frontend code is located in the `web` directory and is bundled and served by the Go backend.

## Build and Deployment

The project uses a `Makefile` for common development tasks like building, running, and testing. A `Dockerfile` is provided for containerizing the application. The Dockerfile is a multi-stage build that first builds the frontend, then the backend, and finally creates a minimal production image.

The application is configured for deployment on Kubernetes, with configuration files located in the `kubernetes` directory.
