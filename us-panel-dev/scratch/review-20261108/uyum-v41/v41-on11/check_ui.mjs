// Syntax and orphaned static DOM bindings; no browser or packages required.
import { readFileSync } from 'node:fs';
import { Script } from 'node:vm';

const html = readFileSync(new URL('./gui.html', import.meta.url), 'utf8');
const script = html.match(/<script>([\s\S]*?)<\/script>/)?.[1];
if (!script) throw new Error('Embedded UI script missing');
new Script(script, { filename: 'gui.html/script' });
const ids = [...html.matchAll(/\bid="([^"]+)"/g)].map(m => m[1]);
if (new Set(ids).size !== ids.length) throw new Error('Duplicate HTML element ID');
for (const match of script.matchAll(/document\.querySelector\(["']#([\w-]+)/g)) {
  if (!ids.includes(match[1])) throw new Error(`Missing UI element: ${match[1]}`);
}
console.log('Embedded JavaScript syntax and static DOM bindings passed.');
