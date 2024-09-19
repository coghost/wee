package fixtures

import (
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

	mux.HandleFunc("/activate", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
<title>Activate</title>
</head>
<body>
<a href="/activate/newtab" target="_blank">open in new tab</a>
<a href="/activate/self">open in same tab</a>
<a href="/activate/delay">open tab with delay</a>
</body>
</html>`))
	})

	mux.HandleFunc("/activate/newtab", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
<title>HelloNewTab</title>
</head>
<body>
<p>Hello New Tab</p>
</body>
</html>`))
	})

	mux.HandleFunc("/activate/self", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
<title>HelloSelf</title>
</head>
<body>
<p>Hello Self</p>
</body>
</html>`))
	})

	mux.HandleFunc("/activate/delay", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Second * 2)
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
<title>HelloSelf</title>
</head>
<body>
<p>Hello Self</p>
</body>
</html>`))
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

	// Add the new dynamic content endpoint
	mux.HandleFunc("/dynamic", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
            <html>
                <body>
                    <div id="initial-content">Initial Content</div>
                    <script>
                        setTimeout(function() {
                            var div = document.createElement('div');
                            div.id = 'dynamic-content';
                            div.textContent = 'Dynamic Content Loaded';
                            document.body.appendChild(div);
                        }, 1000);
                    </script>
                </body>
            </html>
        `))
	})

	// Add the new slow-load endpoint
	mux.HandleFunc("/slow-load", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // Simulate a slow load
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
            <html>
                <body>
                    <div id="loaded-content">Slow Loaded Content</div>
                </body>
            </html>
        `))
	})

	mux.HandleFunc("/click_test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `
        <button id="clickme">Click Me</button>
        <div id="clicked" style="display:none;">Clicked</div>
        <button id="delayed-button">Delayed Button</button>
        <div id="delayed-result" style="display:none;">Delayed Result</div>
        <button id="button1">Button 1</button>
        <button id="button2">Button 2</button>
        <button id="button3">Button 3</button>
        <div id="all-clicked" style="display:none;">All Clicked</div>
        <button id="offscreen-button" style="position:absolute;left:-9999px;">Offscreen</button>
        <input id="input-field" type="text">
        <div id="input-submitted" style="display:none;">Input Submitted</div>
        <button id="js-only-button">JS Only Button</button>
        <div id="js-clicked" style="display:none;">JS Clicked</div>
        <script>
            document.getElementById("clickme").addEventListener("click", function() {
                document.getElementById("clicked").style.display = "block";
            });
            document.getElementById("delayed-button").addEventListener("click", function() {
                setTimeout(() => {
                    document.getElementById("delayed-result").style.display = "block";
                }, 1000);
            });
            let clicked = 0;
            ["button1", "button2", "button3"].forEach(id => {
                document.getElementById(id).addEventListener("click", function() {
                    clicked++;
                    if (clicked === 3) {
                        document.getElementById("all-clicked").style.display = "block";
                    }
                });
            });
            document.getElementById("input-field").addEventListener("keypress", function(e) {
                if (e.key === "Enter") {
                    document.getElementById("input-submitted").style.display = "block";
                }
            });
            document.getElementById("js-only-button").addEventListener("click", function() {
                document.getElementById("js-clicked").style.display = "block";
            });
        </script>
    `)
	})

	mux.HandleFunc("/cookie_test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
				<button id="accept-cookies">Accept Cookies</button>
				<button id="accept-all">Accept All</button>
				<div id="cookies-accepted" style="display:none;">Cookies Accepted</div>
				<script>
					["accept-cookies", "accept-all"].forEach(id => {
						document.getElementById(id).addEventListener("click", function() {
							document.getElementById("cookies-accepted").style.display = "block";
						});
					});
				</script>
			`))
	})

	mux.HandleFunc("/popover_test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
	        <div class="popover" style="display: block;">
	            Popover 1
	            <button class="popover-close">Close</button>
	        </div>
	        <div class="popover" style="display: block;">
	            Popover 2
	            <button class="popover-close">Close</button>
	        </div>
	        <script>
	            document.querySelectorAll(".popover-close").forEach(button => {
	                button.addEventListener("click", function() {
	                    this.closest(".popover").style.display = "none";
	                });
	            });
	        </script>
	    `))
	})

	mux.HandleFunc("/sequential_click_test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
				<button id="button1">Button 1</button>
				<button id="button2">Button 2</button>
				<button id="button3">Button 3</button>
				<div id="all-clicked-in-order" style="display:none;">All Clicked In Order</div>
				<script>
					let clickOrder = [];
					["button1", "button2", "button3"].forEach(id => {
						document.getElementById(id).addEventListener("click", function() {
							clickOrder.push(id);
							if (clickOrder.join(",") === "button1,button2,button3") {
								document.getElementById("all-clicked-in-order").style.display = "block";
							}
						});
					});
				</script>
			`))
	})

	// New endpoints for hijack tests
	mux.HandleFunc("/hijack_test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		resource := r.URL.Query().Get("resource")
		fmt.Fprintf(w, `
        <html>
            <body>
                <h1>Hijack Test</h1>
                <script src="/test-script.js?resource=%s"></script>
                <img src="/test-image.jpg?resource=%s" />
                <script src="/test-script.js"></script>
            </body>
        </html>
        `, resource, resource)
	})

	mux.HandleFunc("/test-script.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		fmt.Fprint(w, "console.log('Test script loaded');")
	})

	mux.HandleFunc("/test-image.jpg", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		// Send a small dummy image
		w.Write([]byte{0xFF, 0xD8, 0xFF, 0xD9})
	})

	mux.HandleFunc("/image_test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `
        <html>
            <body>
                <h1>Image Test</h1>
                <img src="/test-image.jpg" id="test-image" />
                <script>
                    document.getElementById('test-image').onload = function() {
                        console.log('Image loaded');
                    }
                </script>
            </body>
        </html>
        `)
	})

	mux.HandleFunc("/xhr_test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `
        <html>
            <body>
                <h1>XHR Test</h1>
                <script>
                    fetch('/api/data')
                        .then(response => response.json())
                        .then(data => console.log(data));
                </script>
            </body>
        </html>
        `)
	})

	mux.HandleFunc("/api/data", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"message": "XHR response"}`)
	})

	mux.HandleFunc("/hijack_test_multiple", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `
        <html>
            <head>
                <link rel="stylesheet" href="/test.css" />
            </head>
            <body>
                <h1>Multiple Resources Test</h1>
                <script src="/test-script.js"></script>
            </body>
        </html>
        `)
	})

	mux.HandleFunc("/test.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		fmt.Fprint(w, "body { background-color: #f0f0f0; }")
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
