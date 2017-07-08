package server_test

import (
	"net/http"
	"net/url"

	"github.com/mrbuk/mitm/server"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Mapping HTTP headers from Client request", func() {

	var p server.ProxyServer
	var originalRequest, proxyRequest *http.Request

	BeforeEach(func() {
		p = server.ProxyServer{
			Scheme:         "https",
			TargetHostPort: "example.com",
		}

		originalRequest, _ = http.NewRequest("GET", "http://proxy-host", nil)
		proxyRequest = p.CreateProxyRequest(originalRequest)
	})

	It("the 'Host' and 'Scheme' field was changed", func() {

		originalUrl := originalRequest.URL
		proxyUrl := proxyRequest.URL

		Expect(proxyUrl.Host).To(Equal("example.com"))
		Expect(proxyUrl.Scheme).To(Equal("https"))

		resetHostAndScheme(originalUrl)
		resetHostAndScheme(proxyUrl)

		Expect(proxyUrl).To(Equal(originalUrl))

		Expect(proxyRequest.Host).To(Equal("example.com"))

	})

	It("all other fields mapped 1:1", func() {

		// remove Host field from both as this is changed
		originalRequest.Host, proxyRequest.Host = "", ""

		resetHostAndScheme(originalRequest.URL)
		resetHostAndScheme(proxyRequest.URL)

		Expect(originalRequest).To(Equal(proxyRequest))
	})
})

func resetHostAndScheme(u *url.URL) {
	u.Host = ""
	u.Scheme = ""
}
