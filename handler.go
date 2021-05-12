package redirecter

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

func init() {
	caddy.RegisterModule(Handler{})
}

// Handler is an example; put your own type here.
type Handler struct {
	Pgds
	domain     string
	redirecter *Redirecter
}

// CaddyModule returns the Caddy module information.
func (Handler) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handler.redirecter",
		New: func() caddy.Module { return new(Handler) },
	}
}

// Provision sets up m.
func (h *Handler) Provision(ctx caddy.Context) error {
	if h.Host == "" || h.Port == 0 || h.User == "" || h.Password == "" || h.DbName == "" {
		return fmt.Errorf("Some values are missing")
	}

	// TODO: get from context? get from caddyfile?
	h.domain = "vinissimus.com"

	h.redirecter = initRedirecter(h.Pgds, h.domain)
	go h.redirecter.Reload()
	return nil
}

func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var handler Handler
	err := handler.UnmarshalCaddyfile(h.Dispenser)
	return handler, err
}

// UnmarshalCaddyfile sets up the handler from Caddyfile tokens. Syntax:
//
//     redirecter {
//         host "127.0.0.1"
//         port 5432
//         user "patates"
//         password "bullides"
//         db_name "vinissimus"
//     }
//
func (h *Handler) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		args := d.RemainingArgs()
		if len(args) > 0 {
			return d.ArgErr()
		}

		for d.NextBlock(0) {
			switch d.Val() {
			case "host":
				args := d.RemainingArgs()
				if len(args) != 1 {
					return d.ArgErr()
				}
				h.Host = args[0]
			case "port":
				args := d.RemainingArgs()
				if len(args) != 1 {
					return d.ArgErr()
				}
				port, err := strconv.Atoi(args[0])
				if err != nil {
					return err
				}
				h.Port = port
			case "user":
				args := d.RemainingArgs()
				if len(args) != 1 {
					return d.ArgErr()
				}
				h.User = args[0]
			case "password":
				args := d.RemainingArgs()
				if len(args) != 1 {
					return d.ArgErr()
				}
				h.Password = args[0]
			case "db_name":
				args := d.RemainingArgs()
				if len(args) != 1 {
					return d.ArgErr()
				}
				h.DbName = args[0]
			default:
				return d.Errf("unrecognized subdirective %q", d.Val())
			}
		}
	}
	return nil
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	newPath, ok := h.redirecter.FindRedirect(r.URL.Path)
	if !ok {
		err := next.ServeHTTP(w, r)
		return err
	}

	return redirect(w, r, newPath)
}

func redirect(w http.ResponseWriter, r *http.Request, to string) error {
	for strings.HasPrefix(to, "//") {
		// prevent path-based open redirects
		to = strings.TrimPrefix(to, "/")
	}
	http.Redirect(w, r, to, http.StatusPermanentRedirect)
	return nil
}

var (
	_ caddy.Provisioner           = (*Handler)(nil)
	_ caddyhttp.MiddlewareHandler = (*Handler)(nil)
)
