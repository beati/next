package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/beati/next/match"
)

func main() {
	var clientDir string
	var domain string
	var cert string
	var key string
	var turnSecret string
	var dev bool
	flag.StringVar(&clientDir, "client_dir", "client/dist/", "client files directory")
	flag.StringVar(&domain, "domain", "next.beati.io", "domain name")
	flag.StringVar(&cert, "cert", "cert.pem", "certificate")
	flag.StringVar(&key, "key", "key.pem", "private key")
	flag.StringVar(&turnSecret, "turnSecret", "", "turn secret key")
	flag.BoolVar(&dev, "dev", false, "development mode")
	flag.Parse()

	err := loadStaticFiles(clientDir)
	if err != nil {
		log.Fatal(err)
	}

	if dev {
		go reloadStaticFiles(clientDir)
	}

	http.Handle("/", addHeaders(http.HandlerFunc(serveIndex)))

	http.Handle("/static/", addHeaders(http.HandlerFunc(serveStaticContent)))

	matcher := match.NewMatcher(true, turnSecret)
	http.Handle("/match", matcher)

	go func() {
		server := http.Server{
			Addr:         ":2000",
			Handler:      http.RedirectHandler("https://"+domain, http.StatusMovedPermanently),
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		}
		err = server.ListenAndServe()
		log.Fatal(err)
	}()

	server := http.Server{
		Addr:         ":2001",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		TLSConfig: &tls.Config{
			MinVersion:               tls.VersionTLS12,
			PreferServerCipherSuites: true,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
				tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
			},
		},
	}
	err = server.ListenAndServeTLS(cert, key)
	log.Fatal(err)
}

func addHeaders(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Strict-Transport-Security", "max-age=315360000; includeSubDomains")
		w.Header().Set("Content-Security-Policy", "script-src 'self' 'unsafe-eval'")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		h.ServeHTTP(w, r)
	})
}
