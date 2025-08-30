# sv

bun version 1.2.23

```bash
bunx sv create svelte
```

## Developing

```bash
bun run dev --open
svelte-kit sync || echo ''
bun install --lockfile-only
```

## Building

To create a production version of your app:

```bash
bun run build
```

## Cleaning

```bash
bunx eslint --fix ../../.gitignore
bunx prettier --write .
```

You can preview the production build with `bun run preview`.
