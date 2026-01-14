package checkvist

import (
	"context"
	"fmt"
	"time"
)

// tasks.go contains the TaskService for CRUD operations on tasks within a checklist.

// TaskService provides operations on tasks within a specific checklist.
type TaskService struct {
	client      *Client
	checklistID int
}

// Tasks returns a TaskService for performing task operations on the specified checklist.
func (c *Client) Tasks(checklistID int) *TaskService {
	return &TaskService{
		client:      c,
		checklistID: checklistID,
	}
}

// List returns all tasks in the checklist.
func (s *TaskService) List(ctx context.Context) ([]Task, error) {
	path := fmt.Sprintf("/checklists/%d/tasks.json", s.checklistID)

	var tasks []Task
	if err := s.client.doGet(ctx, path, &tasks); err != nil {
		return nil, err
	}

	// Parse due dates
	for i := range tasks {
		parseDueDate(&tasks[i])
	}

	return tasks, nil
}

// Get returns a single task by ID, including its parent hierarchy.
func (s *TaskService) Get(ctx context.Context, taskID int) (*Task, error) {
	path := fmt.Sprintf("/checklists/%d/tasks/%d.json", s.checklistID, taskID)

	var task Task
	if err := s.client.doGet(ctx, path, &task); err != nil {
		return nil, err
	}

	parseDueDate(&task)
	return &task, nil
}

// CreateTaskRequest represents the request body for creating a task.
type CreateTaskRequest struct {
	Content  string `json:"content"`
	ParentID int    `json:"parent_id,omitempty"`
	Position int    `json:"position,omitempty"`
	Due      string `json:"due_date,omitempty"`
	Priority int    `json:"priority,omitempty"`
	Tags     string `json:"tags,omitempty"`
	Repeat   string `json:"repeat,omitempty"`
}

// createTaskWrapper wraps the task fields for the nested JSON format
// expected by the Checkvist API: {"task": {"content": "...", ...}}
type createTaskWrapper struct {
	Task CreateTaskRequest `json:"task"`
}

// TaskBuilder provides a fluent interface for building task creation requests.
type TaskBuilder struct {
	content  string
	parentID int
	position int
	due      string
	priority int
	tags     []string
	repeat   string
}

// NewTask creates a new TaskBuilder with the given content.
func NewTask(content string) *TaskBuilder {
	return &TaskBuilder{content: content}
}

// WithParent sets the parent task ID for creating a subtask.
func (b *TaskBuilder) WithParent(parentID int) *TaskBuilder {
	b.parentID = parentID
	return b
}

// WithPosition sets the position of the task within its siblings.
func (b *TaskBuilder) WithPosition(position int) *TaskBuilder {
	b.position = position
	return b
}

// WithDueDate sets the due date using a DueDate value.
func (b *TaskBuilder) WithDueDate(due DueDate) *TaskBuilder {
	b.due = due.String()
	return b
}

// WithPriority sets the priority level (1 = highest, 2 = high, 0 = normal).
func (b *TaskBuilder) WithPriority(priority int) *TaskBuilder {
	b.priority = priority
	return b
}

// WithTags sets the tags for the task.
func (b *TaskBuilder) WithTags(tags ...string) *TaskBuilder {
	b.tags = tags
	return b
}

// WithRepeat sets the repeat pattern for the task using Checkvist's smart syntax.
// Common patterns include:
//   - "daily" - repeats every day
//   - "weekly" - repeats every week
//   - "monthly" - repeats every month
//   - "yearly" - repeats every year
//   - "every 2 days" - repeats every 2 days
//   - "every week on monday" - repeats weekly on Monday
//   - "every month on 15" - repeats monthly on the 15th
//   - "every 2 weeks on friday" - repeats every 2 weeks on Friday
func (b *TaskBuilder) WithRepeat(pattern string) *TaskBuilder {
	b.repeat = pattern
	return b
}

// build converts the TaskBuilder to a CreateTaskRequest.
func (b *TaskBuilder) build() CreateTaskRequest {
	req := CreateTaskRequest{
		Content:  b.content,
		ParentID: b.parentID,
		Position: b.position,
		Due:      b.due,
		Priority: b.priority,
		Repeat:   b.repeat,
	}
	if len(b.tags) > 0 {
		for i, tag := range b.tags {
			if i > 0 {
				req.Tags += ", "
			}
			req.Tags += tag
		}
	}
	return req
}

// Create creates a new task using a TaskBuilder.
func (s *TaskService) Create(ctx context.Context, builder *TaskBuilder) (*Task, error) {
	path := fmt.Sprintf("/checklists/%d/tasks.json", s.checklistID)
	body := createTaskWrapper{Task: builder.build()}

	var task Task
	if err := s.client.doPost(ctx, path, body, &task); err != nil {
		return nil, err
	}

	parseDueDate(&task)
	return &task, nil
}

// UpdateTaskRequest represents the request body for updating a task.
type UpdateTaskRequest struct {
	Content  *string `json:"content,omitempty"`
	ParentID *int    `json:"parent_id,omitempty"`
	Position *int    `json:"position,omitempty"`
	Due      *string `json:"due_date,omitempty"`
	Priority *int    `json:"priority,omitempty"`
	Tags     *string `json:"tags,omitempty"`
}

// updateTaskWrapper wraps the task fields for PUT requests
type updateTaskWrapper struct {
	Task UpdateTaskRequest `json:"task"`
}

// Update updates an existing task.
func (s *TaskService) Update(ctx context.Context, taskID int, req UpdateTaskRequest) (*Task, error) {
	path := fmt.Sprintf("/checklists/%d/tasks/%d.json", s.checklistID, taskID)
	body := updateTaskWrapper{Task: req}

	var task Task
	if err := s.client.doPut(ctx, path, body, &task); err != nil {
		return nil, err
	}

	parseDueDate(&task)
	return &task, nil
}

// Delete permanently deletes a task.
func (s *TaskService) Delete(ctx context.Context, taskID int) error {
	path := fmt.Sprintf("/checklists/%d/tasks/%d.json", s.checklistID, taskID)
	return s.client.doDelete(ctx, path)
}

// Close marks a task as completed.
func (s *TaskService) Close(ctx context.Context, taskID int) (*Task, error) {
	path := fmt.Sprintf("/checklists/%d/tasks/%d/close.json", s.checklistID, taskID)

	var task Task
	if err := s.client.doPost(ctx, path, nil, &task); err != nil {
		return nil, err
	}

	parseDueDate(&task)
	return &task, nil
}

// Reopen reopens a closed or invalidated task.
func (s *TaskService) Reopen(ctx context.Context, taskID int) (*Task, error) {
	path := fmt.Sprintf("/checklists/%d/tasks/%d/reopen.json", s.checklistID, taskID)

	var task Task
	if err := s.client.doPost(ctx, path, nil, &task); err != nil {
		return nil, err
	}

	parseDueDate(&task)
	return &task, nil
}

// Invalidate marks a task as invalidated (strikethrough).
func (s *TaskService) Invalidate(ctx context.Context, taskID int) (*Task, error) {
	path := fmt.Sprintf("/checklists/%d/tasks/%d/invalidate.json", s.checklistID, taskID)

	var task Task
	if err := s.client.doPost(ctx, path, nil, &task); err != nil {
		return nil, err
	}

	parseDueDate(&task)
	return &task, nil
}

// parseDueDate attempts to parse the DueDateRaw string into a time.Time.
// It supports ISO 8601 date format (YYYY-MM-DD).
func parseDueDate(task *Task) {
	if task.DueDateRaw == "" {
		return
	}

	// Try to parse as ISO date
	if t, err := time.Parse("2006-01-02", task.DueDateRaw); err == nil {
		task.DueDate = &t
	}
}
