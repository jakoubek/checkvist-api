package checkvist

import (
	"fmt"
	"strings"
	"time"
)

// models.go contains data structures for Checkvist entities:
// Checklist, Task, Note, User, Tags, TaskStatus, and DueDate.

// TaskStatus represents the status of a task.
type TaskStatus int

const (
	// StatusOpen indicates the task is open/incomplete.
	StatusOpen TaskStatus = 0
	// StatusClosed indicates the task is completed.
	StatusClosed TaskStatus = 1
	// StatusInvalidated indicates the task has been invalidated.
	StatusInvalidated TaskStatus = 2
)

// String returns the string representation of the TaskStatus.
func (s TaskStatus) String() string {
	switch s {
	case StatusOpen:
		return "open"
	case StatusClosed:
		return "closed"
	case StatusInvalidated:
		return "invalidated"
	default:
		return fmt.Sprintf("unknown(%d)", s)
	}
}

// Tags represents a set of tags as a map for efficient lookup.
type Tags map[string]bool

// APITime wraps time.Time with custom JSON unmarshaling for Checkvist API format.
// The Checkvist API returns timestamps in format "2006/01/02 15:04:05 +0000"
// instead of the standard RFC3339 format that Go expects.
type APITime struct {
	time.Time
}

// UnmarshalJSON handles multiple date formats from the Checkvist API.
func (t *APITime) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	if s == "" || s == "null" {
		return nil
	}

	// Try formats in order of likelihood
	formats := []string{
		"2006/01/02 15:04:05 -0700", // Checkvist API format
		time.RFC3339,                // ISO8601 with timezone
		"2006-01-02T15:04:05Z",      // RFC3339 without offset
	}

	for _, format := range formats {
		if parsed, err := time.Parse(format, s); err == nil {
			t.Time = parsed
			return nil
		}
	}

	return fmt.Errorf("cannot parse %q as time", s)
}

// MarshalJSON outputs time in RFC3339 format.
func (t APITime) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return []byte("null"), nil
	}
	return []byte(`"` + t.Format(time.RFC3339) + `"`), nil
}

// NewAPITime creates an APITime from a time.Time value.
func NewAPITime(t time.Time) APITime {
	return APITime{Time: t}
}

// Checklist represents a Checkvist checklist.
type Checklist struct {
	// ID is the unique identifier of the checklist.
	ID int `json:"id"`
	// Name is the title of the checklist.
	Name string `json:"name"`
	// Public indicates whether the checklist is publicly accessible.
	Public bool `json:"public"`
	// Archived indicates whether the checklist is archived.
	Archived bool `json:"archived"`
	// ReadOnly indicates whether the checklist is read-only for the current user.
	ReadOnly bool `json:"read_only"`
	// TaskCount is the total number of tasks in the checklist.
	TaskCount int `json:"task_count"`
	// TaskCompleted is the number of completed tasks in the checklist.
	TaskCompleted int `json:"task_completed"`
	// Tags contains the parsed tags from TagsAsText.
	Tags Tags `json:"-"`
	// TagsAsText is the raw tags string from the API.
	TagsAsText string `json:"tags_as_text"`
	// UpdatedAt is the timestamp of the last update.
	UpdatedAt APITime `json:"updated_at"`
}

// Task represents a task within a Checkvist checklist.
type Task struct {
	// ID is the unique identifier of the task.
	ID int `json:"id"`
	// ChecklistID is the ID of the checklist this task belongs to.
	ChecklistID int `json:"checklist_id"`
	// ParentID is the ID of the parent task, or 0 if this is a root task.
	ParentID int `json:"parent_id"`
	// Content is the text content of the task.
	Content string `json:"content"`
	// Status is the current status of the task (open, closed, invalidated).
	Status TaskStatus `json:"status"`
	// Position is the position of the task within its siblings.
	Position int `json:"position"`
	// Priority is the priority level (1 = highest, 2 = high, 0 = normal).
	Priority int `json:"priority"`
	// Tags contains the parsed tags from TagsAsText.
	Tags Tags `json:"-"`
	// TagsAsText is the raw tags string from the API.
	TagsAsText string `json:"tags_as_text"`
	// DueDateRaw is the raw due date string from the API.
	DueDateRaw string `json:"due"`
	// DueDate is the parsed due date, if available in ISO format.
	DueDate *time.Time `json:"-"`
	// AssigneeIDs contains the IDs of users assigned to this task.
	AssigneeIDs []int `json:"assignee_ids"`
	// CommentsCount is the number of notes/comments on this task.
	CommentsCount int `json:"comments_count"`
	// UpdateLine contains brief update information.
	UpdateLine string `json:"update_line"`
	// UpdatedAt is the timestamp of the last update.
	UpdatedAt APITime `json:"updated_at"`
	// CreatedAt is the timestamp when the task was created.
	CreatedAt APITime `json:"created_at"`
	// Children contains nested child tasks (when fetched with tree structure).
	Children []*Task `json:"tasks,omitempty"`
	// Notes contains the comments/notes attached to this task.
	Notes []Note `json:"notes,omitempty"`
}

// Note represents a comment/note on a task.
type Note struct {
	// ID is the unique identifier of the note.
	ID int `json:"id"`
	// TaskID is the ID of the task this note belongs to.
	TaskID int `json:"task_id"`
	// Comment is the text content of the note.
	Comment string `json:"comment"`
	// UpdatedAt is the timestamp of the last update.
	UpdatedAt APITime `json:"updated_at"`
	// CreatedAt is the timestamp when the note was created.
	CreatedAt APITime `json:"created_at"`
}

// User represents a Checkvist user.
type User struct {
	// ID is the unique identifier of the user.
	ID int `json:"id"`
	// Username is the user's display name.
	Username string `json:"username"`
	// Email is the user's email address.
	Email string `json:"email"`
}

// DueDate represents a due date for task creation using Checkvist's smart syntax.
type DueDate struct {
	value string
}

// Common due date constants for the Checkvist API.
var (
	// DueToday sets the due date to today.
	DueToday = DueDate{value: "Today"}
	// DueTomorrow sets the due date to tomorrow.
	DueTomorrow = DueDate{value: "Tomorrow"}
)

// DueAt creates a DueDate from a Go time.Time value.
func DueAt(t time.Time) DueDate {
	return DueDate{value: t.Format("2006-01-02")}
}

// DueString creates a DueDate from a raw string.
// Use this for custom date formats (e.g., "2026-02-01", "friday", "next week").
func DueString(s string) DueDate {
	return DueDate{value: s}
}

// DueInDays creates a DueDate for n days from now.
func DueInDays(n int) DueDate {
	t := time.Now().AddDate(0, 0, n)
	return DueDate{value: t.Format("2006-01-02")}
}

// String returns the smart syntax string for the due date.
func (d DueDate) String() string {
	return d.value
}
