package httpserver

import (
	"net/http"
	"github.com/gorilla/mux"
	"time"
)


type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

var apiRoutes = Routes{
	Route{
		"Index",
		"GET",
		"/",
		IpRuleList,
	},
	Route{
		"IpRuleCreate",
		"POST",
		"/iprule",
		IpRuleCreate,
	},
	Route{
		"IpRuleShow",
		"GET",
		"/iprule/{id}",
		IpRule,
	},
	Route{
		"IpRulesList",
		"GET",
		"/iprule",
		IpRuleList,
	},
}

var uiRoutes = Routes{
	Route{
		"Index",
		"GET",
		"/",
		IpRuleList,
	},
}

func NewRouter(routes Routes) *mux.Router {

	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		var handler http.Handler

		handler = route.HandlerFunc
		handler = CorsHandler(Logger(handler, route.Name))
		router.
		Methods(route.Method, "OPTIONS"). //options for cors preflight on all routes
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)

	}

	return router
}


func CorsHandler(inner http.Handler) http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//todo stricter rules
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "origin, content-type, accept")
		inner.ServeHTTP(w, r)
	})
}

func Logger(inner http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		inner.ServeHTTP(w, r)

		requestLogger.Printf(
			"%s\t%s\t%s\t%s",
			r.Method,
			r.RequestURI,
			name,
			time.Since(start),
		)
	})
}
