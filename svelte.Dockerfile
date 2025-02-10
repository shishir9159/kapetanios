FROM oven/bun

WORKDIR /app
COPY web/svelte/package.json package.json
RUN bun install

COPY web/svelte/. .
RUN bun run build

EXPOSE 3000
ENTRYPOINT ["bun", "./build"]
CMD ./sveltekit-bun
