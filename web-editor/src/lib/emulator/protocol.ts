import type { ProjectData } from '$lib/types';

export const SCREEN_SIZE = 240;
export const BUTTONS = ['UP', 'DOWN', 'LEFT', 'RIGHT', 'A', 'B', 'X', 'Y', 'MENU'] as const;
export type EmulatorButton = (typeof BUTTONS)[number];

export type EmulatorCommand =
	| { type: 'start'; project: ProjectData; buttons: SharedArrayBuffer }
	| { type: 'repl'; text: string };

export type EmulatorEvent =
	| { type: 'ready' }
	| { type: 'stdout'; text: string }
	| { type: 'frame'; pixels: ArrayBuffer }
	| { type: 'finished' }
	| { type: 'error'; message: string };

export function buttonIndex(button: EmulatorButton): number {
	return BUTTONS.indexOf(button);
}
