package core_http_server

import (
	"net/http"

	core_http_middleware "gitlab.com/guv41k/it_equipment_manager/internal/core/transport/http/middleware"
)

type ApiVersion string

const ApiVersion1 ApiVersion = "v1"

type APIVersionRouter struct {
	apiVersion ApiVersion
	routes     []Route
	middleware []core_http_middleware.Middleware
}

func NewAPIVersionRouter(apiVersion ApiVersion, middleware ...core_http_middleware.Middleware) *APIVersionRouter {
	return &APIVersionRouter{
		apiVersion: apiVersion,
		middleware: middleware,
	}
}

func (r *APIVersionRouter) AddRoutes(routes ...Route) {
	r.routes = append(r.routes, routes...)
}

func (r *APIVersionRouter) Handlers() map[string]http.Handler {
	handlers := make(map[string]http.Handler, len(r.routes))
	for _, route := range r.routes {
		pattern := route.Method + " /api/" + string(r.apiVersion) + route.Path
		handlers[pattern] = core_http_middleware.ChainMiddleware(route.WithMiddleware(), r.middleware...)
	}
	return handlers
}

func (s *HTTPServer) RegisterAPIRouters(routers ...*APIVersionRouter) {
	for _, router := range routers {
		for pattern, handler := range router.Handlers() {
			s.mux.Handle(pattern, handler)
		}
	}
}

func (s *HTTPServer) RegisterRoutes(routes ...Route) {
	for _, route := range routes {
		s.mux.Handle(route.Method+" "+route.Path, route.WithMiddleware())
	}
}
