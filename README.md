# aliyun-oss-proxy

A simple proxy server for Aliyun OSS.

This can be used as a sidecar container so that the application
does not need to implement the signature algorithm or use OSS SDK to interact with OSS.

Although OSS has an AWS S3 compatible API, it is not fully compatible. For example, OSS supports passing bucket in vhost-style only, in some cases the client application may not be able to
use vhost-style then AWS SDK will not work.
