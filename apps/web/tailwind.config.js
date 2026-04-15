/** @type {import('tailwindcss').Config} */
export default {
  content: [
    './index.html',
    './src/**/*.{js,ts,jsx,tsx}',
  ],
  important: true,
  theme: {
    extend: {
      colors: {
        primary: '#0066ff',
        'primary-hover': '#0052cc',
      },
    },
  },
  plugins: [
    require('@tailwindcss/forms'),
  ],
}

