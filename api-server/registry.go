package router

import (
	"net/url"
	"strings"
	"sync"

	log "MgApplication/api-log"
	"MgApplication/api-server/handler"
	"MgApplication/api-server/route"
	"MgApplication/api-server/swagger"
	"MgApplication/api-server/util/slc"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

type registry struct {
	ct     any
	base   string
	name   string
	mws    []gin.HandlerFunc
	routes []route.Route

	// Cached metas to avoid duplicate computation in both RegisterRoutes and SwaggerDefs
	cachedMetas []route.Meta
	metasOnce   sync.Once
}

func ParseControllers(cts ...handler.Handler) []*registry {
	return slc.Map(cts, newRegistry)
	// return getSwaggerDefs(registry)
}

// ParseGroupedControllers is an Fx-aware variant that collects all provided
// handler.Handler implementations registered under the group "servercontrollers".
// This allows modules to expose handlers via grouped results (fx.ResultTags)
// and have them aggregated into registries for route registration.
func ParseGroupedControllers(p struct {
	fx.In
	Controllers []handler.Handler `group:"servercontrollers"`
}) []*registry {
	// Returning nil instead of empty slice keeps existing behavior when no controllers.
	if len(p.Controllers) == 0 {
		return nil
	}
	return slc.Map(p.Controllers, newRegistry)
}

func newRegistry(ctr handler.Handler) *registry {
	return &registry{
		ct:     ctr,
		base:   ctr.Prefix(),
		name:   ctr.Name(),
		mws:    ctr.Middlewares(),
		routes: ctr.Routes(),
	}
}

func (r *registry) parsePath(path string) string {
	joined, err := url.JoinPath(r.base, path)
	if err != nil {
		// Log the error and use a fallback path construction
		// This should rarely happen as url.JoinPath is quite tolerant
		log.Error(nil, "Failed to join paths '%s' + '%s': %v - using fallback", r.base, path, err)
		// Simple fallback: ensure single slash between base and path
		if r.base == "" {
			return path
		}
		if path == "" {
			return r.base
		}
		// Remove trailing slash from base and leading slash from path, then join
		base := strings.TrimRight(r.base, "/")
		path = strings.TrimLeft(path, "/")
		return base + "/" + path
	}
	return joined
}

func (r *registry) toMeta(h route.Route) route.Meta {
	m := h.Meta()
	if m.Name == "" {
		m.Name = r.parsePath(m.Path)
	}
	m.Path = r.parsePath(m.Path)
	return m
}

// getMetas returns cached metas, computing them only once
func (r *registry) getMetas() []route.Meta {
	r.metasOnce.Do(func() {
		r.cachedMetas = slc.Map(r.routes, r.toMeta)
	})
	return r.cachedMetas
}

func GetSwaggerDefs(rs []*registry) []swagger.EndpointDef {
	return slc.FlatMap(rs, func(r *registry) []swagger.EndpointDef {
		return r.SwaggerDefs()
	})
}

func (r *registry) SwaggerDefs() []swagger.EndpointDef {
	// Use cached metas instead of recomputing
	metas := r.getMetas()
	d := slc.Map(metas, r.toSwagDefinition)

	return d
}

func (r *registry) toSwagDefinition(m route.Meta) swagger.EndpointDef {
	return swagger.EndpointDef{
		RequestType:  m.Req,
		ResponseType: m.Res,
		Group:        r.name,
		Name:         m.Name,
		Endpoint:     m.Path,
		Method:       m.Method,
	}
}
