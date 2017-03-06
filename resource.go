package conductor

import (
	"net/http"
	"path"
)

// A Resource implements handlers for REST routes.
//
// Resources map a specific HTTP methods to route patterns. Each method in the
// interface performs a specific operation on the resource. Each action generally
// corresponds to a CRUD operation typically in a database.
//
// For a `photos` resource:
//	Method     HTTP Method     Path            Used For
//	------     -----------     -----------     --------------------------
//	Index      GET             /photos         display list of all photos
//	Create     POST            /photos         create a new photo
//	Show       GET             /photos/:id     display specific photo
//	Update     PUT             /photos/:id     update a specific photo
//	Destroy    DELETE          /photos/:id     delete a specific photo
type Resource interface {
	Index(http.ResponseWriter, *http.Request)
	Create(http.ResponseWriter, *http.Request)
	Show(http.ResponseWriter, *http.Request)
	Update(http.ResponseWriter, *http.Request)
	Destroy(http.ResponseWriter, *http.Request)
}

// A ResourceHandler routes requests to a Resource. It is used for representing
// a REST endpoint.
type ResourceHandler struct {
	r   Resource
	p   string
	mux *RouteHandler
}

// NewResourceHandler returns a new ResourceHandler instance.
func NewResourceHandler(pathname string, resource Resource) *ResourceHandler {
	collectionPath := pathname
	itemPath := path.Join(collectionPath, ":id")

	h := NewRouteHandler()
	h.HandleRouteFunc(http.MethodGet, collectionPath, resource.Index)
	h.HandleRouteFunc(http.MethodPost, collectionPath, resource.Create)
	h.HandleRouteFunc(http.MethodGet, itemPath, resource.Show)
	h.HandleRouteFunc(http.MethodPut, itemPath, resource.Update)
	h.HandleRouteFunc(http.MethodDelete, itemPath, resource.Destroy)
	return &ResourceHandler{r: resource, p: pathname, mux: h}
}

// ServeHTTP dispatches request to an internal RouteHandler configured to
// map to a Resource.
func (h *ResourceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}
