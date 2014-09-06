go-proxy
--------

A simple web APIs proxy server for rich web application.


## Usage

```sh
go-proxy -h
go-proxy -s dist/web -d apiserver.com
```

* `-d <dest>, --dest <dest>` API Server

* `-s <source>, --source <source>` Local Dir [default: .]

* `-p <port>, --port <port>` Local Server Port [default: 3000]


* __NOTE__

For front-end APIs Requests, using `/api/` prefix for URL.