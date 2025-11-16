package handler

import (
	"MgApplication/api-server/route"

	"github.com/gin-gonic/gin"
)

type Handler interface {
	Routes() []route.Route
	Prefix() string
	Middlewares() []gin.HandlerFunc
	Name() string
}

type Base struct {
	prefix string
	name   string
	mws    []gin.HandlerFunc
}

func New(name string) *Base {
	return &Base{name: name}
}

func (b *Base) Name() string {
	return b.name
}

func (b *Base) Prefix() string {
	return b.prefix
}

func (b *Base) Middlewares() []gin.HandlerFunc {
	return b.mws
}

func (b *Base) Routes() []route.Route {
	panic("need to declare routes for controller: " + b.name)
}
func (b *Base) AddPrefix(p string) *Base {
	b.prefix = b.prefix + p
	return b
}

func (b *Base) SetPrefix(p string) *Base {
	b.prefix = p
	return b
}

func (b *Base) AddMiddleware(mw gin.HandlerFunc) *Base {
	b.mws = append(b.mws, mw)
	return b
}
