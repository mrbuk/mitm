package server

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/mrbuk/mitm/shared"
)

var client *http.Client

type ProxyServer struct {
	Scheme         string
	ListenHostPort string
	TargetHostPort string

	Server *http.Server
}

func (p *ProxyServer) CreateProxyRequest(r *http.Request) *http.Request {

	var u url.URL

	u = *r.URL
	u.Scheme = p.Scheme
	u.Host = p.TargetHostPort

	req := &http.Request{
		Method:     r.Method,
		URL:        &u,
		Proto:      r.Proto,
		ProtoMajor: r.ProtoMajor,
		ProtoMinor: r.ProtoMinor,
		Header:     r.Header,
		Body:       r.Body,
		Host:       u.Host,
	}

	return req
}

func (p *ProxyServer) handler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		id := uuid.New()

		// request data from target
		reqStartTime := time.Now()

		u := r.URL
		u.Scheme = p.Scheme
		u.Host = p.TargetHostPort

		req := p.CreateProxyRequest(r)
		log.Printf("[%s] %s -> %s %v", id, r.RemoteAddr, r.Method, r.URL)

		client := &http.Client{}

		resp, err := client.Do(req)
		if err != nil {
			log.Println("Could not reach", u, "; Error:", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		log.Printf("[%s] %s <- Returned %d and %v bytes", id, resp.Request.RemoteAddr, resp.Status, resp.ContentLength)

		for key, value := range resp.Header {
			w.Header()[key] = value
		}
		io.Copy(w, resp.Body)
		w.WriteHeader(resp.StatusCode)

		// at the moment that same ...
		retStopTime := time.Now()
		reqStopTime := time.Now()

		tr := shared.TimedRequest{
			Timestamp:           reqStartTime.Unix(),
			Source:              r.RemoteAddr,
			RequestURI:          r.RequestURI,
			Target:              u.Host,
			ProxyServerDuration: reqStopTime.Sub(reqStartTime).Seconds(),
			TotalDuration:       retStopTime.Sub(reqStartTime).Seconds(),
			Status:              resp.StatusCode,
		}

		b, _ := json.Marshal(tr)
		log.Println(string(b))
	}
}

func (p *ProxyServer) NewHttpServer() *http.Server {

	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/", p.handler())

	httpServer := &http.Server{
		Addr:           p.ListenHostPort,
		Handler:        httpMux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	return httpServer
}
