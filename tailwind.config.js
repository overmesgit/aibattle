/** @type {import('tailwindcss').Config} */
module.exports = {
    content: ["./pages/**/*.{html,js,gohtml}", "./dist/battle_viewer/**/*.js"],
    theme: {
        extend: {},
    },
    plugins: [
        require('daisyui'),
    ],
}

