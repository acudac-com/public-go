package rest_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/acudac-com/public-go/rest"
)

var client = rest.NewClient(http.DefaultClient, "http://localhost:8080")

type Resource struct {
	ID string `json:"id"`
}

func init() {
	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			panic(err)
		}
	}()
}

func Test_List(t *testing.T) {
	http.HandleFunc("GET /resources", func(w http.ResponseWriter, r *http.Request) {
		resources := []*Resource{
			{ID: "foo"},
			{ID: "bar"},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resources); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	resources := []*Resource{}
	if err := client.Get("/resources", &resources); err != nil {
		t.Fatal(err)
	}
	if len(resources) != 2 {
		t.Fatalf("expected 2 resources, got %d", len(resources))
	}
}

func Test_Get(t *testing.T) {
	http.HandleFunc("GET /resources/{name}", func(w http.ResponseWriter, r *http.Request) {
		resource := &Resource{ID: r.PathValue("name")}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resource); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	resource := &Resource{}
	if err := client.Get("/resources/foo", resource); err != nil {
		t.Fatal(err)
	}
	if resource.ID != "foo" {
		t.Fatalf("expected resource.Name to be foo, got %s", resource.ID)
	}
}

func Test_Post(t *testing.T) {
	http.HandleFunc("POST /resources", func(w http.ResponseWriter, r *http.Request) {
		resource := &Resource{}
		if err := json.NewDecoder(r.Body).Decode(resource); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resource); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	resource := &Resource{ID: "foo"}
	if err := client.Post("/resources", resource, resource); err != nil {
		t.Fatal(err)
	}
	if resource.ID != "foo" {
		t.Fatalf("expected resource.Name to be foo, got %s", resource.ID)
	}
}

func Test_Put(t *testing.T) {
	http.HandleFunc("PUT /resources/{name}", func(w http.ResponseWriter, r *http.Request) {
		resource := &Resource{ID: r.PathValue("name")}
		if err := json.NewDecoder(r.Body).Decode(resource); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resource); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	resource := &Resource{ID: "foo"}
	if err := client.Put("/resources/foo", resource, resource); err != nil {
		t.Fatal(err)
	}
	if resource.ID != "foo" {
		t.Fatalf("expected resource.Name to be foo, got %s", resource.ID)
	}
}

func Test_Delete(t *testing.T) {
	http.HandleFunc("DELETE /resources/{name}", func(w http.ResponseWriter, r *http.Request) {
		resource := &Resource{ID: r.PathValue("name")}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resource); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	resource := &Resource{ID: "foo"}
	if err := client.Delete("/resources/foo", resource); err != nil {
		t.Fatal(err)
	}
	if resource.ID != "foo" {
		t.Fatalf("expected resource.Name to be foo, got %s", resource.ID)
	}
}
