module.exports = {
  future: {
    purgeLayersByDefault: true,
    removeDeprecatedGapUtilities: true,
  },
  plugins: [
require('@tailwindcss/typography'),
require('@tailwindcss/forms')
  ],
  content: [
   "./src/App.svelte",
  ],
};
