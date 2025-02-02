/** @type {import('tailwindcss').Config} */
module.exports = {
    content: ["./pages/**/*.{html,js,gohtml}"],
    theme: {
        extend: {},
    },
    plugins: [
        require('daisyui'),
    ],
}

