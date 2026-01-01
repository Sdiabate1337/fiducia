import type { Config } from "tailwindcss";

const config: Config = {
    content: [
        "./src/pages/**/*.{js,ts,jsx,tsx,mdx}",
        "./src/components/**/*.{js,ts,jsx,tsx,mdx}",
        "./src/app/**/*.{js,ts,jsx,tsx,mdx}",
    ],
    theme: {
        extend: {
            colors: {
                background: "var(--background)",
                foreground: "var(--foreground)",
                primary: {
                    DEFAULT: "#0052FF", // Hyper Blue
                    foreground: "#FFFFFF",
                },
                secondary: {
                    DEFAULT: "rgba(255, 255, 255, 0.1)", // Glass
                    foreground: "#FFFFFF",
                },
                accent: {
                    DEFAULT: "#020617", // Slate
                    foreground: "#FFFFFF",
                },
                success: "#20E070", // Arctic Teal approx
                cream: "#F9F8F6",
                editorial: "#1A1A1A",
                forest: "#1A4D2E",
                bordeaux: "#4F2830",
            },
            fontFamily: {
                sans: ["var(--font-jakarta)", "sans-serif"],
                body: ["var(--font-inter)", "sans-serif"],
                serif: ["var(--font-playfair)", "serif"],
            },
            backgroundImage: {
                "gradient-radial": "radial-gradient(var(--tw-gradient-stops))",
                "gradient-conic":
                    "conic-gradient(from 180deg at 50% 50%, var(--tw-gradient-stops))",
            },
        },
    },
    plugins: [],
};
export default config;
