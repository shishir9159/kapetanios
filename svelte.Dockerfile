FROM oven/bun:1 AS base
WORKDIR /usr/src/app

FROM base AS install
#RUN mkdir -p /tmp/dev
#COPY web/svelte/package.json web/svelte/bun.lock /tmp/dev/
#RUN cd /tmp/dev && bun install --frozen-lockfile

# install with --production (exclude devDependencies)
RUN mkdir -p /tmp/prod
COPY web/svelte/package.json web/svelte/bun.lock /tmp/prod/
#RUN cd /tmp/prod && bun install --frozen-lockfile --production --modules-dir=node_modules
RUN cd /tmp/prod && bun install

FROM base AS prerelease
#COPY --from=install /tmp/dev/node_modules node_modules
COPY --from=install /tmp/prod/node_modules node_modules
COPY web/svelte/ .
COPY web/svelte/.* .
#COPY web/svelte/.svelte-kit .

## [optional] tests & build
ENV NODE_ENV=production
RUN bun test
RUN bun run build

FROM base AS release
COPY --from=install /tmp/prod/node_modules node_modules
COPY --from=prerelease /usr/src/app/build build
COPY --from=prerelease /usr/src/app/.svelte-kit svelte-kit
COPY --from=prerelease /usr/src/app/package.json .

ENV NODE_ENV=production

# run the app
USER bun
EXPOSE 3000/tcp
ENTRYPOINT [ "bun", "run", "build/index.js" ]
#CMD ["/bin/bash", "-c", "while sleep 3600; do :; done"]