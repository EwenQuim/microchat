import { defineConfig } from "orval";

export default defineConfig({
	microchat: {
		input: {
			target: "../doc/openapi.json",
		},
		output: {
			mode: "tags-split",
			target: "./src/lib/api/generated",
			client: "react-query",
			clean: true,
			httpClient: "fetch",
			biome: true,
			override: {
				mutator: {
					path: "./src/lib/api/mutator.ts",
					name: "customFetch",
				},
			},
		},
	},
});
