import { mkdir, writeFile } from 'node:fs/promises';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

import { OpenAPI } from './index.js';

const packageRoot = fileURLToPath(new URL('../', import.meta.url));

const outputPaths = [
  resolve(packageRoot, 'openapi.json'),
  resolve(packageRoot, '../../apps/backend/static/openapi.json'),
];

const document = `${JSON.stringify(OpenAPI, null, 2)}\n`;

await Promise.all(
  outputPaths.map(async (filePath) => {
    await mkdir(dirname(filePath), { recursive: true });

    await writeFile(filePath, document, 'utf8');

    console.info(`Generated: ${filePath}`);
  }),
);
