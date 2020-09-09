package main

import (
	"net"
	"net/http"
	"net/http/cookiejar"
	"regexp"
	"time"
)

type JobSpider struct {
}

const (
	JobSourceCjol  = "cjol"
	JobSource51job = "51job"
	JobSourceLagou = "lagou"
	JobSourceBoss  = "boss"
)

var digitsRegexp = regexp.MustCompile(`(\d+)`)

var client = &http.Client{}

func init() {
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}
	client.Jar = jar
	// http.DefaultTransport =>
	client.Transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}
