package router

import "github.com/gin-gonic/gin"

type MiddleWare func(gin.HandlerFunc) gin.HandlerFunc

type Router struct {
	RelativePath string
	Method       string
	Handle       gin.HandlerFunc
	MiddleWares  []MiddleWare
}

type ModuleRoute struct {
	MiddleWares []MiddleWare
	Routers     []*Router
}

func NewRouter(relativePath, method string, handle gin.HandlerFunc, middlewares ...MiddleWare) *Router {
	return &Router{
		RelativePath: relativePath,
		Method:       method,
		Handle:       handle,
		MiddleWares:  middlewares,
	}
}

func BuildHandler(prefixMiddleWares []MiddleWare, moduleRoutes ...ModuleRoute) *gin.Engine {
	router := gin.Default()
	for _, module := range moduleRoutes {
		for _, route := range module.Routers {
			middlewares := make([]MiddleWare, 0, len(prefixMiddleWares)+len(module.MiddleWares)+len(route.MiddleWares))
			middlewares = append(middlewares, route.MiddleWares...)
			middlewares = append(middlewares, module.MiddleWares...)
			middlewares = append(middlewares, prefixMiddleWares...)

			var handle = route.Handle
			for _, middle := range middlewares {
				handle = middle(handle)
			}
			router.Handle(route.Method, route.RelativePath, handle)
		}
	}
	return router
}
