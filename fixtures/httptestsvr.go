package fixtures

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"time"

	"github.com/ungerik/go-dry"
)

// a copy from colly
var (
	serverIndexResponse = mustFixturesFileBytes("hellowee.html")
	robotsFile          = `
User-agent: *
Allow: /allowed
Disallow: /disallowed
Disallow: /allowed*q=
`
)

func newUnstartedTestServer() *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write(serverIndexResponse)
	})

	mux.HandleFunc("/hellowee", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write(serverIndexResponse)
	})

	mux.HandleFunc("/html", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
<title>Test Page</title>
</head>
<body>
<h1>Hello World</h1>
<p class="description">This is a test page</p>
<p class="description">This is a test paragraph</p>
</body>
</html>
		`))
	})

	mux.HandleFunc("/xml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<page>
	<title>Test Page</title>
	<paragraph type="description">This is a test page</paragraph>
	<paragraph type="description">This is a test paragraph</paragraph>
</page>
		`))
	})

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(r.FormValue("name")))
		}
	})

	mux.HandleFunc("/set_cookie", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Set-Cookie", "sessionid=sessionid001")
		w.Header().Add("Set-Cookie", "tz=Asia/Shanghai")
		w.Header().Add("Set-Cookie", "logged_in=yes")

		w.WriteHeader(200)
		w.Write([]byte("ok"))
	})

	mux.HandleFunc("/check_cookie", func(w http.ResponseWriter, r *http.Request) {
		ckArr := r.Cookies()
		if len(ckArr) == 0 {
			w.WriteHeader(200)
			w.Write([]byte("ok"))
		}

		arr := []string{}
		for _, ck := range ckArr {
			arr = append(arr, fmt.Sprintf("%s=%s", ck.Name, ck.Value))
		}

		w.WriteHeader(200)
		w.Write([]byte(strings.Join(arr, "; ")))
	})

	mux.HandleFunc("/500", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(500)
		w.Write([]byte("<p>error</p>"))
	})

	mux.HandleFunc("/user_agent", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(r.Header.Get("User-Agent")))
	})

	mux.HandleFunc("/host_header", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(r.Host))
	})

	mux.HandleFunc("/headers", func(w http.ResponseWriter, r *http.Request) {
		raw := `<!DOCTYPE html>
<html>
<head>
<title>Test Page</title>
</head>
<body>
%s
</body>
</html>
		`
		w.WriteHeader(200)
		arr := []string{"<ul>"}
		for key, hdr := range r.Header {
			arr = append(arr, fmt.Sprintf("<li>%s: %s</li>", key, strings.Join(hdr, "\n")))
		}
		arr = append(arr, "</ul>")

		raw = fmt.Sprintf(raw, strings.Join(arr, "\n"))

		w.Write([]byte(raw))
	})

	mux.HandleFunc("/base", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
<title>Test Page</title>
<base href="http://xy.com/" />
</head>
<body>
<a href="z">link</a>
</body>
</html>
		`))
	})

	mux.HandleFunc("/base_relative", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
<title>Test Page</title>
<base href="/foobar/" />
</head>
<body>
<a href="z">link</a>
</body>
</html>
		`))
	})

	mux.HandleFunc("/tabs_and_newlines", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
<title>Test Page</title>
<base href="/foo	bar/" />
</head>
<body>
<a href="x
y">link</a>
</body>
</html>
		`))
	})

	mux.HandleFunc("/foobar/xy", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
<title>Test Page</title>
</head>
<body>
<p>hello</p>
</body>
</html>
		`))
	})

	mux.HandleFunc("/large_binary", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		ww := bufio.NewWriter(w)
		defer ww.Flush()
		for {
			// have to check error to detect client aborting download
			if _, err := ww.Write([]byte{0x41}); err != nil {
				return
			}
		}
	})

	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)

		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		i := 0

		for {
			select {
			case <-r.Context().Done():
				return
			case t := <-ticker.C:
				fmt.Fprintf(w, "%s\n", t)
				if flusher, ok := w.(http.Flusher); ok {
					flusher.Flush()
				}
				i++
				if i == 10 {
					return
				}
			}
		}
	})

	return httptest.NewUnstartedServer(mux)
}

func NewTestServer() *httptest.Server {
	svr := newUnstartedTestServer()
	svr.Start()

	return svr
}

func NewTestServerWithPort(port string) *httptest.Server {
	lis, err := net.Listen("tcp", "127.0.0.1:"+port)
	if err != nil {
		log.Fatal(err)
	}

	tsSvr := newUnstartedTestServer()
	tsSvr.Listener.Close()
	tsSvr.Listener = lis
	tsSvr.Start()

	return tsSvr
}

// ServedFixtureFile serves fixtures/xxx as `file:///xxx`
func ServedFixtureFile(path string) string {
	f, err := filepath.Abs(filepath.FromSlash("fixtures/" + path))
	dry.PanicIfErr(err)

	return "file://" + f
}

func mustFixturesFileBytes(path string) []byte {
	f, err := filepath.Abs(filepath.FromSlash("fixtures/" + path))
	dry.PanicIfErr(err)

	bytes, err := dry.FileGetBytes(f)
	dry.PanicIfErr(err)

	return bytes
}
