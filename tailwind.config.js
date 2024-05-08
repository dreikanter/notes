/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "../notes/templates/**/*.html.erb"
  ],
  safelist: [
    "bg-slate-100",
    "bg-slate-500",
    "block",
    "border",
    "border-slate-200",
    "border-slate-600",
    "font-normal",
    "no-underline",
    "rounded-md",
    "text-slate-100",
    "text-slate-400",
    "text-slate-800",
    {
      "pattern": /m(l|r|x|y)-.+/,
    },
    {
      "pattern": /p(l|r|x|y)-.+/,
    },
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
