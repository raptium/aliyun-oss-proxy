package oss_proxy

import (
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

const RangeUpperBound = "9223372036854775806"

var passThroughHeaders = []string{
	"Range",
}

type OSSProxy struct {
	upstreamSchema string
	upstreamHost   string
	signer         OSSRequestSigner
	client         *http.Client
	proxy          *httputil.ReverseProxy
}

func NewOSSProxy(
	upstreamSchema string,
	upstreamHost string,
	signer OSSRequestSigner,
) *OSSProxy {
	p := &OSSProxy{
		upstreamSchema: upstreamSchema,
		upstreamHost:   upstreamHost,
		signer:         signer,
		client:         http.DefaultClient,
		proxy:          nil,
	}
	p.proxy = &httputil.ReverseProxy{
		Rewrite:        p.rewriteRequest,
		ModifyResponse: ModifyResponse,
	}
	return p
}

func ModifyResponse(resp *http.Response) error {
	if resp.Request.Method == "POST" && resp.Request.URL.RawQuery == "delete" && resp.Header.Get("Content-Length") == "0" {
		// Delete Quiet Mode Patch
		// AWS S3 returns an XML document even if quiet mode is enabled
		// OSS returns an empty body
		resp.Header.Set("Content-Type", "application/xml")
		resp.Header.Del("Content-Length")
		resp.Body = io.NopCloser(strings.NewReader(`<?xml version="1.0" encoding="UTF-8"?><DeleteResult></DeleteResult>`))
		return nil
	}
	return nil
}

func (p *OSSProxy) rewriteRequest(req *httputil.ProxyRequest) {
	resource := p.signer.GetResource(req.In)
	upstreamUrl := &url.URL{
		Scheme:   p.upstreamSchema,
		Host:     resource.BucketName + "." + p.upstreamHost,
		Path:     "/" + resource.ObjectName,
		RawQuery: req.In.URL.RawQuery,
	}
	header, err := p.signer.GetSignedHeaders(
		req.In.Method,
		resource.String(),
		req.In.Header,
	)
	if err != nil {
		log.Printf("failed to sign: %v", err)
		return
	}

	// pass through headers
	for _, h := range passThroughHeaders {
		v := req.In.Header.Get(h)
		// fix range upper bound
		if h == "Range" && strings.HasSuffix(v, "-"+RangeUpperBound) {
			// OSS does not support range out of bound
			v = strings.TrimSuffix(v, RangeUpperBound)
		}
		if v != "" {
			header.Set(h, v)
		}
	}

	req.Out.URL = upstreamUrl
	req.Out.Host = ""
	req.Out.Header = header
}

func (p *OSSProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" || r.URL.Path == "/livez" {
		_, _ = w.Write([]byte("OK"))
		return
	}
	if strings.Count(r.URL.Path, "/") < 2 {
		w.WriteHeader(404)
		return
	}
	p.proxy.ServeHTTP(w, r)
}
