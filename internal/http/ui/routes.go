package ui

import (
	"fmt"
	"net/http"

	"github.com/slok/go-http-metrics/middleware/std"
)

const (
	urlParamRoomID      = "roomID"
	queryParamSSEStream = "stream"
	queryParamCursor    = "cursor"

	uuidRegex = "[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}"
)

func (u ui) registerRoutes() {
	u.router.Handle("/static/*", http.StripPrefix(u.servePrefix+"/static", http.FileServer(http.FS(u.staticFS))))

	u.wrapGet("/", u.index())
	u.wrapPost("/create-room", u.createRoom())
	u.wrapGet(fmt.Sprintf("/login/{%s:%s}", urlParamRoomID, uuidRegex), u.login())
	u.wrapPost(fmt.Sprintf("/login/{%s:%s}/manage-user", urlParamRoomID, uuidRegex), u.manageUser())
	u.wrapGet(fmt.Sprintf("/room/{%s:%s}", urlParamRoomID, uuidRegex), u.room())
	u.wrapPost(fmt.Sprintf("/room/{%s:%s}/new-dice-roll", urlParamRoomID, uuidRegex), u.newDiceRoll())
	u.wrapGet(fmt.Sprintf("/room/{%s:%s}/dice-roll-history", urlParamRoomID, uuidRegex), u.diceRollHistory())
	u.wrapGet(fmt.Sprintf("/room/{%s:%s}/dice-roll-history/more-items", urlParamRoomID, uuidRegex), u.diceRollHistoryMoreItems())
	u.router.Mount("/subscribe/room/dice-roll-history", u.diceRollHistorySubscribe())
}

func (u ui) wrapGet(pattern string, h http.HandlerFunc) {
	u.router.With(
		// Add endpoint middlewares.
		std.HandlerProvider(pattern, u.metricsMiddleware),
	).Get(pattern, h)
}

func (u ui) wrapPost(pattern string, h http.HandlerFunc) {
	u.router.With(
		// Add endpoint middlewares.
		std.HandlerProvider(pattern, u.metricsMiddleware),
	).Post(pattern, h)
}
