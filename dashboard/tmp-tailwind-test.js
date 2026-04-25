const plugin = require('@tailwindcss/postcss');
const postcss = require('postcss');
console.log('plugin type', typeof plugin, plugin.postcss, plugin.name || plugin.postcssPlugin);
const pp = plugin({});
console.log('plugin result type', typeof pp, Object.keys(pp).slice(0, 10));
postcss([pp]).process('@tailwind base; @tailwind components; @tailwind utilities;', { from: undefined })
  .then(res => console.log('done', res.css.slice(0, 100)))
  .catch(err => {
    console.error('error stack');
    console.error(err.stack);
    console.error('message', err.message);
  });
