import type { Handle } from '@sveltejs/kit';

// Live emulator input is shared with a worker even while MicroPython is inside
// a tight loop. Production static hosts must serve these same two headers.
export const handle: Handle = async ({ event, resolve }) => {
	const response = await resolve(event);
	response.headers.set('Cross-Origin-Opener-Policy', 'same-origin');
	response.headers.set('Cross-Origin-Embedder-Policy', 'require-corp');
	return response;
};
