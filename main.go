package main

import (
	"github.com/raptium/aliyun-oss-proxy/oss_proxy"
	"github.com/urfave/negroni"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

const DefaultPort = 3000
const DefaultUpstreamSchema = "https"
const DefaultUpstreamHost = "oss-cn-shanghai.aliyuncs.com"

func main() {
	var err error
	var accessKeyId = os.Getenv("ACCESS_KEY_ID")
	var secretAccessKey = os.Getenv("SECRET_ACCESS_KEY")
	if accessKeyId == "" || secretAccessKey == "" {
		panic("ACCESS_KEY_ID and SECRET_ACCESS_KEY must be set")
	}

	var port = DefaultPort
	var portStr = os.Getenv("PORT")
	if portStr != "" {
		port, err = strconv.Atoi(portStr)
		if err != nil {
			port = DefaultPort
		}
	}

	var upstreamEndpoint = os.Getenv("UPSTREAM_ENDPOINT")
	var upstreamSchema = DefaultUpstreamSchema
	var upstreamHost = DefaultUpstreamHost
	if upstreamEndpoint != "" {
		p, err := url.Parse(upstreamEndpoint)
		if err == nil {
			upstreamSchema = p.Scheme
			upstreamHost = p.Host
		}
	}

	log.Printf("Upstream schema: %s", upstreamSchema)
	log.Printf("Upstream host: %s", upstreamHost)
	log.Printf("Starting server on port %d", port)

	signer, err := oss_proxy.NewOSSRequestSigner(
		accessKeyId,
		secretAccessKey,
		upstreamSchema,
		upstreamHost,
	)
	if err != nil {
		log.Panicf("Error creating signer: %s", err.Error())
	}

	proxy := oss_proxy.NewOSSProxy(upstreamSchema, upstreamHost, signer)

	n := negroni.Classic()
	n.UseHandler(proxy)

	err = http.ListenAndServe(":"+strconv.Itoa(port), n)
	if err != nil {
		log.Panicf("Error starting server: %s", err.Error())
		return
	}
}
