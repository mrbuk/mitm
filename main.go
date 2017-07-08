package main

import (
	"log"
	"os"
	"sync"

	"github.com/mrbuk/mitm/server"
	"github.com/mrbuk/ntt/shared"
)

func main() {

	log.SetOutput(os.Stdout)

	httpServerPort := shared.GetOsEnv("MITM_HTTP_SERVER_PORT", "10080")
	httpsServerPort := shared.GetOsEnv("MITM_HTTPS_SERVER_PORT", "10443")
	httpTargetScheme := shared.GetOsEnv("MITM_HTTP_TARGET_SCHEME", "http")
	httpsTargetScheme := shared.GetOsEnv("MITM_HTTPS_TARGET_SCHEME", "https")

	target := shared.GetOsEnv("MITM_TARGET", "localhost:443")
	certificate := shared.GetOsEnv("MITM_CERT", "./cert.pem")
	key := shared.GetOsEnv("MITM_KEY", "./key.pem")

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		httpServerAddr := ":" + httpServerPort
		log.Println("Starting HTTP server on ", httpServerAddr)
		p := server.ProxyServer{
			Scheme:         httpTargetScheme,
			ListenHostPort: httpServerAddr,
			TargetHostPort: target,
		}
		httpProxy := p.NewHttpServer()

		log.Fatal(httpProxy.ListenAndServe())
	}()

	go func() {
		defer wg.Done()

		// ssl
		httpsServerAddr := ":" + httpsServerPort
		log.Println("Starting HTTPS server on ", httpsServerAddr)
		p := server.ProxyServer{
			Scheme:         httpsTargetScheme,
			ListenHostPort: httpsServerAddr,
			TargetHostPort: target,
		}
		httpsProxy := p.NewHttpServer()

		log.Fatal(httpsProxy.ListenAndServeTLS(certificate, key))
	}()

	wg.Wait()
}
