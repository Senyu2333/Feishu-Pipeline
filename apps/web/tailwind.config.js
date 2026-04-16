/** @type {import('tailwindcss').Config} */
export default {
  content: [
    './index.html',
    './src/**/*.{js,ts,jsx,tsx}',
  ],
  important: true,
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        primary: '#0058bc',
        'primary-container': '#0070eb',
        'on-primary': '#ffffff',
        'on-primary-container': '#fefcff',
        secondary: '#405e96',
        'secondary-container': '#a1befd',
        'on-secondary': '#ffffff',
        'on-secondary-container': '#2d4c83',
        tertiary: '#9e3d00',
        'tertiary-container': '#c64f00',
        'on-tertiary': '#ffffff',
        'on-tertiary-container': '#fffbff',
        background: '#f4faff',
        'on-background': '#001f2a',
        surface: '#f4faff',
        'surface-dim': '#c0dfee',
        'surface-bright': '#f4faff',
        'surface-container': '#d9f2ff',
        'surface-container-low': '#e6f6ff',
        'surface-container-lowest': '#ffffff',
        'surface-container-high': '#ceedfd',
        'surface-container-highest': '#c9e7f7',
        'on-surface': '#001f2a',
        'on-surface-variant': '#414755',
        outline: '#717786',
        'outline-variant': '#c1c6d7',
        error: '#ba1a1a',
        'error-container': '#ffdad6',
        'on-error': '#ffffff',
        'on-error-container': '#93000a',
      },
    },
  },
  plugins: [
    require('@tailwindcss/forms'),
  ],
}

