package graphql

import (
	"context"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/solo-io/gloo/pkg/log"
	"github.com/vektah/gqlgen/graphql"
	"github.com/vektah/gqlgen/handler"
)

type Router struct {
	routes *routerSwapper
}

func NewRouter() *Router {
	return &Router{
		routes: &routerSwapper{
			mux: mux.NewRouter(),
		},
	}
}

type Endpoint struct {
	// name of the schema this endpoint serves
	SchemaName string
	// Where the playground will be served
	RootPath string
	// Where the query path will be served
	QueryPath string
	// the executable schema to serve
	ExecSchema graphql.ExecutableSchema
}

func (s *Router) UpdateEndpoints(endpoints ...*Endpoint) {
	m := mux.NewRouter()
	for _, endpoint := range endpoints {
		m.Handle(endpoint.RootPath, handler.Playground(endpoint.SchemaName, endpoint.QueryPath))
		m.Handle(endpoint.QueryPath, handler.GraphQL(endpoint.ExecSchema,
			handler.ResolverMiddleware(func(ctx context.Context, next graphql.Resolver) (res interface{}, err error) {
				rc := graphql.GetResolverContext(ctx)
				log.Printf("%v: Entered", endpoint.SchemaName, rc.Object, rc.Field.Name)
				res, err = next(ctx)
				log.Printf("%v: Left", endpoint.SchemaName, rc.Object, rc.Field.Name, "=>", res, err)
				return res, err
			}),
		))
	}
	s.routes.swap(m)
}

func (s *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.routes.serveHTTP(w, r)
}

// allows changing the routes being served
type routerSwapper struct {
	mu  sync.Mutex
	mux *mux.Router
}

func (rs *routerSwapper) swap(newMux *mux.Router) {
	rs.mu.Lock()
	rs.mux = newMux
	rs.mu.Unlock()
}

func (rs *routerSwapper) serveHTTP(w http.ResponseWriter, r *http.Request) {
	rs.mu.Lock()
	root := rs.mux
	rs.mu.Unlock()
	root.ServeHTTP(w, r)
}
