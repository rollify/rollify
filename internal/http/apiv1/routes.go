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
		Doc("ping pong check").
		Writes("").
		Returns(http.StatusOK, "OK", ""))

	a.apiws.Route(a.wrapWSGet("/dice/types").
		To(a.listDiceTypes()).
		Metadata(restfulspec.KeyOpenAPITags, []string{"dice"}).
		Doc("lists available dice types").
		Writes(listDiceTypesResponse{}).
		Returns(http.StatusOK, "OK", listDiceTypesResponse{}))

	a.apiws.Route(a.wrapWSPost("/dice/roll").
		To(a.createDiceRoll()).
		Metadata(restfulspec.KeyOpenAPITags, []string{"dice"}).
		Doc("creates a dice roll").
		Writes(createDiceRollResponse{}).
		Reads(createDiceRollRequest{}).
		Returns(http.StatusCreated, "Created", createDiceRollResponse{}).
		Returns(http.StatusBadRequest, "", nil))

	a.apiws.Route(a.wrapWSPost("/room").
		To(a.createRoom()).
		Metadata(restfulspec.KeyOpenAPITags, []string{"room"}).
		Doc("creates a room").
		Writes(createRoomResponse{}).
		Reads(createRoomRequest{}).
		Returns(http.StatusCreated, "Created", createRoomResponse{}).
		Returns(http.StatusBadRequest, "", nil).
		Returns(http.StatusConflict, "room already exists", nil))

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

func (a *apiv1) wrapWSPost(route string) *restful.RouteBuilder {
	return a.wrapMiddleware(route, a.apiws.POST(route))
}

// wrapMiddleware wraps a routebuilder with filters/middlewares.
func (a *apiv1) wrapMiddleware(route string, rb *restful.RouteBuilder) *restful.RouteBuilder {
	// TODO(slok).
	return rb
}
