package checkvist

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestChecklists_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/auth/login.json":
			json.NewEncoder(w).Encode(map[string]string{"token": "test-token"})
		case "/checklists.json":
			if r.URL.Query().Get("archived") != "" {
				t.Error("unexpected archived parameter in List")
			}
			w.Write(loadFixture(t, "testdata/checklists/list.json"))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClient("user@example.com", "api-key", WithBaseURL(server.URL))
	checklists, err := client.Checklists().List(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(checklists) != 2 {
		t.Fatalf("expected 2 checklists, got %d", len(checklists))
	}
	if checklists[0].ID != 1 {
		t.Errorf("expected ID 1, got %d", checklists[0].ID)
	}
	if checklists[0].Name != "My First Checklist" {
		t.Errorf("expected name 'My First Checklist', got %s", checklists[0].Name)
	}
	if checklists[1].TaskCount != 25 {
		t.Errorf("expected TaskCount 25, got %d", checklists[1].TaskCount)
	}
}

func TestChecklists_ListArchived(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/auth/login.json":
			json.NewEncoder(w).Encode(map[string]string{"token": "test-token"})
		case "/checklists.json":
			if r.URL.Query().Get("archived") != "true" {
				t.Error("expected archived=true parameter")
			}
			w.Write(loadFixture(t, "testdata/checklists/list_archived.json"))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClient("user@example.com", "api-key", WithBaseURL(server.URL))
	checklists, err := client.Checklists().ListWithOptions(context.Background(), ListOptions{Archived: true})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(checklists) != 1 {
		t.Fatalf("expected 1 checklist, got %d", len(checklists))
	}
	if !checklists[0].Archived {
		t.Error("expected checklist to be archived")
	}
}

func TestChecklists_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/auth/login.json":
			json.NewEncoder(w).Encode(map[string]string{"token": "test-token"})
		case "/checklists/1.json":
			w.Write(loadFixture(t, "testdata/checklists/single.json"))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClient("user@example.com", "api-key", WithBaseURL(server.URL))
	checklist, err := client.Checklists().Get(context.Background(), 1)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if checklist.ID != 1 {
		t.Errorf("expected ID 1, got %d", checklist.ID)
	}
	if checklist.Name != "My First Checklist" {
		t.Errorf("expected name 'My First Checklist', got %s", checklist.Name)
	}
}

func TestChecklists_Get_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/auth/login.json":
			json.NewEncoder(w).Encode(map[string]string{"token": "test-token"})
		case "/checklists/999.json":
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error": "Checklist not found"}`))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClient("user@example.com", "api-key", WithBaseURL(server.URL))
	_, err := client.Checklists().Get(context.Background(), 999)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestChecklists_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/auth/login.json":
			json.NewEncoder(w).Encode(map[string]string{"token": "test-token"})
		case "/checklists.json":
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}

			var req createChecklistRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("failed to decode request: %v", err)
			}
			if req.Name != "New Checklist" {
				t.Errorf("expected name 'New Checklist', got %s", req.Name)
			}

			response := Checklist{
				ID:        42,
				Name:      req.Name,
				Public:    false,
				Archived:  false,
				UpdatedAt: time.Now(),
			}
			json.NewEncoder(w).Encode(response)
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClient("user@example.com", "api-key", WithBaseURL(server.URL))
	checklist, err := client.Checklists().Create(context.Background(), "New Checklist")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if checklist.ID != 42 {
		t.Errorf("expected ID 42, got %d", checklist.ID)
	}
	if checklist.Name != "New Checklist" {
		t.Errorf("expected name 'New Checklist', got %s", checklist.Name)
	}
}

func TestChecklists_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/auth/login.json":
			json.NewEncoder(w).Encode(map[string]string{"token": "test-token"})
		case "/checklists/1.json":
			if r.Method != http.MethodPut {
				t.Errorf("expected PUT, got %s", r.Method)
			}

			var req updateChecklistRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("failed to decode request: %v", err)
			}
			if req.Name != "Updated Name" {
				t.Errorf("expected name 'Updated Name', got %s", req.Name)
			}

			response := Checklist{
				ID:        1,
				Name:      req.Name,
				UpdatedAt: time.Now(),
			}
			json.NewEncoder(w).Encode(response)
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClient("user@example.com", "api-key", WithBaseURL(server.URL))
	checklist, err := client.Checklists().Update(context.Background(), 1, "Updated Name")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if checklist.Name != "Updated Name" {
		t.Errorf("expected name 'Updated Name', got %s", checklist.Name)
	}
}

func TestChecklists_Delete(t *testing.T) {
	var deleteCalled bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/auth/login.json":
			json.NewEncoder(w).Encode(map[string]string{"token": "test-token"})
		case "/checklists/1.json":
			if r.Method != http.MethodDelete {
				t.Errorf("expected DELETE, got %s", r.Method)
			}
			deleteCalled = true
			w.WriteHeader(http.StatusOK)
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClient("user@example.com", "api-key", WithBaseURL(server.URL))
	err := client.Checklists().Delete(context.Background(), 1)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !deleteCalled {
		t.Error("expected DELETE to be called")
	}
}

func TestChecklists_Archive(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/auth/login.json":
			json.NewEncoder(w).Encode(map[string]string{"token": "test-token"})
		case "/checklists/1.json":
			if r.Method != http.MethodPut {
				t.Errorf("expected PUT, got %s", r.Method)
			}

			var req archiveRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("failed to decode request: %v", err)
			}
			if !req.Archived {
				t.Error("expected archived=true")
			}

			response := Checklist{
				ID:        1,
				Name:      "Archived Checklist",
				Archived:  true,
				UpdatedAt: time.Now(),
			}
			json.NewEncoder(w).Encode(response)
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClient("user@example.com", "api-key", WithBaseURL(server.URL))
	checklist, err := client.Checklists().Archive(context.Background(), 1)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !checklist.Archived {
		t.Error("expected checklist to be archived")
	}
}

func TestChecklists_Unarchive(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/auth/login.json":
			json.NewEncoder(w).Encode(map[string]string{"token": "test-token"})
		case "/checklists/1.json":
			if r.Method != http.MethodPut {
				t.Errorf("expected PUT, got %s", r.Method)
			}

			var req archiveRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("failed to decode request: %v", err)
			}
			if req.Archived {
				t.Error("expected archived=false")
			}

			response := Checklist{
				ID:        1,
				Name:      "Unarchived Checklist",
				Archived:  false,
				UpdatedAt: time.Now(),
			}
			json.NewEncoder(w).Encode(response)
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClient("user@example.com", "api-key", WithBaseURL(server.URL))
	checklist, err := client.Checklists().Unarchive(context.Background(), 1)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if checklist.Archived {
		t.Error("expected checklist to not be archived")
	}
}
