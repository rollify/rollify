package apiv1

import (
	"net/http"

	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
)

func (a apiv1) registerRoutes(prefix string) {
	// Register handlers.
	a.apiws.Route(a.wrapWSGet("/ping").
		To(a.pong()).
		Metadata(restfulspec.KeyOpenAPITags, []string{"check"}).
		Doc("ping pong check handler").
		Writes("").
		Returns(http.StatusOK, "OK", ""))

	// Register docs.
	// Important: Needs to be the last route registed, because it needs to know what were
	// the registered endpoints.
	a.restContainer.Add(restfulspec.NewOpenAPIService(restfulspec.Config{
		WebServices: a.restContainer.RegisteredWebServices(),
		APIPath:     prefix + "/apidocs.json",
		APIVersion:  "v1",
	}))
}

func (a *apiv1) wrapWSGet(route string) *restful.RouteBuilder {
	return a.wrapMiddleware(route, a.apiws.GET(route))
}

// wrapMiddleware wraps a routebuilder with filters/middlewares.
func (a *apiv1) wrapMiddleware(route string, rb *restful.RouteBuilder) *restful.RouteBuilder {
	// TODO(slok).
	return rb
}
