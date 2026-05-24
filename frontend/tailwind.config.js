/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        'net-bg': '#121212',
        'net-surface': '#1e1e1e',
        'net-border': '#333333',
        'net-text': '#e0e0e0',
        'net-muted': '#888888',
        'net-accent': '#4fc3f7',
        'net-success': '#66bb6a',
        'net-warning': '#ffa726',
        'net-danger': '#ef5350',
        'net-hover': '#2a2a2a',
      },
      fontFamily: {
        'sans': ['-apple-system', 'BlinkMacSystemFont', 'Segoe UI', 'Roboto', 'sans-serif'],
        'mono': ['Cascadia Code', 'Fira Code', 'Consolas', 'monospace'],
      },
    },
  },
  plugins: [],
}
