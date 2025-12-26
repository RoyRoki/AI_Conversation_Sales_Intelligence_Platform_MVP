import type { Config } from "tailwindcss";

const config: Config = {
  content: [
    "./pages/**/*.{js,ts,jsx,tsx,mdx}",
    "./components/**/*.{js,ts,jsx,tsx,mdx}",
    "./app/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      colors: {
        primary: {
          DEFAULT: "#2563EB",
          light: "#3B82F6",
          dark: "#1E40AF",
        },
        accent: "#60A5FA",
      },
      backgroundImage: {
        "gradient-blue": "linear-gradient(135deg, #3B82F6 0%, #2563EB 100%)",
        "gradient-blue-light": "linear-gradient(135deg, #60A5FA 0%, #3B82F6 100%)",
        "gradient-blue-dark": "linear-gradient(135deg, #1E40AF 0%, #1D4ED8 100%)",
      },
      boxShadow: {
        professional: "0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06)",
        "professional-lg": "0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05)",
        "professional-xl": "0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04)",
      },
    },
  },
  plugins: [],
};

export default config;

