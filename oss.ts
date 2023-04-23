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
		contentMd5: string | undefined | null = null,
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
			contentMd5 ?? "",
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
		contentMd5: string | undefined | null = null,
	): Record<string, string> {
		const date = new Date().toUTCString();
		const signature = this.generateSignature(
			method,
			bucket,
			objectName,
			contentType,
			contentMd5 ?? "",
			date,
		);
		const ret: Record<string, string> = {
			authorization: `OSS ${this.accessKeyId}:${signature}`,
			date: date,
			"content-type": contentType,
		};
		if (contentMd5) {
			ret["content-md5"] = contentMd5;
		}
		return ret;
	}

	public generateSignature(
		method: string,
		bucket: string,
		objectName: string,
		contentType: string,
		contentMd5: string,
		dateOrExpires: string,
	): string {
		const verb = method;
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
