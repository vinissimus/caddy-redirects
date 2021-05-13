package redirecter

import (
	"errors"
	"net/http"

	"github.com/caddyserver/caddy/v2"
)

func init() {
	caddy.RegisterModule(RedirecterAdmin{})
}

type RedirecterAdmin struct {
}

func (RedirecterAdmin) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "admin.api.redirecter",
		New: func() caddy.Module { return new(RedirecterAdmin) },
	}
}

func (r RedirecterAdmin) Routes() []caddy.AdminRoute {
	return []caddy.AdminRoute{
		{
			Pattern: "/redirecter/reload",
			Handler: caddy.AdminHandlerFunc(r.reload),
		},
	}
}

func (RedirecterAdmin) reload(w http.ResponseWriter, r *http.Request) error {
	var err error
	if redirecter != nil {
		err = redirecter.Reload()
	} else {
		err = errors.New("Redirecter is nil")
	}
	if err != nil {
		return caddy.APIError{
			HTTPStatus: http.StatusInternalServerError,
			Err:        err,
		}
	}
	return nil
}
