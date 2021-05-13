package redirecter

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(Middleware{})
	httpcaddyfile.RegisterHandlerDirective("redirecter", parseCaddyfile)
}

var mutex = &sync.Mutex{}
var redirecter *Redirecter

type Middleware struct {
	Pgds
	logger *zap.Logger
}

func (Middleware) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.redirecter",
		New: func() caddy.Module { return new(Middleware) },
	}
}

func (h *Middleware) Provision(ctx caddy.Context) error {
	h.logger = ctx.Logger(h)

	if h.Host == "" || h.Port == 0 || h.User == "" || h.Password == "" || h.DbName == "" {
		return fmt.Errorf("Some values are missing")
	}

	mutex.Lock()
	defer mutex.Unlock()
	if redirecter == nil {
		h.logger.Info("Initializing redirecter singleton")
		redirecter = initRedirecter(h.Pgds, h.logger)
		err := redirecter.Reload()
		if err != nil {
			h.logger.Error("Failed to load redirecter", zap.Error(err))
		}
		return err
	}
	return nil
}

func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	handler := &Middleware{}
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
func (h *Middleware) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
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

func (h *Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	newPath, ok := redirecter.FindRedirect(buildUrlWithoutQuery(r))
	if ok {
		http.Redirect(w, r, newPath, http.StatusPermanentRedirect)
		return nil
	} else {
		return next.ServeHTTP(w, r)
	}
}

func buildUrlWithoutQuery(r *http.Request) string {
	newUrl := *r.URL
	newUrl.Host = r.Host
	newUrl.Scheme = "http"
	if r.TLS != nil {
		newUrl.Scheme = "https"
	}
	newUrl.RawQuery = ""
	return newUrl.String()
}

func getStringVar(ctx *caddy.Context, name string) string {
	switch vv := caddyhttp.GetVar(ctx, name); vv.(type) {
	case string:
		return (vv).(string)
	default:
		return ""
	}
}

var (
	_ caddy.Provisioner           = (*Middleware)(nil)
	_ caddyfile.Unmarshaler       = (*Middleware)(nil)
	_ caddyhttp.MiddlewareHandler = (*Middleware)(nil)
)
