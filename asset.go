package main

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
)

type asset struct {
	raw     []byte
	gz      []byte
	deflate []byte
}

var staticFilesLock sync.RWMutex
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

			staticFilesLock.Lock()
			staticFiles[file.Name()] = a
			staticFilesLock.Unlock()
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
	staticFilesLock.RLock()
	asset, ok := staticFiles[file]
	staticFilesLock.RUnlock()
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

func reloadStaticFiles(dir string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	err = watcher.Add(dir)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case <-watcher.Events:
			loadStaticFiles(dir)
		case err := <-watcher.Errors:
			log.Fatal(err)
		}
	}
}
