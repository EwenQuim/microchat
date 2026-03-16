import { fileURLToPath, URL } from "node:url";
import babel from "@rolldown/plugin-babel";
import tailwindcss from "@tailwindcss/vite";
import { devtools } from "@tanstack/devtools-vite";
import { tanstackRouter } from "@tanstack/router-plugin/vite";
import viteReact, { reactCompilerPreset } from "@vitejs/plugin-react";
import { defineConfig } from "vite";
import { VitePWA } from "vite-plugin-pwa";

const enablePWA = process.env.VITE_ENABLE_PWA === "true";

// https://vitejs.dev/config/
export default defineConfig(() => ({
	base: process.env.BASE_URL || "/",
	plugins: [
		devtools(),
		tanstackRouter({
			target: "react",
			autoCodeSplitting: true,
		}),
		viteReact(),
		// eslint-disable-next-line @typescript-eslint/ban-ts-comment
		// @ts-ignore — @rolldown/plugin-babel types have a spurious required-fields error
		babel({ presets: [reactCompilerPreset()] }),
		tailwindcss(),
		...(enablePWA
			? [
					VitePWA({
						registerType: "autoUpdate",
						includeAssets: [
							"favicon.ico",
							"logo192.png",
							"logo512.png",
							"logo1024.png",
							"robots.txt",
						],
						manifest: false,
						workbox: {
							globPatterns: ["**/*.{js,css,html,ico,png,svg,woff,woff2}"],
							navigateFallbackDenylist: [/^\/api/],
						},
						devOptions: {
							enabled: false,
							type: "module",
						},
					}),
				]
			: []),
	],
	resolve: {
		alias: {
			"@": fileURLToPath(new URL("./src", import.meta.url)),
		},
	},
}));
