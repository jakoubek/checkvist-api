package checkvist

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTasks_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/auth/login.json":
			json.NewEncoder(w).Encode(map[string]string{"token": "test-token"})
		case "/checklists/1/tasks.json":
			w.Write(loadFixture(t, "testdata/tasks/list.json"))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClient("user@example.com", "api-key", WithBaseURL(server.URL))
	tasks, err := client.Tasks(1).List(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}
	if tasks[0].ID != 101 {
		t.Errorf("expected ID 101, got %d", tasks[0].ID)
	}
	if tasks[0].Content != "First task" {
		t.Errorf("expected content 'First task', got %s", tasks[0].Content)
	}
	if tasks[1].Priority != 1 {
		t.Errorf("expected priority 1, got %d", tasks[1].Priority)
	}
}

func TestTasks_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/auth/login.json":
			json.NewEncoder(w).Encode(map[string]string{"token": "test-token"})
		case "/checklists/1/tasks/101.json":
			w.Write(loadFixture(t, "testdata/tasks/single.json"))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClient("user@example.com", "api-key", WithBaseURL(server.URL))
	task, err := client.Tasks(1).Get(context.Background(), 101)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.ID != 101 {
		t.Errorf("expected ID 101, got %d", task.ID)
	}
	if task.Content != "First task" {
		t.Errorf("expected content 'First task', got %s", task.Content)
	}
}

func TestTasks_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/auth/login.json":
			json.NewEncoder(w).Encode(map[string]string{"token": "test-token"})
		case "/checklists/1/tasks.json":
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}

			var req CreateTaskRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("failed to decode request: %v", err)
			}
			if req.Content != "New task" {
				t.Errorf("expected content 'New task', got %s", req.Content)
			}

			response := Task{
				ID:          200,
				ChecklistID: 1,
				Content:     req.Content,
				Status:      StatusOpen,
				CreatedAt:   NewAPITime(time.Now()),
				UpdatedAt:   NewAPITime(time.Now()),
			}
			json.NewEncoder(w).Encode(response)
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClient("user@example.com", "api-key", WithBaseURL(server.URL))
	task, err := client.Tasks(1).Create(context.Background(), NewTask("New task"))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.ID != 200 {
		t.Errorf("expected ID 200, got %d", task.ID)
	}
	if task.Content != "New task" {
		t.Errorf("expected content 'New task', got %s", task.Content)
	}
}

func TestTasks_Create_WithBuilder(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/auth/login.json":
			json.NewEncoder(w).Encode(map[string]string{"token": "test-token"})
		case "/checklists/1/tasks.json":
			var req CreateTaskRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("failed to decode request: %v", err)
			}

			if req.Content != "Task with options" {
				t.Errorf("expected content 'Task with options', got %s", req.Content)
			}
			if req.Priority != 1 {
				t.Errorf("expected priority 1, got %d", req.Priority)
			}
			if req.Due != "^tomorrow" {
				t.Errorf("expected due '^tomorrow', got %s", req.Due)
			}
			if req.Tags != "tag1, tag2" {
				t.Errorf("expected tags 'tag1, tag2', got %s", req.Tags)
			}
			if req.ParentID != 100 {
				t.Errorf("expected parent_id 100, got %d", req.ParentID)
			}

			response := Task{
				ID:          201,
				ChecklistID: 1,
				ParentID:    req.ParentID,
				Content:     req.Content,
				Priority:    req.Priority,
				DueDateRaw:  "2026-01-15",
				TagsAsText:  req.Tags,
			}
			json.NewEncoder(w).Encode(response)
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClient("user@example.com", "api-key", WithBaseURL(server.URL))
	builder := NewTask("Task with options").
		WithPriority(1).
		WithDueDate(DueTomorrow).
		WithTags("tag1", "tag2").
		WithParent(100)

	task, err := client.Tasks(1).Create(context.Background(), builder)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.ID != 201 {
		t.Errorf("expected ID 201, got %d", task.ID)
	}
	if task.ParentID != 100 {
		t.Errorf("expected ParentID 100, got %d", task.ParentID)
	}
}

func TestTasks_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/auth/login.json":
			json.NewEncoder(w).Encode(map[string]string{"token": "test-token"})
		case "/checklists/1/tasks/101.json":
			if r.Method != http.MethodPut {
				t.Errorf("expected PUT, got %s", r.Method)
			}

			response := Task{
				ID:          101,
				ChecklistID: 1,
				Content:     "Updated content",
				UpdatedAt:   NewAPITime(time.Now()),
			}
			json.NewEncoder(w).Encode(response)
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClient("user@example.com", "api-key", WithBaseURL(server.URL))
	content := "Updated content"
	task, err := client.Tasks(1).Update(context.Background(), 101, UpdateTaskRequest{
		Content: &content,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Content != "Updated content" {
		t.Errorf("expected content 'Updated content', got %s", task.Content)
	}
}

func TestTasks_Delete(t *testing.T) {
	var deleteCalled bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/auth/login.json":
			json.NewEncoder(w).Encode(map[string]string{"token": "test-token"})
		case "/checklists/1/tasks/101.json":
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
	err := client.Tasks(1).Delete(context.Background(), 101)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !deleteCalled {
		t.Error("expected DELETE to be called")
	}
}

func TestTasks_Close(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/auth/login.json":
			json.NewEncoder(w).Encode(map[string]string{"token": "test-token"})
		case "/checklists/1/tasks/101/close.json":
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			response := Task{
				ID:     101,
				Status: StatusClosed,
			}
			json.NewEncoder(w).Encode(response)
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClient("user@example.com", "api-key", WithBaseURL(server.URL))
	task, err := client.Tasks(1).Close(context.Background(), 101)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Status != StatusClosed {
		t.Errorf("expected status Closed, got %v", task.Status)
	}
}

func TestTasks_Reopen(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/auth/login.json":
			json.NewEncoder(w).Encode(map[string]string{"token": "test-token"})
		case "/checklists/1/tasks/101/reopen.json":
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			response := Task{
				ID:     101,
				Status: StatusOpen,
			}
			json.NewEncoder(w).Encode(response)
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClient("user@example.com", "api-key", WithBaseURL(server.URL))
	task, err := client.Tasks(1).Reopen(context.Background(), 101)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Status != StatusOpen {
		t.Errorf("expected status Open, got %v", task.Status)
	}
}

func TestTasks_Invalidate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/auth/login.json":
			json.NewEncoder(w).Encode(map[string]string{"token": "test-token"})
		case "/checklists/1/tasks/101/invalidate.json":
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			response := Task{
				ID:     101,
				Status: StatusInvalidated,
			}
			json.NewEncoder(w).Encode(response)
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClient("user@example.com", "api-key", WithBaseURL(server.URL))
	task, err := client.Tasks(1).Invalidate(context.Background(), 101)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Status != StatusInvalidated {
		t.Errorf("expected status Invalidated, got %v", task.Status)
	}
}

func TestDueDate_Parsing(t *testing.T) {
	tests := []struct {
		name     string
		dueRaw   string
		expected *time.Time
	}{
		{
			name:     "ISO date",
			dueRaw:   "2026-01-20",
			expected: timePtr(time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC)),
		},
		{
			name:     "empty string",
			dueRaw:   "",
			expected: nil,
		},
		{
			name:     "invalid format",
			dueRaw:   "tomorrow",
			expected: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			task := &Task{DueDateRaw: tc.dueRaw}
			parseDueDate(task)

			if tc.expected == nil {
				if task.DueDate != nil {
					t.Errorf("expected nil DueDate, got %v", task.DueDate)
				}
			} else {
				if task.DueDate == nil {
					t.Fatal("expected DueDate to be set")
				}
				if !task.DueDate.Equal(*tc.expected) {
					t.Errorf("expected %v, got %v", tc.expected, task.DueDate)
				}
			}
		})
	}
}

func TestTaskBuilder(t *testing.T) {
	builder := NewTask("Test content").
		WithParent(50).
		WithPosition(3).
		WithDueDate(DueNextWeek).
		WithPriority(2).
		WithTags("work", "urgent")

	req := builder.build()

	if req.Content != "Test content" {
		t.Errorf("expected content 'Test content', got %s", req.Content)
	}
	if req.ParentID != 50 {
		t.Errorf("expected ParentID 50, got %d", req.ParentID)
	}
	if req.Position != 3 {
		t.Errorf("expected Position 3, got %d", req.Position)
	}
	if req.Due != "^next week" {
		t.Errorf("expected Due '^next week', got %s", req.Due)
	}
	if req.Priority != 2 {
		t.Errorf("expected Priority 2, got %d", req.Priority)
	}
	if req.Tags != "work, urgent" {
		t.Errorf("expected Tags 'work, urgent', got %s", req.Tags)
	}
}

func timePtr(t time.Time) *time.Time {
	return &t
}
