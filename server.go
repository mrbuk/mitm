package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/mrbuk/mitm/shared"
)

var client *http.Client

func proxy(target, scheme string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		id := uuid.New()

		// request data from target
		reqStartTime := time.Now()

		u := r.URL
		u.Scheme = scheme
		u.Host = target

		req := &http.Request{
			Method:     r.Method,
			URL:        u,
			Proto:      r.Proto,
			ProtoMajor: r.ProtoMajor,
			ProtoMinor: r.ProtoMinor,
			Header:     r.Header,
			Body:       r.Body,
			Host:       u.Host,
		}

		log.Printf("[%s] %s -> %s %v", id, r.RemoteAddr, r.Method, r.URL)

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

func main() {

	log.SetOutput(os.Stdout)

	httpServerPort := shared.GetOsEnv("MITM_HTTP_SERVER_PORT", "10080")
	httpsServerPort := shared.GetOsEnv("MITM_HTTPS_SERVER_PORT", "10443")
	httpTargetScheme := shared.GetOsEnv("MITM_HTTP_TARGET_SCHEME", "http")
	httpsTargetScheme := shared.GetOsEnv("MITM_HTTPS_TARGET_SCHEME", "https")

	target := shared.GetOsEnv("MITM_TARGET", "localhost:443")
	certificate := shared.GetOsEnv("MITM_CERT", "./cert.pem")
	key := shared.GetOsEnv("MITM_KEY", "./key.pem")

	client = &http.Client{}

	go func() {
		httpServerAddr := ":" + httpServerPort
		log.Println("Starting HTTP server on ", httpServerAddr)

		httpMux := http.NewServeMux()
		httpMux.HandleFunc("/", proxy(target, httpTargetScheme))
		httpServer := &http.Server{
			Addr:           httpServerAddr,
			Handler:        httpMux,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}
		log.Fatal(httpServer.ListenAndServe())
	}()

	// ssl
	httpsServerAddr := ":" + httpsServerPort
	log.Println("Starting HTTPS server on ", httpsServerAddr)

	httpsMux := http.NewServeMux()
	httpsMux.HandleFunc("/", proxy(target, httpsTargetScheme))
	httpsServer := &http.Server{
		Addr:           httpsServerAddr,
		Handler:        httpsMux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Fatal(httpsServer.ListenAndServeTLS(certificate, key))
}
