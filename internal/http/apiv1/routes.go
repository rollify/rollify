package apiv1

import (
	"net/http"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	gohttpmetrics "github.com/slok/go-http-metrics/middleware/gorestful"
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

	a.apiws.Route(a.wrapWSPost("/dice/rolls").
		To(a.createDiceRoll()).
		Metadata(restfulspec.KeyOpenAPITags, []string{"dice"}).
		Doc("creates a dice roll").
		Writes(createDiceRollResponse{}).
		Reads(createDiceRollRequest{}).
		Returns(http.StatusCreated, "Created", createDiceRollResponse{}).
		Returns(http.StatusBadRequest, "", nil))

	a.apiws.Route(a.wrapWSGet("/dice/rolls").
		To(a.listDiceRolls()).
		Metadata(restfulspec.KeyOpenAPITags, []string{"dice"}).
		Doc("lists dice rolls").
		Param(a.apiws.QueryParameter(listDiceRollsParamUserID, "identifier of the user").DataType("string")).
		Param(a.apiws.QueryParameter(listDiceRollsurlParamRoomID, "identifier of the room").DataType("string")).
		Param(a.apiws.QueryParameter(listDiceRollsPaginationCursor, "cursor for next page of dice rolls").DataType("string")).
		Param(a.apiws.QueryParameter(listDiceRollsPaginationOrder, "order of the cursor, 'desc' or 'asc'").DataType("string")).
		Writes(listDiceRollsResponse{}).
		Returns(http.StatusOK, "OK", listDiceRollsResponse{}).
		Returns(http.StatusBadRequest, "", nil))

	a.apiws.Route(a.wrapWSPost("/rooms").
		To(a.createRoom()).
		Metadata(restfulspec.KeyOpenAPITags, []string{"room"}).
		Doc("creates a room").
		Writes(createRoomResponse{}).
		Reads(createRoomRequest{}).
		Returns(http.StatusCreated, "Created", createRoomResponse{}).
		Returns(http.StatusBadRequest, "", nil).
		Returns(http.StatusConflict, "room already exists", nil))

	a.apiws.Route(a.wrapWSGet("/rooms/{id}").
		To(a.getRoom()).
		Metadata(restfulspec.KeyOpenAPITags, []string{"room"}).
		Doc("gets a room").
		Param(a.apiws.PathParameter("id", "identifier of the room").DataType("string")).
		Writes(createRoomResponse{}).
		Returns(http.StatusOK, "OK", getRoomResponse{}).
		Returns(http.StatusBadRequest, "", nil).
		Returns(http.StatusNotFound, "room does not exists", nil))

	a.apiws.Route(a.wrapWSPost("/users").
		To(a.createUser()).
		Metadata(restfulspec.KeyOpenAPITags, []string{"user"}).
		Doc("creates a user in a room").
		Writes(createUserResponse{}).
		Reads(createUserRequest{}).
		Returns(http.StatusCreated, "Created", createUserResponse{}).
		Returns(http.StatusBadRequest, "", nil).
		Returns(http.StatusConflict, "user already exists", nil))

	a.apiws.Route(a.wrapWSGet("/users").
		To(a.listUsers()).
		Metadata(restfulspec.KeyOpenAPITags, []string{"user"}).
		Doc("list room users").
		Param(a.apiws.QueryParameter(listUsersurlParamRoomID, "identifier of the room").DataType("string")).
		Writes(listUsersResponse{}).
		Returns(http.StatusOK, "OK", listUsersResponse{}).
		Returns(http.StatusBadRequest, "", nil))

	a.apiws.Route(a.wrapWSGet("/ws/rooms/{id}").
		To(a.wsRoomEvents()).
		Metadata(restfulspec.KeyOpenAPITags, []string{"websocket"}).
		Doc("websocket connection for room events").
		Param(a.apiws.PathParameter("id", "identifier of the room").DataType("string")))

	// Register docs.
	// Important: Needs to be the last route registered, because it needs to know what were
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
	rb = rb.Filter(gohttpmetrics.Handler(route, a.metricsMiddleware))
	// Next middlewares...
	return rb
}
