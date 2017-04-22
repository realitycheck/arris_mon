// Arris Monitor Server 1.0
// 2017

package main

import (
	"flag"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"
)

const (
	UrlStatus           = "/status.html"
	UrlStaticPrefix     = "/static/"
	PathStatic          = "static/"
	PathDefaultTemplate = "templates/arris_mon.html"
)

var (
	sourceUrl = "http://192.168.100.1/cgi-bin/status_cgi"
	addr      = "127.0.0.1:81"
)

func init() {
	flag.StringVar(&addr, "addr", addr, "Server address")
	flag.StringVar(&sourceUrl, "src", sourceUrl, "Source url")

	flag.Parse()
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get(sourceUrl)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, err.Error())
		return
	}
	defer resp.Body.Close()
	io.Copy(w, resp.Body)
}

func handleDefault(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles(PathDefaultTemplate)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, err.Error())
		return
	}
	tmpl.Execute(w, nil)
}

func main() {
	handleStatic := http.StripPrefix(UrlStaticPrefix, http.FileServer(http.Dir(PathStatic)))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var handler http.HandlerFunc

		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		switch {
		case r.URL.Path == UrlStatus:
			handler = handleStatus
			break
		case strings.HasPrefix(r.URL.Path, UrlStaticPrefix):
			handler = handleStatic.ServeHTTP
			break
		default:
			handler = handleDefault
			break
		}
		handler(w, r)
	})

	log.Println("Listening...")
	log.Fatal(http.ListenAndServe(addr, nil))
}
