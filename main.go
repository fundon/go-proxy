package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"strings"

	"github.com/visionmedia/docopt"
)

var client *http.Client

type Static struct {
	Dir http.FileSystem
}

func (s *Static) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "HEAD" {
		return
	}
	file := r.URL.Path
	f, err := s.Dir.Open(file)
	if err != nil {
		return
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return
	}

	// try to serve index file
	if fi.IsDir() {
		// redirect if missing trailing slash
		if !strings.HasSuffix(r.URL.Path, "/") {
			http.Redirect(rw, r, r.URL.Path+"/", http.StatusFound)
			return
		}

		file = path.Join(file, "index.html")
		f, err = s.Dir.Open(file)
		if err != nil {
			return
		}
		defer f.Close()

		fi, err = f.Stat()
		if err != nil || fi.IsDir() {
			return
		}
	}

	http.ServeContent(rw, r, file, fi.ModTime(), f)
	log.Println("Static " + r.URL.String() + ", Method " + r.Method)
}

type Proxy struct {
	Scheme, Host string
}

func (p *Proxy) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	client = &http.Client{}
	r.RequestURI = ""
	r.URL.Scheme = p.Scheme
	r.URL.Host = p.Host

	for k, _ := range r.Form {
		log.Println(k)
	}

	rwp, err := client.Do(r)
	if err != nil {
		//log.Fatal(err)
		if rwp == nil {
			rw.WriteHeader(408)
			log.Println("Request time out.")
		} else {
			rw.Write([]byte(rwp.Status))
		}
		return
	}
	header := rw.Header()
	for k, vs := range rwp.Header {
		for _, v := range vs {
			header.Add(k, v)
		}
	}
	rw.WriteHeader(rwp.StatusCode)
	data, err := ioutil.ReadAll(rwp.Body)
	rw.Write(data)
	log.Println("API " + r.URL.String() + ", Method " + r.Method)

	dst := new(bytes.Buffer)
	if err == nil {
		json.Indent(dst, data, "", "  ")
		log.Printf("JSON Data\n%v\n", dst)
	} else {
		log.Println(err)
	}
}

const Version = "0.0.1"

const Usage = `
  Usage:
    go-proxy [options]
    go-proxy -h | --help
    go-proxy --version

  Options:
    -s <source>, --source <source>    static resources dir [default: .]
    -p <port>, --port <port>          local server port [default: 3000]
    -d <dest>, --dest <dest>          proxy server
    -h, --help                        output help information
    -v, --version                     output version
`

func main() {
	args, err := docopt.Parse(Usage, nil, true, Version, false)
	if err != nil {
		fmt.Println(err)
	}

	var sourceDir = args["--source"].(string)
	var port = args["--port"].(string)
	dir := http.Dir(sourceDir)
	static := &Static{dir}
	http.HandleFunc("/", static.ServeHTTP)

	var proxyAddr = args["--dest"]
	if proxyAddr != nil && len(proxyAddr.(string)) > 0 {
		proxy := &Proxy{"http", proxyAddr.(string)}
		http.HandleFunc("/api/", proxy.ServeHTTP)
	}
	err = http.ListenAndServe(":"+port, nil)
	log.Fatal(err)
}
