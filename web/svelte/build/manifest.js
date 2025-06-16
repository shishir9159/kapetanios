export const manifest = (() => {
function __memo(fn) {
	let value;
	return () => value ??= (value = fn());
}

return {
	appDir: "_app",
	appPath: "_app",
	assets: new Set(["favicon.png"]),
	mimeTypes: {".png":"image/png"},
	_: {
		client: {start:"_app/immutable/entry/start.wgXJubd5.js",app:"_app/immutable/entry/app.Dba7PK_Q.js",imports:["_app/immutable/entry/start.wgXJubd5.js","_app/immutable/chunks/COfwV3k6.js","_app/immutable/chunks/DUIfCcmH.js","_app/immutable/chunks/DIeogL5L.js","_app/immutable/chunks/DV0nxC6P.js","_app/immutable/entry/app.Dba7PK_Q.js","_app/immutable/chunks/DUIfCcmH.js","_app/immutable/chunks/DIeogL5L.js","_app/immutable/chunks/DmpyFbEB.js","_app/immutable/chunks/B5Xqpwiy.js","_app/immutable/chunks/CWj6FrbW.js","_app/immutable/chunks/DV0nxC6P.js","_app/immutable/chunks/BsmY1r07.js"],stylesheets:[],fonts:[],uses_env_dynamic_public:false},
		nodes: [
			__memo(() => import('./server/nodes/0.js')),
			__memo(() => import('./server/nodes/1.js')),
			__memo(() => import('./server/nodes/2.js')),
			__memo(() => import('./server/nodes/3.js')),
			__memo(() => import('./server/nodes/4.js')),
			__memo(() => import('./server/nodes/5.js'))
		],
		routes: [
			{
				id: "/",
				pattern: /^\/$/,
				params: [],
				page: { layouts: [0,], errors: [1,], leaf: 3 },
				endpoint: null
			},
			{
				id: "/minor-upgrade",
				pattern: /^\/minor-upgrade\/?$/,
				params: [],
				page: { layouts: [0,2,], errors: [1,,], leaf: 4 },
				endpoint: null
			},
			{
				id: "/minor-upgrade/upgrading",
				pattern: /^\/minor-upgrade\/upgrading\/?$/,
				params: [],
				page: { layouts: [0,2,], errors: [1,,], leaf: 5 },
				endpoint: null
			}
		],
		prerendered_routes: new Set([]),
		matchers: async () => {
			
			return {  };
		},
		server_assets: {}
	}
}
})();

export const prerendered = new Set([]);
