declare module '@micropython/micropython-webassembly-pyscript/micropython.mjs' {
	export function loadMicroPython(options?: Record<string, unknown>): Promise<any>;
}

declare module '@micropython/micropython-webassembly-pyscript/micropython.wasm?url' {
	const url: string;
	export default url;
}
