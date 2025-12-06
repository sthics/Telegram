# Mini Telegram Web Client

This is the frontend web application for the Mini Telegram project. It is built using modern web technologies to provide a fast, responsive, and robust chat experience.

## 🚀 Technologies

*   **Core**: [React 19](https://react.dev/), [TypeScript](https://www.typescriptlang.org/)
*   **Build Tool**: [Vite](https://vitejs.dev/)
*   **Styling**: [Tailwind CSS](https://tailwindcss.com/)
*   **State Management**: [Zustand](https://zustand-demo.pmnd.rs/) (Global state), [TanStack Query](https://tanstack.com/query/latest) (Server state)
*   **Routing**: [React Router DOM](https://reactrouter.com/)
*   **Icons**: [Lucide React](https://lucide.dev/)
*   **HTTP Client**: [Axios](https://axios-http.com/)

## 🛠️ Getting Started

### Prerequisites

*   Node.js (v18 or higher recommended)
*   npm (or yarn/pnpm)

### Installation

1.  Navigate to the web directory:
    ```bash
    cd web
    ```

2.  Install dependencies:
    ```bash
    npm install
    ```

### Development

Start the development server:

```bash
npm run dev
```

The application will generally be available at `http://localhost:5173`.
Ensure your backend API (Gateway) is running at `http://localhost:8080`.

### Build

Build the project for production:

```bash
npm run build
```

This will run type checking (`tsc`) and bundle the application into the `dist` directory.

### Linting

Run ESLint to check for code quality issues:

```bash
npm run lint
```

## 📁 Project Structure

*   `src/features`: Feature-based architecture (Auth, Chat, etc.)
*   `src/shared`: Shared components, hooks, and utilities.
*   `src/pages`: Top-level page components.
*   `src/stores`: Global state stores.
