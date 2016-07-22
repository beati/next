package main

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"crypto/tls"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/beati/next/match"
)

func main() {
	var clientDir string
	var domain string
	var cert string
	var key string
	var turnSecret string
	flag.StringVar(&clientDir, "client_dir", "client/dist/", "client files directory")
	flag.StringVar(&domain, "domain", "next.beati.io", "domain name")
	flag.StringVar(&cert, "cert", "cert.pem", "certificate")
	flag.StringVar(&key, "key", "key.pem", "private key")
	flag.StringVar(&turnSecret, "turnSecret", "", "turn secret key")
	flag.Parse()

	err := loadStaticFiles(clientDir)
	if err != nil {
		log.Fatal(err)
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
			},
		},
	}
	err = server.ListenAndServeTLS(cert, key)
	log.Fatal(err)
}

type asset struct {
	raw     []byte
	gz      []byte
	deflate []byte
}

var staticFiles = make(map[string]asset)

func loadStaticFiles(dir string) error {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		ext := path.Ext(file.Name())
		load := ext == ".html"
		load = load || ext == ".js"
		load = load || ext == ".css"
		load = load && !file.IsDir()
		if load {
			a := asset{}
			a.raw, err = ioutil.ReadFile(dir + file.Name())
			if err != nil {
				return err
			}

			b := &bytes.Buffer{}
			gz, err := gzip.NewWriterLevel(b, gzip.BestCompression)
			if err != nil {
				return err
			}
			_, err = gz.Write(a.raw)
			if err != nil {
				return err
			}
			err = gz.Close()
			if err != nil {
				return err
			}
			a.gz = b.Bytes()

			b = &bytes.Buffer{}
			deflate, err := flate.NewWriter(b, flate.BestCompression)
			if err != nil {
				return err
			}
			_, err = deflate.Write(a.raw)
			if err != nil {
				return err
			}
			err = deflate.Close()
			if err != nil {
				return err
			}
			a.deflate = b.Bytes()

			staticFiles[file.Name()] = a
		}
	}

	return nil
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	serveAsset(w, r, "index.html", false)
}

func serveStaticContent(w http.ResponseWriter, r *http.Request) {
	file := strings.TrimPrefix(r.URL.Path, "/static/")
	serveAsset(w, r, file, true)
}

func serveAsset(w http.ResponseWriter, r *http.Request, file string, cache bool) {
	asset, ok := staticFiles[file]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if cache {
		w.Header().Set("Cache-Control", "public, max-age=315360000")
	}

	ext := path.Ext(file)
	switch ext {
	case ".html":
		w.Header().Set("Content-Type", "text/html")
	case ".css":
		w.Header().Set("Content-Type", "text/css")
	case ".js":
		w.Header().Set("Content-Type", "application/javascript")
	}

	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		w.Header().Set("Content-Encoding", "gzip")
		w.Write(asset.gz)
	} else if strings.Contains(r.Header.Get("Accept-Encoding"), "deflate") {
		w.Header().Set("Content-Encoding", "deflate")
		w.Write(asset.deflate)
	} else {
		w.Write(asset.raw)
	}
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
