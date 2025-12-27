/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      // ========================================
      // TYPOGRAPHY SYSTEM
      // ========================================
      fontFamily: {
        sans: ['Inter', '-apple-system', 'BlinkMacSystemFont', 'Segoe UI', 'sans-serif'],
        mono: ['JetBrains Mono', 'Fira Code', 'Consolas', 'monospace'],
      },
      fontSize: {
        // Display - hero moments, empty states
        'display': ['2.25rem', { lineHeight: '1.2', fontWeight: '600', letterSpacing: '-0.02em' }],
        // Headings
        'h1': ['1.5rem', { lineHeight: '1.3', fontWeight: '600' }],
        'h2': ['1.25rem', { lineHeight: '1.35', fontWeight: '600' }],
        'h3': ['1rem', { lineHeight: '1.4', fontWeight: '600' }],
        'h4': ['0.875rem', { lineHeight: '1.4', fontWeight: '600' }],
        // Body
        'body': ['0.875rem', { lineHeight: '1.5', fontWeight: '400' }],
        'body-sm': ['0.8125rem', { lineHeight: '1.5', fontWeight: '400' }],
        // UI
        'label': ['0.75rem', { lineHeight: '1.4', fontWeight: '500' }],
        'caption': ['0.6875rem', { lineHeight: '1.4', fontWeight: '400' }],
      },

      // ========================================
      // SPACING SYSTEM (8pt base, 4pt micro)
      // ========================================
      spacing: {
        '0': '0',
        'px': '1px',
        '0.5': '2px',   // micro
        '1': '4px',     // micro
        '1.5': '6px',   // micro
        '2': '8px',     // base unit
        '2.5': '10px',
        '3': '12px',
        '3.5': '14px',
        '4': '16px',    // 2x base
        '5': '20px',
        '6': '24px',    // 3x base
        '7': '28px',
        '8': '32px',    // 4x base
        '9': '36px',
        '10': '40px',   // 5x base
        '11': '44px',
        '12': '48px',   // 6x base
        '14': '56px',
        '16': '64px',   // 8x base
        '20': '80px',
        '24': '96px',
        '28': '112px',
        '32': '128px',
      },

      // ========================================
      // COLOR SYSTEM WITH DEPTH LAYERS
      // ========================================
      colors: {
        // Background depth layers (darkest to lightest)
        bg: {
          base: '#0a0b0d',      // Deepest background
          DEFAULT: '#0F1115',   // App background (was 'app')
          raised: '#161920',    // Cards, sidebars
          elevated: '#1c1f28',  // Hover states, dropdowns
          overlay: '#242832',   // Active states, overlays
        },
        // Surface (for components)
        surface: {
          DEFAULT: '#161920',
          hover: '#1c1f28',
          active: '#242832',
        },
        // Text hierarchy
        text: {
          primary: '#f4f4f5',   // Slightly off-white (zinc-100)
          secondary: '#a1a1aa', // zinc-400
          tertiary: '#71717a', // zinc-500
          disabled: '#52525b', // zinc-600
          inverse: '#0F1115',  // For light backgrounds
        },
        // Border system with opacity
        border: {
          subtle: 'rgba(255, 255, 255, 0.06)',
          DEFAULT: 'rgba(255, 255, 255, 0.1)',
          strong: 'rgba(255, 255, 255, 0.15)',
          focus: 'rgba(59, 130, 246, 0.5)',
        },
        // Brand colors with full scale
        brand: {
          50: '#eff6ff',
          100: '#dbeafe',
          200: '#bfdbfe',
          300: '#93c5fd',
          400: '#60a5fa',
          500: '#3b82f6',   // Primary
          600: '#2563eb',   // Hover
          700: '#1d4ed8',   // Active/pressed
          800: '#1e40af',
          900: '#1e3a8a',
          950: '#172554',
          // Semantic aliases
          DEFAULT: '#3b82f6',
          primary: '#3b82f6',
          hover: '#2563eb',
          active: '#1d4ed8',
        },
        // Semantic status colors
        success: {
          light: '#4ade80',
          DEFAULT: '#22c55e',
          dark: '#16a34a',
        },
        warning: {
          light: '#fbbf24',
          DEFAULT: '#f59e0b',
          dark: '#d97706',
        },
        error: {
          light: '#f87171',
          DEFAULT: '#ef4444',
          dark: '#dc2626',
        },
        info: {
          light: '#60a5fa',
          DEFAULT: '#3b82f6',
          dark: '#2563eb',
        },
        // Legacy support (map old names)
        app: '#0F1115',
        status: {
          success: '#22c55e',
          error: '#ef4444',
        },
      },

      // ========================================
      // SHADOW & DEPTH SYSTEM
      // ========================================
      boxShadow: {
        'none': 'none',
        'xs': '0 1px 2px rgba(0, 0, 0, 0.4)',
        'sm': '0 2px 4px rgba(0, 0, 0, 0.4), 0 1px 2px rgba(0, 0, 0, 0.3)',
        'DEFAULT': '0 4px 8px rgba(0, 0, 0, 0.4), 0 2px 4px rgba(0, 0, 0, 0.3)',
        'md': '0 4px 8px rgba(0, 0, 0, 0.4), 0 2px 4px rgba(0, 0, 0, 0.3)',
        'lg': '0 8px 16px rgba(0, 0, 0, 0.4), 0 4px 8px rgba(0, 0, 0, 0.3)',
        'xl': '0 16px 32px rgba(0, 0, 0, 0.4), 0 8px 16px rgba(0, 0, 0, 0.3)',
        '2xl': '0 24px 48px rgba(0, 0, 0, 0.5), 0 12px 24px rgba(0, 0, 0, 0.4)',
        // Glow effects for focus/active states
        'glow-brand': '0 0 20px rgba(59, 130, 246, 0.35)',
        'glow-brand-sm': '0 0 10px rgba(59, 130, 246, 0.25)',
        'glow-success': '0 0 20px rgba(34, 197, 94, 0.35)',
        'glow-error': '0 0 20px rgba(239, 68, 68, 0.35)',
        // Inner shadows for pressed states
        'inner-sm': 'inset 0 1px 2px rgba(0, 0, 0, 0.3)',
        'inner': 'inset 0 2px 4px rgba(0, 0, 0, 0.3)',
        // Focus ring shadow (combines with ring utilities)
        'focus': '0 0 0 2px rgba(59, 130, 246, 0.4)',
      },

      // ========================================
      // BORDER RADIUS
      // ========================================
      borderRadius: {
        'none': '0',
        'sm': '4px',
        'DEFAULT': '6px',
        'md': '8px',
        'lg': '12px',
        'xl': '16px',
        '2xl': '20px',
        '3xl': '24px',
        'full': '9999px',
      },

      // ========================================
      // ANIMATIONS
      // ========================================
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        fadeOut: {
          '0%': { opacity: '1' },
          '100%': { opacity: '0' },
        },
        slideUp: {
          '0%': { transform: 'translateY(8px)', opacity: '0' },
          '100%': { transform: 'translateY(0)', opacity: '1' },
        },
        slideDown: {
          '0%': { transform: 'translateY(-8px)', opacity: '0' },
          '100%': { transform: 'translateY(0)', opacity: '1' },
        },
        slideInRight: {
          '0%': { transform: 'translateX(100%)', opacity: '0' },
          '100%': { transform: 'translateX(0)', opacity: '1' },
        },
        slideOutRight: {
          '0%': { transform: 'translateX(0)', opacity: '1' },
          '100%': { transform: 'translateX(100%)', opacity: '0' },
        },
        scaleIn: {
          '0%': { transform: 'scale(0.95)', opacity: '0' },
          '100%': { transform: 'scale(1)', opacity: '1' },
        },
        scaleOut: {
          '0%': { transform: 'scale(1)', opacity: '1' },
          '100%': { transform: 'scale(0.95)', opacity: '0' },
        },
        shimmer: {
          '0%': { backgroundPosition: '-200% 0' },
          '100%': { backgroundPosition: '200% 0' },
        },
        pulse: {
          '0%, 100%': { opacity: '1' },
          '50%': { opacity: '0.5' },
        },
        pulseSoft: {
          '0%, 100%': { opacity: '1' },
          '50%': { opacity: '0.7' },
        },
        bounce: {
          '0%, 100%': { transform: 'translateY(0)' },
          '50%': { transform: 'translateY(-4px)' },
        },
        pop: {
          '0%': { transform: 'scale(1)' },
          '50%': { transform: 'scale(1.15)' },
          '100%': { transform: 'scale(1)' },
        },
        spin: {
          '0%': { transform: 'rotate(0deg)' },
          '100%': { transform: 'rotate(360deg)' },
        },
        typing: {
          '0%': { opacity: '0.3' },
          '50%': { opacity: '1' },
          '100%': { opacity: '0.3' },
        },
        ripple: {
          '0%': { transform: 'scale(0)', opacity: '0.5' },
          '100%': { transform: 'scale(1)', opacity: '0' },
        },
        shake: {
          '0%, 100%': { transform: 'translateX(0)' },
          '25%': { transform: 'translateX(-4px)' },
          '75%': { transform: 'translateX(4px)' },
        },
        wiggle: {
          '0%, 100%': { transform: 'rotate(0deg)' },
          '25%': { transform: 'rotate(-3deg)' },
          '75%': { transform: 'rotate(3deg)' },
        },
      },
      animation: {
        'fade-in': 'fadeIn 150ms ease-out',
        'fade-out': 'fadeOut 150ms ease-in',
        'slide-up': 'slideUp 200ms ease-out',
        'slide-down': 'slideDown 200ms ease-out',
        'slide-in-right': 'slideInRight 200ms ease-out',
        'slide-out-right': 'slideOutRight 150ms ease-in',
        'scale-in': 'scaleIn 150ms ease-out',
        'scale-out': 'scaleOut 100ms ease-in',
        'shimmer': 'shimmer 2s infinite linear',
        'pulse': 'pulse 2s infinite ease-in-out',
        'pulse-soft': 'pulseSoft 2s infinite ease-in-out',
        'bounce': 'bounce 600ms ease-in-out',
        'pop': 'pop 300ms ease-out',
        'spin': 'spin 1s linear infinite',
        'typing-1': 'typing 1.4s infinite ease-in-out',
        'typing-2': 'typing 1.4s infinite ease-in-out 0.2s',
        'typing-3': 'typing 1.4s infinite ease-in-out 0.4s',
        'ripple': 'ripple 600ms ease-out forwards',
        'shake': 'shake 400ms ease-in-out',
        'wiggle': 'wiggle 200ms ease-in-out',
      },

      // ========================================
      // TRANSITIONS
      // ========================================
      transitionDuration: {
        '0': '0ms',
        '75': '75ms',
        '100': '100ms',
        '150': '150ms',
        '200': '200ms',
        '250': '250ms',
        '300': '300ms',
        '400': '400ms',
        '500': '500ms',
      },
      transitionTimingFunction: {
        'ease-out-expo': 'cubic-bezier(0.16, 1, 0.3, 1)',
        'ease-in-expo': 'cubic-bezier(0.7, 0, 0.84, 0)',
        'ease-out-back': 'cubic-bezier(0.34, 1.56, 0.64, 1)',
      },

      // ========================================
      // Z-INDEX SCALE
      // ========================================
      zIndex: {
        'behind': '-1',
        '0': '0',
        '10': '10',
        '20': '20',
        '30': '30',
        '40': '40',
        'dropdown': '50',
        'sticky': '100',
        'overlay': '200',
        'modal': '300',
        'popover': '400',
        'toast': '500',
        'tooltip': '600',
      },
    },
  },
  plugins: [],
}
