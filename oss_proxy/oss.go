package oss_proxy

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

// copied from https://github.com/aliyun/aliyun-oss-go-sdk/blob/master/oss/conn.go#L30
var signKeyList = []string{
	"acl", "uploads", "location", "cors",
	"logging", "website", "referer", "lifecycle",
	"delete", "append", "tagging", "objectMeta",
	"uploadId", "partNumber", "security-token",
	"position", "img", "style", "styleName",
	"replication", "replicationProgress",
	"replicationLocation", "cname", "bucketInfo",
	"comp", "qos", "live", "status", "vod",
	"startTime", "endTime", "symlink",
	"x-oss-process", "response-content-type", "x-oss-traffic-limit",
	"response-content-language", "response-expires",
	"response-cache-control", "response-content-disposition",
	"response-content-encoding", "udf", "udfName", "udfImage",
	"udfId", "udfImageDesc", "udfApplication", "comp",
	"udfApplicationLog", "restore", "callback", "callback-var", "qosInfo",
	"policy", "stat", "encryption", "versions", "versioning", "versionId", "requestPayment",
	"x-oss-request-payer", "sequential",
	"inventory", "inventoryId", "continuation-token", "asyncFetch",
	"worm", "wormId", "wormExtend", "withHashContext",
	"x-oss-enable-md5", "x-oss-enable-sha1", "x-oss-enable-sha256",
	"x-oss-hash-ctx", "x-oss-md5-ctx", "transferAcceleration",
	"regionList", "cloudboxes", "x-oss-ac-source-ip", "x-oss-ac-subnet-mask", "x-oss-ac-vpc-id", "x-oss-ac-forward-allow",
	"metaQuery", "resourceGroup", "rtc",
}

type OSSResource struct {
	BucketName  string
	ObjectName  string
	SubResource string
}

func (r OSSResource) String() string {
	ret := fmt.Sprintf("/%s/%s", r.BucketName, r.ObjectName)
	if r.SubResource != "" {
		ret = fmt.Sprintf("%s?%s", ret, r.SubResource)
	}
	return ret
}

type OSSRequestSigner interface {
	GetResource(req *http.Request) OSSResource
	GetSignedHeaders(method, resource string, header http.Header) (http.Header, error)
	GenerateSignature(method, resource string, header http.Header, dateOrExpires string) (string, error)
}

type DefaultOSSRequestSigner struct {
	accessKeyID     string
	secretAccessKey string
	upstreamSchema  string
	upstreamHost    string
}

func NewOSSRequestSigner(accessKeyID, secretAccessKey, upstreamSchema, upstreamHost string) (OSSRequestSigner, error) {
	return &DefaultOSSRequestSigner{
		accessKeyID:     accessKeyID,
		secretAccessKey: secretAccessKey,
		upstreamSchema:  upstreamSchema,
		upstreamHost:    upstreamHost,
	}, nil
}

func (s DefaultOSSRequestSigner) GetResource(req *http.Request) OSSResource {
	path := req.URL.Path
	bucket := strings.Split(path, "/")[1]
	objectName := path[len(bucket)+2:]

	subResource := ""
	kv := make(map[string]string)
	keys := make([]string, 0, 8)
	for k1, vl := range req.URL.Query() {
		for _, k2 := range signKeyList {
			if k1 == k2 {
				kv[k1] = vl[0]
				keys = append(keys, k1)
			}
		}
	}
	sort.Strings(keys)
	for _, k := range keys {
		if subResource != "" {
			subResource += "&"
		}
		subResource += k
		if kv[k] != "" {
			subResource += "="
			subResource += url.QueryEscape(kv[k])
		}
	}

	return OSSResource{
		BucketName:  bucket,
		ObjectName:  objectName,
		SubResource: subResource,
	}
}

func (s DefaultOSSRequestSigner) GetSignedHeaders(method, resource string, header http.Header) (http.Header, error) {
	date := time.Now().UTC().Format(http.TimeFormat)
	signature, err := s.GenerateSignature(method, resource, header, date)
	if err != nil {
		return nil, err
	}
	ret := http.Header{}
	ret.Set("Date", date)
	ret.Set("Authorization", fmt.Sprintf("OSS %s:%s", s.accessKeyID, signature))

	contentType := header.Get("Content-Type")
	if contentType != "" {
		ret.Set("Content-Type", contentType)
	}

	contentMD5 := header.Get("Content-MD5")
	if contentMD5 != "" {
		ret.Set("Content-MD5", contentMD5)
	}

	return ret, nil
}

func (s DefaultOSSRequestSigner) GenerateSignature(method, resource string, header http.Header, dateOrExpires string) (string, error) {
	sb := strings.Builder{}

	sb.WriteString(method)
	sb.WriteString("\n")

	sb.WriteString(header.Get("Content-MD5"))
	sb.WriteString("\n")

	sb.WriteString(header.Get("Content-Type"))
	sb.WriteString("\n")

	sb.WriteString(dateOrExpires)
	sb.WriteString("\n")

	// CanonicalizedOSSHeaders Empty
	sb.WriteString(resource)

	h := hmac.New(sha1.New, []byte(s.secretAccessKey))
	_, err := h.Write([]byte(sb.String()))
	if err != nil {
		return "", err
	}
	digest := h.Sum(nil)
	return base64.StdEncoding.EncodeToString(digest), nil
}
