/** @type {import('tailwindcss').Config} */
export default {
  content: ['./src/**/*.{html,js,svelte,ts}'],
  theme: {
    extend: {
      colors: {
        'primary': '#242734',
      },
      fontFamily: {
        'sans': ['IBMPlexSans', 'sans'],
      },
    },
  },
  plugins: [],
}

