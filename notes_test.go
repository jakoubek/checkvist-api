package checkvist

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNotes_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/auth/login.json":
			json.NewEncoder(w).Encode(map[string]string{"token": "test-token"})
		case "/checklists/1/tasks/101/comments.json":
			if r.Method != http.MethodGet {
				t.Errorf("expected GET, got %s", r.Method)
			}
			w.Write(loadFixture(t, "testdata/notes/list.json"))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClient("user@example.com", "api-key", WithBaseURL(server.URL))
	notes, err := client.Notes(1, 101).List(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(notes) != 2 {
		t.Fatalf("expected 2 notes, got %d", len(notes))
	}
	if notes[0].ID != 501 {
		t.Errorf("expected ID 501, got %d", notes[0].ID)
	}
	if notes[0].Comment != "First comment on task" {
		t.Errorf("expected comment 'First comment on task', got %s", notes[0].Comment)
	}
	if notes[0].TaskID != 101 {
		t.Errorf("expected TaskID 101, got %d", notes[0].TaskID)
	}
}

func TestNotes_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/auth/login.json":
			json.NewEncoder(w).Encode(map[string]string{"token": "test-token"})
		case "/checklists/1/tasks/101/comments.json":
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}

			var req createNoteRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("failed to decode request: %v", err)
			}
			if req.Comment != "New note content" {
				t.Errorf("expected comment 'New note content', got %s", req.Comment)
			}

			response := Note{
				ID:        600,
				TaskID:    101,
				Comment:   req.Comment,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			json.NewEncoder(w).Encode(response)
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClient("user@example.com", "api-key", WithBaseURL(server.URL))
	note, err := client.Notes(1, 101).Create(context.Background(), "New note content")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if note.ID != 600 {
		t.Errorf("expected ID 600, got %d", note.ID)
	}
	if note.Comment != "New note content" {
		t.Errorf("expected comment 'New note content', got %s", note.Comment)
	}
}

func TestNotes_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/auth/login.json":
			json.NewEncoder(w).Encode(map[string]string{"token": "test-token"})
		case "/checklists/1/tasks/101/comments/501.json":
			if r.Method != http.MethodPut {
				t.Errorf("expected PUT, got %s", r.Method)
			}

			var req updateNoteRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("failed to decode request: %v", err)
			}
			if req.Comment != "Updated comment" {
				t.Errorf("expected comment 'Updated comment', got %s", req.Comment)
			}

			response := Note{
				ID:        501,
				TaskID:    101,
				Comment:   req.Comment,
				UpdatedAt: time.Now(),
			}
			json.NewEncoder(w).Encode(response)
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClient("user@example.com", "api-key", WithBaseURL(server.URL))
	note, err := client.Notes(1, 101).Update(context.Background(), 501, "Updated comment")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if note.Comment != "Updated comment" {
		t.Errorf("expected comment 'Updated comment', got %s", note.Comment)
	}
}

func TestNotes_Delete(t *testing.T) {
	var deleteCalled bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/auth/login.json":
			json.NewEncoder(w).Encode(map[string]string{"token": "test-token"})
		case "/checklists/1/tasks/101/comments/501.json":
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
	err := client.Notes(1, 101).Delete(context.Background(), 501)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !deleteCalled {
		t.Error("expected DELETE to be called")
	}
}
