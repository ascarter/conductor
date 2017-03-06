package conductor_test

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/ascarter/conductor"
)

type Employee struct {
	Name string `json:"name"`
}

type EmployeeResource struct {
	mu        sync.RWMutex
	lastID    int
	Employees map[string]Employee `json:"employees"`
}

func NewEmployeeResource() *EmployeeResource {
	return &EmployeeResource{lastID: 0}
}

// GET /employees
func (er *EmployeeResource) Index(w http.ResponseWriter, r *http.Request) {
	// return list of employees
	if err := conductor.WriteJSON(w, er.Employees); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}

// POST /employees
func (er *EmployeeResource) Create(w http.ResponseWriter, r *http.Request) {
	// Read employee object
	var v Employee
	if err := conductor.ReadJSON(r, &v); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	er.mu.Lock()
	defer er.mu.Unlock()

	// Add employee
	id := er.lastID + 1
	er.Employees[strconv.Itoa(id)] = v
	er.lastID++

	// Return ID of new object
	fmt.Fprintf(w, "%d", id)
}

// GET /employees/:id
func (er *EmployeeResource) Show(w http.ResponseWriter, r *http.Request) {
	params, ok := conductor.RouteParamsFromContext(r.Context())
	if !ok {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	id := params["id"]

	er.mu.RLock()
	defer er.mu.RUnlock()

	// Find existing employee
	e, ok := er.Employees[id]
	if !ok {
		http.NotFound(w, r)
	}

	// Return employee
	if err := conductor.WriteJSON(w, e); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	return
}

// PUT /employees/:id
func (er *EmployeeResource) Update(w http.ResponseWriter, r *http.Request) {
	params, ok := conductor.RouteParamsFromContext(r.Context())
	if !ok {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	id := params["id"]

	er.mu.Lock()
	defer er.mu.Unlock()

	// Find existing employee
	if _, ok := er.Employees[id]; !ok {
		http.NotFound(w, r)
	}

	// Read updated employee object
	var v Employee
	if err := conductor.ReadJSON(r, &v); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	er.Employees[id] = v

	w.WriteHeader(http.StatusOK)
	return
}

// DELETE /employees/:id
func (er *EmployeeResource) Destroy(w http.ResponseWriter, r *http.Request) {
	params, ok := conductor.RouteParamsFromContext(r.Context())
	if !ok {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	id := params["id"]

	er.mu.Lock()
	defer er.mu.Unlock()

	// Find existing employee
	if _, ok := er.Employees[id]; !ok {
		http.NotFound(w, r)
	}

	delete(er.Employees, id)
	w.WriteHeader(http.StatusOK)
	return
}

func Example_ResourceHandler() {
	mux := conductor.NewRouter()

	// Define middleware in order
	mux.Use(conductor.RequestIDComponent())
	mux.Use(conductor.DefaultRequestLogComponent)

	// Add resource
	h := conductor.NewResourceHandler(`/posts`, NewEmployeeResource())
	mux.Handle("/posts", h)

	// Start server
	log.Fatal(http.ListenAndServe(":8080", mux))
}
