const { createHmac } = await import("node:crypto");

export class OSSRequestSigner {
	private readonly accessKeyId: string;
	private readonly secretAccessKey: string;
	private readonly upstreamSchema: string;
	private readonly upstreamHost: string;

	constructor(
		accessKeyId: string,
		secretAccessKey: string,
		upstreamSchema: string,
		upstreamHost: string,
	) {
		this.accessKeyId = accessKeyId;
		this.secretAccessKey = secretAccessKey;
		this.upstreamSchema = upstreamSchema;
		this.upstreamHost = upstreamHost;
	}

	public getPreSignedUrl(
		method: string,
		bucket: string,
		objectName: string,
		contentType: string,
		expiresInSeconds: number = 300,
	): string {
		const date = new Date();
		const expires = Math.trunc(date.getTime() / 1000) + expiresInSeconds;
		const url = new URL(
			`${this.upstreamSchema}://${bucket}.${this.upstreamHost}/${objectName}`,
		);
		const signature = this.generateSignature(
			method,
			bucket,
			objectName,
			contentType,
			expires.toString(),
		);
		url.searchParams.set("OSSAccessKeyId", this.accessKeyId);
		url.searchParams.set("Expires", expires.toString());
		url.searchParams.set("Signature", signature);
		return url.toString();
	}

	public getSignedHeaders(
		method: string,
		bucket: string,
		objectName: string,
		contentType: string,
	): Record<string, string> {
		const date = new Date().toUTCString();
		const signature = this.generateSignature(
			method,
			bucket,
			objectName,
			contentType,
			date,
		);
		return {
			authorization: `OSS ${this.accessKeyId}:${signature}`,
			date: date,
			"content-type": contentType,
		};
	}

	public generateSignature(
		method: string,
		bucket: string,
		objectName: string,
		contentType: string,
		dateOrExpires: string,
	): string {
		let verb = method;
		if (verb === "HEAD") verb = "GET";

		const contentMd5 = "";
		const date = dateOrExpires;
		const canonicalizedOSSHeaders = "";
		const canonicalizedResource = `/${bucket}/${objectName}`;

		const stringToSign =
			verb +
			"\n" +
			contentMd5 +
			"\n" +
			contentType +
			"\n" +
			date +
			"\n" +
			canonicalizedOSSHeaders +
			canonicalizedResource;

		return createHmac("sha1", this.secretAccessKey)
			.update(stringToSign)
			.digest("base64");
	}
}
