import esbuild from 'esbuild';

const prod = process.argv[2] === 'production';

const ctx = await esbuild.context({
  entryPoints: ['src/main.ts'],
  bundle: true,
  external: ['obsidian', 'electron', 'codemirror', '@codemirror/*'],
  format: 'cjs',
  // platform: 'node' is correct for this plugin because room-client.ts
  // imports 'ws' (Node.js WebSocket library) for custom Origin headers.
  // Browser WebSocket API doesn't support custom headers.
  // Do NOT change to 'browser' — it will break room downloads.
  platform: 'node',
  target: 'node16',
  logLevel: 'info',
  sourcemap: prod ? false : 'inline',
  treeShaking: true,
  outfile: 'main.js',
  minify: prod,
});

if (prod) {
  await ctx.rebuild();
  process.exit(0);
} else {
  await ctx.watch();
}
