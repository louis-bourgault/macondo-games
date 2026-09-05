import { readFile, writeFile } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';
import { dirname, resolve } from 'node:path';

const here = dirname(fileURLToPath(import.meta.url));
const sourcePath = resolve(here, '../../software/internal/helpers/text.go');
const outputPath = resolve(here, '../src/lib/emulator/font.generated.ts');
const apiSourcePath = resolve(here, '../../software/cmd/embed/host.c');
const apiOutputPath = resolve(here, '../src/lib/emulator/api.generated.ts');
const source = await readFile(sourcePath, 'utf8');
const table = source.match(/var font8x8_basic = \[128\]\[8\]byte\{([\s\S]*?)\n\}/);

if (!table) throw new Error(`Could not find font8x8_basic in ${sourcePath}`);

const rows = [...table[1].matchAll(/\{([^}]+)\}/g)].map((entry) =>
	[...entry[1].matchAll(/0x[0-9a-f]+/gi)].map((value) => Number.parseInt(value[0], 16))
);

if (rows.length !== 128 || rows.some((row) => row.length !== 8)) {
	throw new Error(`Expected 128 glyphs of 8 bytes, found ${rows.length}`);
}

const generated = `// Generated from software/internal/helpers/text.go. Do not edit.\n` +
	`export const FONT_8X8 = new Uint8Array(${JSON.stringify(rows.flat())});\n`;
await writeFile(outputPath, generated);

const apiSource = await readFile(apiSourcePath, 'utf8');
const registration = apiSource.match(/void register_ferret_module\(void\) \{([\s\S]*?)\n\}/);
if (!registration) throw new Error(`Could not find register_ferret_module in ${apiSourcePath}`);
const names = [...registration[1].matchAll(/qstr_from_str\("([a-z0-9_]+)"\)/g)]
	.map((match) => match[1])
	.filter((name) => name !== 'ferret');
await writeFile(apiOutputPath,
	`// Generated from software/cmd/embed/host.c. Do not edit.\n` +
	`export const FERRET_API = ${JSON.stringify(names)} as const;\n` +
	`export type FerretApiName = (typeof FERRET_API)[number];\n`
);
