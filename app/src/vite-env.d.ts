/// <reference types="vite/client" />
/// <reference types="vite-plugin-pwa/client" />

interface ImportMetaEnv {
	readonly VITE_ENABLE_PWA: string | undefined;
}

interface ImportMeta {
	readonly env: ImportMetaEnv;
}
