/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "../notes/templates/**/*.html.erb"
  ],
  theme: {
    extend: {
      typography: {
        DEFAULT: {
          css: {
            'code::before': {
              content: '""'
            },
            'code::after': {
              content: '""'
            }
          }
        }
      },
    },
  },
  plugins: [
    require('@tailwindcss/typography')
  ],
}

