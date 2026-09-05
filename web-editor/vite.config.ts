import tailwindcss from '@tailwindcss/vite';
import adapter from '@sveltejs/adapter-auto';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

const isolationHeaders = {
	'Cross-Origin-Opener-Policy': 'same-origin',
	'Cross-Origin-Embedder-Policy': 'require-corp',
	'Cross-Origin-Resource-Policy': 'same-origin'
};

function crossOriginIsolation() {
	const install = (middlewares: { use: (handler: (request: unknown, response: { setHeader: (name: string, value: string) => void }, next: () => void) => void) => void }) => {
		middlewares.use((_request, response, next) => {
			for (const [name, value] of Object.entries(isolationHeaders)) response.setHeader(name, value);
			next();
		});
	};
	return {
		name: 'ferret-cross-origin-isolation',
		configureServer: (server: { middlewares: Parameters<typeof install>[0] }) => install(server.middlewares),
		configurePreviewServer: (server: { middlewares: Parameters<typeof install>[0] }) => install(server.middlewares)
	};
}

export default defineConfig({
	worker: {
		format: 'es'
	},
	server: {
		headers: isolationHeaders
	},
	preview: {
		headers: isolationHeaders
	},
	plugins: [
		crossOriginIsolation(),
		tailwindcss(),
		sveltekit({
			compilerOptions: {
				// Force runes mode for the project, except for libraries. Can be removed in svelte 6.
				runes: ({ filename }) =>
					filename.split(/[/\\]/).includes('node_modules') ? undefined : true
			},

			// adapter-auto only supports some environments, see https://svelte.dev/docs/kit/adapter-auto for a list.
			// If your environment is not supported, or you settled on a specific environment, switch out the adapter.
			// See https://svelte.dev/docs/kit/adapters for more information about adapters.
			adapter: adapter()
		})
	]
});
