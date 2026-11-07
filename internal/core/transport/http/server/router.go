package core_http_server

import (
	"fmt"
	"net/http"
)

type ApiVersion string

const ApiVersion1 ApiVersion = "v1"

type APIVersionRouter struct {
	*http.ServeMux
	ApiVersion ApiVersion
}

func NewAPIVersionRouter(apiVersion ApiVersion) *APIVersionRouter {
	return &APIVersionRouter{
		ServeMux:   http.NewServeMux(),
		ApiVersion: apiVersion,
	}
}

func (r *APIVersionRouter) RegisterRoutes(routes ...Route) {
	for _, route := range routes {
		pattern := fmt.Sprintf("%s %s", route.Method, route.Path)
		r.Handle(pattern, route.Handler)
	}
}

func (s *HTTPServer) RegisterAPIRouters(routers ...*APIVersionRouter) {
	for _, router := range routers {
		prefix := "/api/" + string(router.ApiVersion)
		s.mux.Handle(prefix+"/", http.StripPrefix(prefix, router))
	}
}
