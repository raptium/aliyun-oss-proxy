const { OSSRequestSigner } = await import("./oss.js");

const accessKeyId = process.env.ACCESS_KEY_ID || "";
const secretAccessKey = process.env.SECRET_ACCESS_KEY || "";
const port = process.env.PORT || 3000;
const signMode = process.env.SIGN_MODE || "PROXY";

let upstreamSchema = "https";
let upstreamHost = "oss-cn-shanghai.aliyuncs.com";

// validate configs
if (!(accessKeyId && secretAccessKey)) {
	throw new Error("accessKeyId and secretAccessKey must be set");
}

if (process.env.UPSTREAM_ENDPOINT) {
	const parse = new URL(process.env.UPSTREAM_ENDPOINT);
	upstreamSchema = parse.protocol;
	upstreamHost = parse.host;
}

const signer = new OSSRequestSigner(
	accessKeyId,
	secretAccessKey,
	upstreamSchema,
	upstreamHost,
);

function accessLogMiddleware(
	handler: (request: Request) => Promise<Response>,
): (request: Request) => Promise<Response> {
	return async (request) => {
		const requestStart = Date.now();
		const response = await handler(request);
		const requestEnd = Date.now();
		const elpased = requestEnd - requestStart;
		console.log(
			`${request.method} ${request.url} ${response.status} ${elpased}ms`,
		);
		return response;
	};
}

async function router(request: Request): Promise<Response> {
	const url = new URL(request.url);
	const path = url.pathname;
	if (path === "/" || path === "/livez") {
		return await ok(request);
	} else {
		return await handleResource(request);
	}
}

async function ok(request: Request): Promise<Response> {
	return new Response("OK");
}

async function handleResource(request: Request): Promise<Response> {
	const url = new URL(request.url);
	const path = url.pathname;
	const bucket = path.split("/")[1];
	const objectName = path.substring(bucket.length + 2);
	const contentType = request.headers.get("content-type") || "";

	if (signMode === "REDIRECT") {
		const redirectUrl = signer.getPreSignedUrl(
			request.method,
			bucket,
			objectName,
			contentType,
		);
		return new Response(null, {
			status: 307,
			headers: {
				location: redirectUrl,
			},
		});
	} else {
		const upstreamUrl = new URL(
			`${upstreamSchema}://${bucket}.${upstreamHost}/${objectName}`,
		);
		upstreamUrl.search = url.search; // copy query string
		const headers = signer.getSignedHeaders(
			request.method,
			bucket,
			objectName,
			contentType,
		);
		return await fetch(upstreamUrl, {
			headers: {
				...headers,
				range: request.headers.get("range") || "",
			},
		});
	}
}

// start

console.log(`Upstream: ${upstreamSchema}://${upstreamHost}`);
console.log(`Listening on port ${port}`);

Bun.serve({
	port: port,
	fetch: accessLogMiddleware(router),
});
