package redirecter

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

func next(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func TestParse(t *testing.T) {
	tests := []struct {
		caddyfile      string
		reqPath        string
		locationHeader string
	}{
		{`redirecter {
			domain "vinissimus.com"
			host "127.0.0.1"
			port 5432
			user "patates"
			password "bullides"
			db_name "vinissimus"
		}`, "/old-page-needs-redirect", "/new-page"},
		{`redirecter {
			domain "vinissimus.com"
			host "127.0.0.1"
			port 5432
			user "patates"
			password "bullides"
			db_name "vinissimus"
		}`, "/working-page", ""},
	}

	loader = func(r *Redirecter) (map[string]string, error) {
		newUrlMap := make(map[string]string)
		newUrlMap["/old-page-needs-redirect"] = "/new-page"
		return newUrlMap, nil
	}

	for i, test := range tests {
		h := httpcaddyfile.Helper{
			Dispenser: caddyfile.NewTestDispenser(test.caddyfile),
		}
		actual, err := parseCaddyfile(h)
		if err != nil {
			panic(err)
		}
		handler := actual.(Handler)
		errProv := handler.Provision(caddy.Context{})
		if errProv != nil {
			panic(errProv)
		}

		// Wait until goroutine ran
		for j := 1; j <= 10; j++ {
			if handler.redirecter.urlMap != nil {
				break
			}
			time.Sleep(1 * time.Second)
		}

		r := httptest.NewRequest("GET", test.reqPath, strings.NewReader(""))
		r.RemoteAddr = "1.2.3.4"
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, r, caddyhttp.HandlerFunc(next))

		headers := w.Header()
		location, ok := headers["Location"]
		expectHeader := len(test.locationHeader) > 0
		if expectHeader {
			if !ok {
				t.Errorf("Text %v: Expected redirect but Location header is missing", i)
			} else if test.locationHeader != location[0] {
				t.Errorf("Test %v: Expected %s got %s", i, test.locationHeader, location[0])
			}
		} else {
			if ok {
				t.Errorf("Test %v: Did not expect Location header but got %s", i, location[0])
			}
		}
	}
}
