/**
 * Design System Tokens
 *
 * These tokens mirror the Tailwind config and provide runtime access
 * to design system values for JavaScript/TypeScript usage.
 */

// ========================================
// SPACING (8pt base grid)
// ========================================
export const spacing = {
  0: '0',
  px: '1px',
  0.5: '2px',
  1: '4px',
  1.5: '6px',
  2: '8px',
  2.5: '10px',
  3: '12px',
  3.5: '14px',
  4: '16px',
  5: '20px',
  6: '24px',
  7: '28px',
  8: '32px',
  9: '36px',
  10: '40px',
  11: '44px',
  12: '48px',
  14: '56px',
  16: '64px',
  20: '80px',
  24: '96px',
} as const;

// ========================================
// TYPOGRAPHY
// ========================================
export const typography = {
  fontFamily: {
    sans: "'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif",
    mono: "'JetBrains Mono', 'Fira Code', Consolas, monospace",
  },
  fontSize: {
    display: { size: '2.25rem', lineHeight: 1.2, weight: 600, letterSpacing: '-0.02em' },
    h1: { size: '1.5rem', lineHeight: 1.3, weight: 600 },
    h2: { size: '1.25rem', lineHeight: 1.35, weight: 600 },
    h3: { size: '1rem', lineHeight: 1.4, weight: 600 },
    h4: { size: '0.875rem', lineHeight: 1.4, weight: 600 },
    body: { size: '0.875rem', lineHeight: 1.5, weight: 400 },
    bodySm: { size: '0.8125rem', lineHeight: 1.5, weight: 400 },
    label: { size: '0.75rem', lineHeight: 1.4, weight: 500 },
    caption: { size: '0.6875rem', lineHeight: 1.4, weight: 400 },
  },
} as const;

// ========================================
// COLORS
// ========================================
export const colors = {
  // Background depth layers
  bg: {
    base: '#0a0b0d',
    default: '#0F1115',
    raised: '#161920',
    elevated: '#1c1f28',
    overlay: '#242832',
  },
  // Surface colors
  surface: {
    default: '#161920',
    hover: '#1c1f28',
    active: '#242832',
  },
  // Text hierarchy
  text: {
    primary: '#f4f4f5',
    secondary: '#a1a1aa',
    tertiary: '#71717a',
    disabled: '#52525b',
    inverse: '#0F1115',
  },
  // Border colors
  border: {
    subtle: 'rgba(255, 255, 255, 0.06)',
    default: 'rgba(255, 255, 255, 0.1)',
    strong: 'rgba(255, 255, 255, 0.15)',
    focus: 'rgba(59, 130, 246, 0.5)',
  },
  // Brand scale
  brand: {
    50: '#eff6ff',
    100: '#dbeafe',
    200: '#bfdbfe',
    300: '#93c5fd',
    400: '#60a5fa',
    500: '#3b82f6',
    600: '#2563eb',
    700: '#1d4ed8',
    800: '#1e40af',
    900: '#1e3a8a',
    950: '#172554',
  },
  // Semantic colors
  success: { light: '#4ade80', default: '#22c55e', dark: '#16a34a' },
  warning: { light: '#fbbf24', default: '#f59e0b', dark: '#d97706' },
  error: { light: '#f87171', default: '#ef4444', dark: '#dc2626' },
  info: { light: '#60a5fa', default: '#3b82f6', dark: '#2563eb' },
} as const;

// ========================================
// SHADOWS
// ========================================
export const shadows = {
  none: 'none',
  xs: '0 1px 2px rgba(0, 0, 0, 0.4)',
  sm: '0 2px 4px rgba(0, 0, 0, 0.4), 0 1px 2px rgba(0, 0, 0, 0.3)',
  md: '0 4px 8px rgba(0, 0, 0, 0.4), 0 2px 4px rgba(0, 0, 0, 0.3)',
  lg: '0 8px 16px rgba(0, 0, 0, 0.4), 0 4px 8px rgba(0, 0, 0, 0.3)',
  xl: '0 16px 32px rgba(0, 0, 0, 0.4), 0 8px 16px rgba(0, 0, 0, 0.3)',
  '2xl': '0 24px 48px rgba(0, 0, 0, 0.5), 0 12px 24px rgba(0, 0, 0, 0.4)',
  // Glows
  glowBrand: '0 0 20px rgba(59, 130, 246, 0.35)',
  glowBrandSm: '0 0 10px rgba(59, 130, 246, 0.25)',
  glowSuccess: '0 0 20px rgba(34, 197, 94, 0.35)',
  glowError: '0 0 20px rgba(239, 68, 68, 0.35)',
  // Inner
  innerSm: 'inset 0 1px 2px rgba(0, 0, 0, 0.3)',
  inner: 'inset 0 2px 4px rgba(0, 0, 0, 0.3)',
} as const;

// ========================================
// BORDER RADIUS
// ========================================
export const radii = {
  none: '0',
  sm: '4px',
  default: '6px',
  md: '8px',
  lg: '12px',
  xl: '16px',
  '2xl': '20px',
  '3xl': '24px',
  full: '9999px',
} as const;

// ========================================
// TRANSITIONS
// ========================================
export const transitions = {
  duration: {
    fast: '100ms',
    normal: '150ms',
    slow: '250ms',
    slower: '400ms',
  },
  easing: {
    default: 'ease-out',
    in: 'ease-in',
    out: 'ease-out',
    inOut: 'ease-in-out',
    outExpo: 'cubic-bezier(0.16, 1, 0.3, 1)',
    inExpo: 'cubic-bezier(0.7, 0, 0.84, 0)',
    outBack: 'cubic-bezier(0.34, 1.56, 0.64, 1)',
  },
} as const;

// ========================================
// Z-INDEX SCALE
// ========================================
export const zIndex = {
  behind: -1,
  base: 0,
  dropdown: 50,
  sticky: 100,
  overlay: 200,
  modal: 300,
  popover: 400,
  toast: 500,
  tooltip: 600,
} as const;

// ========================================
// BREAKPOINTS
// ========================================
export const breakpoints = {
  sm: '640px',
  md: '768px',
  lg: '1024px',
  xl: '1280px',
  '2xl': '1536px',
} as const;

// ========================================
// COMPONENT SIZES
// ========================================
export const componentSizes = {
  button: {
    sm: { height: '32px', padding: '0 12px', fontSize: '0.8125rem' },
    default: { height: '36px', padding: '0 16px', fontSize: '0.875rem' },
    lg: { height: '40px', padding: '0 20px', fontSize: '0.875rem' },
    icon: { size: '36px' },
    iconSm: { size: '32px' },
    iconLg: { size: '40px' },
  },
  input: {
    sm: { height: '32px', padding: '0 12px' },
    default: { height: '40px', padding: '0 16px' },
    lg: { height: '48px', padding: '0 20px' },
  },
  avatar: {
    xs: '24px',
    sm: '32px',
    md: '40px',
    lg: '48px',
    xl: '64px',
    '2xl': '80px',
  },
  iconSize: {
    xs: '14px',
    sm: '16px',
    default: '20px',
    lg: '24px',
    xl: '32px',
  },
} as const;

// ========================================
// LAYOUT
// ========================================
export const layout = {
  sidebar: {
    width: '320px',
    minWidth: '280px',
    collapsedWidth: '72px',
  },
  header: {
    height: '56px',
  },
  messageInput: {
    minHeight: '48px',
    maxHeight: '200px',
  },
  contentMaxWidth: '1200px',
} as const;

// ========================================
// EXPORTS
// ========================================
export const tokens = {
  spacing,
  typography,
  colors,
  shadows,
  radii,
  transitions,
  zIndex,
  breakpoints,
  componentSizes,
  layout,
} as const;

export default tokens;
