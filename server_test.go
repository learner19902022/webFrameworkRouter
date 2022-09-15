package web

import (
	"net/http"
	"testing"
)

func BenchmarkFindRouter(b *testing.B) {
	testRoutes := []struct {
		method string
		path   string
	}{
		{
			method: http.MethodGet,
			path:   "/",
		},
		{
			method: http.MethodGet,
			path:   "/week4",
		},
		{
			method: http.MethodGet,
			path:   "/*",
		},
		{
			method: http.MethodGet,
			path:   "/user",
		},
		{
			method: http.MethodGet,
			path:   "/user/*",
		},
		{
			method: http.MethodGet,
			path:   "/user/alex/:age",
		},
		{
			method: http.MethodGet,
			path:   "/creeps",
		},
		{
			method: http.MethodGet,
			path:   "/items/:name",
		},
		{
			method: http.MethodGet,
			path:   "/items/:name/:cost(^d+$)",
		},
		{
			method: http.MethodGet,
			path:   "/items/gloves/100",
		},
		{
			method: http.MethodGet,
			path:   "/items/gloves/200",
		},
		{
			method: http.MethodGet,
			path:   "/items/shoes/:size",
		},
		{
			method: http.MethodPost,
			path:   "/",
		},
		{
			method: http.MethodPost,
			path:   "/items/shoe/:speedup(^d+$)",
		},
		{
			method: http.MethodPost,
			path:   "/items/glove/:warmup(^[tooCold]|[tooWarm]$)",
		},
		{
			method: http.MethodPost,
			path:   "/:order",
		},
		{
			method: http.MethodPost,
			path:   "/travelMethod/boat",
		},
		{
			method: http.MethodPost,
			path:   "/travelMethod/bike",
		},
		{
			method: http.MethodPost,
			path:   "/travelMethod/swim",
		},
		{
			method: http.MethodPost,
			path:   "/travelMethod/car",
		},
		{
			method: http.MethodPost,
			path:   "/travelMethod/*",
		},
		{
			method: http.MethodPost,
			path:   "/travelMethod/boat/*",
		},
		{
			method: http.MethodPost,
			path:   "/travelMethod/boat/model",
		},
		{
			method: http.MethodPost,
			path:   "/travelMethod/boat/model/:tax([Paid])",
		},
		{
			method: http.MethodPost,
			path:   "/travelMethod/boat/model/isPay",
		},
		{
			method: http.MethodPost,
			path:   "/travelMethod/boat/model/:tax([Paid])/queue/find/here/are/you/kidding/me/let/us/test/if/you/can/really/Arrive/here/just/to/check/the/deeper/tree/performance/do/you/agree",
		},
		{
			method: http.MethodPut,
			path:   "/",
		},
		{
			method: http.MethodPut,
			path:   "/:id(qqq)/boat",
		},
		{
			method: http.MethodPut,
			path:   "/travelMethod/boat",
		},
		{
			method: http.MethodDelete,
			path:   "/",
		},
		{
			method: http.MethodDelete,
			path:   "/:id(qqq)/boat",
		},
		{
			method: http.MethodDelete,
			path:   "/travelMethod/boat",
		},
	}
	mockHandler := func(ctx *Context) {}
	//add new router
	r := newRouter()
	for _, tr := range testRoutes {
		r.addRoute(tr.method, tr.path, mockHandler)
	}

	testCases := []struct {
		method string
		path   string
	}{
		{
			method: http.MethodGet,
			path:   "/",
		},
		{
			method: http.MethodGet,
			path:   "/week4",
		},
		{
			method: http.MethodGet,
			path:   "/week5",
		},
		{
			method: http.MethodGet,
			path:   "/*",
		},
		{
			method: http.MethodGet,
			path:   "/user",
		},
		{
			method: http.MethodGet,
			path:   "/user/*",
		},
		{
			method: http.MethodGet,
			path:   "/user/alex",
		},
		{
			method: http.MethodGet,
			path:   "/user/alex/15",
		},
		{
			method: http.MethodGet,
			path:   "/user/alex/abc",
		},
		{
			method: http.MethodGet,
			path:   "/creeps/attack",
		},
		{
			method: http.MethodGet,
			path:   "/items/keyboard",
		}, {
			method: http.MethodGet,
			path:   "/items/keyboard/4719",
		},
		{
			method: http.MethodGet,
			path:   "/items/gloves/43789",
		},
		{
			method: http.MethodGet,
			path:   "/items/gloves/notfound?",
		},
		{
			method: http.MethodGet,
			path:   "/items/gloves/100",
		},
		{
			method: http.MethodGet,
			path:   "/items/gloves/200",
		},
		{
			method: http.MethodGet,
			path:   "/items/shoes/38",
		},
		{
			method: http.MethodPost,
			path:   "/",
		},
		{
			method: http.MethodPost,
			path:   "/items/basketball/12",
		},
		{
			method: http.MethodPost,
			path:   "/items/glove/tooCold",
		},
		{
			method: http.MethodPost,
			path:   "/its/glove/tooCold",
		},
		{
			method: http.MethodPost,
			path:   "/items/glo/tooCold",
		},
		{
			method: http.MethodPost,
			path:   "/4127942198",
		},
		{
			method: http.MethodPost,
			path:   "/travelMethod/boat",
		},
		{
			method: http.MethodPost,
			path:   "/travelMethod/skydiving",
		},
		{
			method: http.MethodPost,
			path:   "/travelMethod/bike",
		},
		{
			method: http.MethodPost,
			path:   "/travelMethod/swim",
		},
		{
			method: http.MethodPost,
			path:   "/travelMethod/car",
		},
		{
			method: http.MethodPost,
			path:   "/travelMethod/*",
		},
		{
			method: http.MethodPost,
			path:   "/travelMethod/boat/*",
		},
		{
			method: http.MethodPost,
			path:   "/travelMethod/boat/model",
		},
		{
			method: http.MethodPost,
			path:   "/travelMethod/boat/model/unPaid",
		},
		{
			method: http.MethodPost,
			path:   "/travelMethod/boat/model/isPay",
		},
		{
			method: http.MethodPost,
			path:   "/travelMethod/boat/model/Paid/queue/find/here/are/you/kidding/me/let/us/test/if/you/can/really/Arrive/here/just/to/check/the/deeper/tree/performance/do/you/agree",
		},
		{
			method: http.MethodPost,
			path:   "/travelMethod/boat/model/Paid/queue/find/here/are/you/kidding/me/let/us/test/if/you/can/really/Arrive/here/just/to/check/the/deeper/tree/performance/do/you/ae",
		},
		{
			method: http.MethodPut,
			path:   "/",
		},
		{
			method: http.MethodPut,
			path:   "/TTqqq/boat",
		},
		{
			method: http.MethodPut,
			path:   "/travelMethod/boat",
		},
		{
			method: http.MethodPut,
			path:   "/travelMethod/boat",
		},
		{
			method: http.MethodDelete,
			path:   "/",
		},
		{
			method: http.MethodDelete,
			path:   "/4240/boat",
		},
		{
			method: http.MethodDelete,
			path:   "/travelMethod/boat",
		},
		{
			method: http.MethodDelete,
			path:   "/travelMethod/boat",
		},
	}

	for i := 0; i < b.N; i++ {
		for _, tc := range testCases {
			_, _ = r.findRoute(tc.method, tc.path)
		}
	}
}
