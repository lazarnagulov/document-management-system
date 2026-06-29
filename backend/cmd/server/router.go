package main

import (
	"github.com/gin-gonic/gin"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "bitbucket.org/HyperEther/document-management-system/docs"
)

type Routable interface {
	RegisterRoutes(rg *gin.RouterGroup)
}

type Router struct {
	api []Routable
	env string
}

func NewRouter(env string) *Router {
	return &Router{env: env}
}

func (r *Router) RegisterRoutes(routers ...Routable) {
	r.api = append(r.api, routers...)
}

func (r *Router) Build() *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Recovery())

	if r.env != "production" {
		engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	api := engine.Group("/api/v1")
	for _, route := range r.api {
		route.RegisterRoutes(api)
	}

	return engine
}
