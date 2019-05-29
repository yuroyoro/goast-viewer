package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Result struct {
	*Ast   `json:"ast"`
	Source string `json:"source"`
	Dump   string `json:"dump"`
}

func handleAsset(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[len("/"):]
	if path == "" {
		path = "index.html"
	}

	body, err := Asset("assets/" + path)
	if err != nil {
		http.Error(w, "Server Error", http.StatusInternalServerError)
		log.Printf("handleAsset: path %s :  %+v\n", path, err)
		return
	}

	http.ServeContent(w, r, path, now, bytes.NewReader(body))
}

func handleParse(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Header().Set("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var source, filename string

	if file, handler, err := r.FormFile("sourcefile"); err == nil {
		filename = handler.Filename
		body, err := ioutil.ReadAll(file)
		if err != nil {
			http.Error(w, "Server Error", http.StatusInternalServerError)
			log.Printf("handleParse : Failed to read file %+v\n", err)
			return
		}
		source = string(body)
	} else {
		source = r.FormValue("source")
	}

	if len(source) == 0 {
		http.Error(w, "Server Error", http.StatusBadRequest)
		return
	}

	source = strings.Replace(source, "\r", "", -1)

	// log.Printf("%#v : %s, %T %#v", file, handler.Filename, data, data)

	ast, dump, err := Parse(filename, source)
	if err != nil {
		http.Error(w, "Server Error", http.StatusInternalServerError)
		log.Printf("handleParse : Failed to convert Ast to json %+v\n", err)
		return
	}

	result := Result{Ast: ast, Source: source, Dump: dump}

	body, err := json.Marshal(result)
	if err != nil {
		http.Error(w, "Server Error", http.StatusInternalServerError)
		log.Printf("handleParse: Failed to marshal Ast %+v\n", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(body))
}

func NewHandlers() *http.ServeMux {
	mux := http.NewServeMux()

	// serve assets
	mux.HandleFunc("/", handleAsset)

	// handle post
	mux.HandleFunc("/parse.json", handleParse)
	return mux
}

var now time.Time

func main() {
	// Flags
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s\n", os.Args[0])
		flag.PrintDefaults()
	}
	addr := flag.String("addr", ":7650", "listen address")

	flag.Parse()

	now = time.Now()
	// listen and serve http
	log.Printf("Server listening on : %s", *addr)
	http.ListenAndServe(*addr, NewHandlers())
}
