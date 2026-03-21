import type { Config } from 'tailwindcss'

// ── Luminance Design System ── Claymorphism / Educational Platform ──────────
export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        primary:   '#2563EB', // Blue — links, active states
        secondary: '#3B82F6', // Lighter blue — hover
        cta:       '#F97316', // Orange — primary CTA buttons
        surface:   '#F8FAFC', // Off-white background
        text:      '#1E293B', // Dark slate body text
      },
      fontFamily: {
        heading: ['"Baloo 2"',   'sans-serif'],
        body:    ['"Comic Neue"', 'cursive'],
      },
      borderRadius: {
        clay: '20px', // Claymorphism default
      },
      borderWidth: {
        clay: '3px',
      },
      boxShadow: {
        // Claymorphism double-shadow
        clay:     '4px 4px 0px 0px rgba(37,99,235,0.25), 8px 8px 0px 0px rgba(37,99,235,0.10)',
        'clay-cta': '4px 4px 0px 0px rgba(249,115,22,0.30), 8px 8px 0px 0px rgba(249,115,22,0.12)',
      },
      transitionTimingFunction: {
        'clay-out': 'cubic-bezier(0.34, 1.56, 0.64, 1)',
      },
    },
  },
  plugins: [],
} satisfies Config
