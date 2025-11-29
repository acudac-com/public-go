package rest_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/acudac-com/public-go/rest"
)

type Resource struct {
	ID string `json:"id"`
}

func testHandler(path string, handler func(w http.ResponseWriter, r *http.Request)) (*httptest.Server, *rest.Client) {
	m := http.NewServeMux()
	m.HandleFunc(path, handler)
	srv := httptest.NewServer(m)
	return srv, rest.NewClient(http.DefaultClient, srv.URL)
}

func Test_List(t *testing.T) {
	srv, client := testHandler("GET /resources", func(w http.ResponseWriter, r *http.Request) {
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
	defer srv.Close()
	resources := []*Resource{}
	if err := client.Get("/resources", &resources); err != nil {
		t.Fatal(err)
	}
	if len(resources) != 2 {
		t.Fatalf("expected 2 resources, got %d", len(resources))
	}
}

func Test_Get(t *testing.T) {
	srv, client := testHandler("GET /resources/{name}", func(w http.ResponseWriter, r *http.Request) {
		resource := &Resource{ID: r.PathValue("name")}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resource); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	defer srv.Close()
	resource := &Resource{}
	if err := client.Get("/resources/foo", resource); err != nil {
		t.Fatal(err)
	}
	if resource.ID != "foo" {
		t.Fatalf("expected resource.Name to be foo, got %s", resource.ID)
	}
}

func Test_Post(t *testing.T) {
	srv, client := testHandler("POST /resources", func(w http.ResponseWriter, r *http.Request) {
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
	defer srv.Close()
	resource := &Resource{ID: "foo"}
	if err := client.Post("/resources", resource, resource); err != nil {
		t.Fatal(err)
	}
	if resource.ID != "foo" {
		t.Fatalf("expected resource.Name to be foo, got %s", resource.ID)
	}
}

func Test_Put(t *testing.T) {
	srv, client := testHandler("PUT /resources/{name}", func(w http.ResponseWriter, r *http.Request) {
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
	defer srv.Close()
	resource := &Resource{ID: "foo"}
	if err := client.Put("/resources/foo", resource, resource); err != nil {
		t.Fatal(err)
	}
	if resource.ID != "foo" {
		t.Fatalf("expected resource.Name to be foo, got %s", resource.ID)
	}
}

func Test_Delete(t *testing.T) {
	srv, client := testHandler("DELETE /resources/{name}", func(w http.ResponseWriter, r *http.Request) {
		resource := &Resource{ID: r.PathValue("name")}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resource); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	defer srv.Close()
	resource := &Resource{ID: "foo"}
	if err := client.Delete("/resources/foo", resource); err != nil {
		t.Fatal(err)
	}
	if resource.ID != "foo" {
		t.Fatalf("expected resource.Name to be foo, got %s", resource.ID)
	}
}
