
// this file is generated — do not edit it


/// <reference types="@sveltejs/kit" />

/**
 * Environment variables [loaded by Vite](https://vitejs.dev/guide/env-and-mode.html#env-files) from `.env` files and `process.env`. Like [`$env/dynamic/private`](https://svelte.dev/docs/kit/$env-dynamic-private), this module cannot be imported into client-side code. This module only includes variables that _do not_ begin with [`config.kit.env.publicPrefix`](https://svelte.dev/docs/kit/configuration#env) _and do_ start with [`config.kit.env.privatePrefix`](https://svelte.dev/docs/kit/configuration#env) (if configured).
 * 
 * _Unlike_ [`$env/dynamic/private`](https://svelte.dev/docs/kit/$env-dynamic-private), the values exported from this module are statically injected into your bundle at build time, enabling optimisations like dead code elimination.
 * 
 * ```ts
 * import { API_KEY } from '$env/static/private';
 * ```
 * 
 * Note that all environment variables referenced in your code should be declared (for example in an `.env` file), even if they don't have a value until the app is deployed:
 * 
 * ```
 * MY_FEATURE_FLAG=""
 * ```
 * 
 * You can override `.env` values from the command line like so:
 * 
 * ```bash
 * MY_FEATURE_FLAG="enabled" npm run dev
 * ```
 */
declare module '$env/static/private' {
	export const QT_AUTO_SCREEN_SCALE_FACTOR: string;
	export const XDG_SESSION_PATH: string;
	export const KDE_SESSION_UID: string;
	export const LC_TIME: string;
	export const KDE_APPLICATIONS_AS_SCOPE: string;
	export const XDG_SEAT_PATH: string;
	export const MAIL: string;
	export const FORCE_COLOR: string;
	export const PATH: string;
	export const LOGNAME: string;
	export const XDG_MENU_PREFIX: string;
	export const XDG_CONFIG_DIRS: string;
	export const XAUTHORITY: string;
	export const LC_PAPER: string;
	export const LC_MEASUREMENT: string;
	export const XDG_SESSION_ID: string;
	export const npm_config_color: string;
	export const MOTD_SHOWN: string;
	export const XDG_SEAT: string;
	export const XDG_VTNR: string;
	export const XDG_SESSION_DESKTOP: string;
	export const DBUS_SESSION_BUS_ADDRESS: string;
	export const QT_LINUX_ACCESSIBILITY_ALWAYS_ON: string;
	export const LC_TELEPHONE: string;
	export const INVOCATION_ID: string;
	export const LC_ADDRESS: string;
	export const ICEAUTHORITY: string;
	export const XDG_DATA_DIRS: string;
	export const SHELL: string;
	export const DEBUG_COLORS: string;
	export const SESSION_MANAGER: string;
	export const GTK_MODULES: string;
	export const XDG_SESSION_CLASS: string;
	export const COLORTERM: string;
	export const LC_IDENTIFICATION: string;
	export const DISPLAY: string;
	export const HOME: string;
	export const MEMORY_PRESSURE_WATCH: string;
	export const XDG_CURRENT_DESKTOP: string;
	export const DEBUGINFOD_URLS: string;
	export const KDE_SESSION_VERSION: string;
	export const LC_NAME: string;
	export const LC_NUMERIC: string;
	export const MOCHA_COLORS: string;
	export const GTK_RC_FILES: string;
	export const LANG: string;
	export const GTK3_MODULES: string;
	export const GTK2_RC_FILES: string;
	export const QT_WAYLAND_RECONNECT: string;
	export const MEMORY_PRESSURE_WRITE: string;
	export const SYSTEMD_EXEC_PID: string;
	export const XDG_RUNTIME_DIR: string;
	export const PAM_KWALLET5_LOGIN: string;
	export const OLDPWD: string;
	export const MANAGERPID: string;
	export const DESKTOP_SESSION: string;
	export const USER: string;
	export const KDE_FULL_SESSION: string;
	export const XDG_SESSION_TYPE: string;
	export const LC_MONETARY: string;
	export const PWD: string;
	export const JOURNAL_STREAM: string;
	export const NODE_ENV: string;
}

/**
 * Similar to [`$env/static/private`](https://svelte.dev/docs/kit/$env-static-private), except that it only includes environment variables that begin with [`config.kit.env.publicPrefix`](https://svelte.dev/docs/kit/configuration#env) (which defaults to `PUBLIC_`), and can therefore safely be exposed to client-side code.
 * 
 * Values are replaced statically at build time.
 * 
 * ```ts
 * import { PUBLIC_BASE_URL } from '$env/static/public';
 * ```
 */
declare module '$env/static/public' {
	
}

/**
 * This module provides access to runtime environment variables, as defined by the platform you're running on. For example if you're using [`adapter-node`](https://github.com/sveltejs/kit/tree/main/packages/adapter-node) (or running [`vite preview`](https://svelte.dev/docs/kit/cli)), this is equivalent to `process.env`. This module only includes variables that _do not_ begin with [`config.kit.env.publicPrefix`](https://svelte.dev/docs/kit/configuration#env) _and do_ start with [`config.kit.env.privatePrefix`](https://svelte.dev/docs/kit/configuration#env) (if configured).
 * 
 * This module cannot be imported into client-side code.
 * 
 * Dynamic environment variables cannot be used during prerendering.
 * 
 * ```ts
 * import { env } from '$env/dynamic/private';
 * console.log(env.DEPLOYMENT_SPECIFIC_VARIABLE);
 * ```
 * 
 * > In `dev`, `$env/dynamic` always includes environment variables from `.env`. In `prod`, this behavior will depend on your adapter.
 */
declare module '$env/dynamic/private' {
	export const env: {
		QT_AUTO_SCREEN_SCALE_FACTOR: string;
		XDG_SESSION_PATH: string;
		KDE_SESSION_UID: string;
		LC_TIME: string;
		KDE_APPLICATIONS_AS_SCOPE: string;
		XDG_SEAT_PATH: string;
		MAIL: string;
		FORCE_COLOR: string;
		PATH: string;
		LOGNAME: string;
		XDG_MENU_PREFIX: string;
		XDG_CONFIG_DIRS: string;
		XAUTHORITY: string;
		LC_PAPER: string;
		LC_MEASUREMENT: string;
		XDG_SESSION_ID: string;
		npm_config_color: string;
		MOTD_SHOWN: string;
		XDG_SEAT: string;
		XDG_VTNR: string;
		XDG_SESSION_DESKTOP: string;
		DBUS_SESSION_BUS_ADDRESS: string;
		QT_LINUX_ACCESSIBILITY_ALWAYS_ON: string;
		LC_TELEPHONE: string;
		INVOCATION_ID: string;
		LC_ADDRESS: string;
		ICEAUTHORITY: string;
		XDG_DATA_DIRS: string;
		SHELL: string;
		DEBUG_COLORS: string;
		SESSION_MANAGER: string;
		GTK_MODULES: string;
		XDG_SESSION_CLASS: string;
		COLORTERM: string;
		LC_IDENTIFICATION: string;
		DISPLAY: string;
		HOME: string;
		MEMORY_PRESSURE_WATCH: string;
		XDG_CURRENT_DESKTOP: string;
		DEBUGINFOD_URLS: string;
		KDE_SESSION_VERSION: string;
		LC_NAME: string;
		LC_NUMERIC: string;
		MOCHA_COLORS: string;
		GTK_RC_FILES: string;
		LANG: string;
		GTK3_MODULES: string;
		GTK2_RC_FILES: string;
		QT_WAYLAND_RECONNECT: string;
		MEMORY_PRESSURE_WRITE: string;
		SYSTEMD_EXEC_PID: string;
		XDG_RUNTIME_DIR: string;
		PAM_KWALLET5_LOGIN: string;
		OLDPWD: string;
		MANAGERPID: string;
		DESKTOP_SESSION: string;
		USER: string;
		KDE_FULL_SESSION: string;
		XDG_SESSION_TYPE: string;
		LC_MONETARY: string;
		PWD: string;
		JOURNAL_STREAM: string;
		NODE_ENV: string;
		[key: `PUBLIC_${string}`]: undefined;
		[key: `${string}`]: string | undefined;
	}
}

/**
 * Similar to [`$env/dynamic/private`](https://svelte.dev/docs/kit/$env-dynamic-private), but only includes variables that begin with [`config.kit.env.publicPrefix`](https://svelte.dev/docs/kit/configuration#env) (which defaults to `PUBLIC_`), and can therefore safely be exposed to client-side code.
 * 
 * Note that public dynamic environment variables must all be sent from the server to the client, causing larger network requests — when possible, use `$env/static/public` instead.
 * 
 * Dynamic environment variables cannot be used during prerendering.
 * 
 * ```ts
 * import { env } from '$env/dynamic/public';
 * console.log(env.PUBLIC_DEPLOYMENT_SPECIFIC_VARIABLE);
 * ```
 */
declare module '$env/dynamic/public' {
	export const env: {
		[key: `PUBLIC_${string}`]: string | undefined;
	}
}
