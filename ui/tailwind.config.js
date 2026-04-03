/** @type {import('tailwindcss').Config} */
export default {
  darkMode: 'class',
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        // 基于截图的深蓝色主题
        background: {
          DEFAULT: '#1e2329',
          secondary: '#2b3139',
          tertiary: '#363c47',
        },
        surface: {
          DEFAULT: '#2b3139',
          hover: '#3a4048',
          active: '#4b5563',
        },
        accent: {
          DEFAULT: '#f7931e',
          hover: '#ff9f3a',
          muted: '#e6851a',
        },
        text: {
          primary: '#f8f9fa',
          secondary: '#d1d5db',
          muted: '#9ca3af',
        },
        border: {
          DEFAULT: '#374151',
          light: '#4b5563',
        },
      },
    },
  },
  plugins: [],
}
