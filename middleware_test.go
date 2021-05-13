package redirecter

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

func next(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func TestUrlWithoutQuery(t *testing.T) {
	url := url.URL{}
	url.Host = "domain.cat"
	url.Scheme = "https"
	url.Path = "/path"
	url.RawQuery = "?a=1&b=2"
	got := urlWithoutQuery(url)
	expected := "https://domain.cat/path"
	if expected != got {
		t.Errorf("Expected %s got %s", expected, got)
	}
}

func TestRedirect(t *testing.T) {
	tests := []struct {
		caddyfile      string
		reqPath        string
		locationHeader string
	}{
		{`redirecter {
			host "127.0.0.1"
			port 5432
			user "patates"
			password "bullides"
			db_name "vinissimus"
		}`, "/old-page-needs-redirect", "/new-page"},
		{`redirecter {
			host "127.0.0.1"
			port 5432
			user "patates"
			password "bullides"
			db_name "vinissimus"
		}`, "/working-page", ""},
	}

	loader = func(r *Redirecter) (map[string]string, error) {
		newUrlMap := make(map[string]string)
		newUrlMap["https://sub.domain.cat/old-page-needs-redirect"] = "/new-page"
		return newUrlMap, nil
	}

	for i, test := range tests {
		redirecter = nil
		h := httpcaddyfile.Helper{
			Dispenser: caddyfile.NewTestDispenser(test.caddyfile),
		}
		actual, err := parseCaddyfile(h)
		if err != nil {
			panic(err)
		}
		handler := actual.(*Middleware)
		errProv := handler.Provision(caddy.Context{})
		if errProv != nil {
			panic(errProv)
		}

		r := httptest.NewRequest("GET", test.reqPath, strings.NewReader(""))
		r.URL.Host = "sub.domain.cat"
		r.URL.Scheme = "https"
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
