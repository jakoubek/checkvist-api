package checkvist

import (
	"testing"
	"time"
)

func TestFilter_WithTag(t *testing.T) {
	tasks := []Task{
		{ID: 1, Content: "Task 1", TagsAsText: "important, urgent"},
		{ID: 2, Content: "Task 2", TagsAsText: "urgent"},
		{ID: 3, Content: "Task 3", TagsAsText: ""},
	}

	tests := []struct {
		name     string
		tag      string
		expected []int
	}{
		{"filter by important", "important", []int{1}},
		{"filter by urgent", "urgent", []int{1, 2}},
		{"filter by nonexistent", "nonexistent", []int{}},
		{"case insensitive", "IMPORTANT", []int{1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewFilter(tasks).WithTag(tt.tag).Apply()
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d tasks, got %d", len(tt.expected), len(result))
				return
			}
			for i, task := range result {
				if task.ID != tt.expected[i] {
					t.Errorf("expected task ID %d at index %d, got %d", tt.expected[i], i, task.ID)
				}
			}
		})
	}
}

func TestFilter_WithMultipleTags(t *testing.T) {
	tasks := []Task{
		{ID: 1, Content: "Task 1", TagsAsText: "important, urgent, work"},
		{ID: 2, Content: "Task 2", TagsAsText: "urgent, work"},
		{ID: 3, Content: "Task 3", TagsAsText: "important"},
	}

	tests := []struct {
		name     string
		tags     []string
		expected []int
	}{
		{"single tag", []string{"urgent"}, []int{1, 2}},
		{"two tags AND", []string{"urgent", "work"}, []int{1, 2}},
		{"three tags AND", []string{"important", "urgent", "work"}, []int{1}},
		{"no match", []string{"important", "urgent", "missing"}, []int{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewFilter(tasks).WithTags(tt.tags...).Apply()
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d tasks, got %d", len(tt.expected), len(result))
				return
			}
			for i, task := range result {
				if task.ID != tt.expected[i] {
					t.Errorf("expected task ID %d at index %d, got %d", tt.expected[i], i, task.ID)
				}
			}
		})
	}
}

func TestFilter_WithStatus(t *testing.T) {
	tasks := []Task{
		{ID: 1, Content: "Open task", Status: StatusOpen},
		{ID: 2, Content: "Closed task", Status: StatusClosed},
		{ID: 3, Content: "Invalidated task", Status: StatusInvalidated},
		{ID: 4, Content: "Another open", Status: StatusOpen},
	}

	tests := []struct {
		name     string
		status   TaskStatus
		expected []int
	}{
		{"open tasks", StatusOpen, []int{1, 4}},
		{"closed tasks", StatusClosed, []int{2}},
		{"invalidated tasks", StatusInvalidated, []int{3}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewFilter(tasks).WithStatus(tt.status).Apply()
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d tasks, got %d", len(tt.expected), len(result))
				return
			}
			for i, task := range result {
				if task.ID != tt.expected[i] {
					t.Errorf("expected task ID %d at index %d, got %d", tt.expected[i], i, task.ID)
				}
			}
		})
	}
}

func TestFilter_WithDueBefore(t *testing.T) {
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	tomorrow := now.AddDate(0, 0, 1)
	nextWeek := now.AddDate(0, 0, 7)

	tasks := []Task{
		{ID: 1, Content: "Yesterday", DueDate: &yesterday},
		{ID: 2, Content: "Tomorrow", DueDate: &tomorrow},
		{ID: 3, Content: "Next week", DueDate: &nextWeek},
		{ID: 4, Content: "No due date", DueDate: nil},
	}

	result := NewFilter(tasks).WithDueBefore(now).Apply()
	if len(result) != 1 {
		t.Errorf("expected 1 task, got %d", len(result))
		return
	}
	if result[0].ID != 1 {
		t.Errorf("expected task ID 1, got %d", result[0].ID)
	}
}

func TestFilter_WithDueAfter(t *testing.T) {
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	tomorrow := now.AddDate(0, 0, 1)
	nextWeek := now.AddDate(0, 0, 7)

	tasks := []Task{
		{ID: 1, Content: "Yesterday", DueDate: &yesterday},
		{ID: 2, Content: "Tomorrow", DueDate: &tomorrow},
		{ID: 3, Content: "Next week", DueDate: &nextWeek},
		{ID: 4, Content: "No due date", DueDate: nil},
	}

	result := NewFilter(tasks).WithDueAfter(now).Apply()
	if len(result) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(result))
		return
	}
	if result[0].ID != 2 || result[1].ID != 3 {
		t.Errorf("expected task IDs [2, 3], got [%d, %d]", result[0].ID, result[1].ID)
	}
}

func TestFilter_WithDueOn(t *testing.T) {
	today := time.Now().Truncate(24 * time.Hour)
	todayMorning := today.Add(9 * time.Hour)
	todayEvening := today.Add(18 * time.Hour)
	tomorrow := today.AddDate(0, 0, 1)

	tasks := []Task{
		{ID: 1, Content: "Today morning", DueDate: &todayMorning},
		{ID: 2, Content: "Today evening", DueDate: &todayEvening},
		{ID: 3, Content: "Tomorrow", DueDate: &tomorrow},
		{ID: 4, Content: "No due date", DueDate: nil},
	}

	result := NewFilter(tasks).WithDueOn(today).Apply()
	if len(result) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(result))
		return
	}
	if result[0].ID != 1 || result[1].ID != 2 {
		t.Errorf("expected task IDs [1, 2], got [%d, %d]", result[0].ID, result[1].ID)
	}
}

func TestFilter_WithOverdue(t *testing.T) {
	today := time.Now().Truncate(24 * time.Hour)
	yesterday := today.AddDate(0, 0, -1)
	tomorrow := today.AddDate(0, 0, 1)

	tasks := []Task{
		{ID: 1, Content: "Overdue open", DueDate: &yesterday, Status: StatusOpen},
		{ID: 2, Content: "Overdue closed", DueDate: &yesterday, Status: StatusClosed},
		{ID: 3, Content: "Future open", DueDate: &tomorrow, Status: StatusOpen},
		{ID: 4, Content: "No due date", DueDate: nil, Status: StatusOpen},
	}

	result := NewFilter(tasks).WithOverdue().Apply()
	if len(result) != 1 {
		t.Errorf("expected 1 task, got %d", len(result))
		return
	}
	if result[0].ID != 1 {
		t.Errorf("expected task ID 1, got %d", result[0].ID)
	}
}

func TestFilter_WithSearch(t *testing.T) {
	tasks := []Task{
		{ID: 1, Content: "Buy groceries"},
		{ID: 2, Content: "Call the doctor"},
		{ID: 3, Content: "Review PR for grocery app"},
		{ID: 4, Content: "Send email"},
	}

	tests := []struct {
		name     string
		query    string
		expected []int
	}{
		{"exact match", "groceries", []int{1}},
		{"partial match", "grocer", []int{1, 3}},
		{"case insensitive", "CALL", []int{2}},
		{"no match", "xyz", []int{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewFilter(tasks).WithSearch(tt.query).Apply()
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d tasks, got %d", len(tt.expected), len(result))
				return
			}
			for i, task := range result {
				if task.ID != tt.expected[i] {
					t.Errorf("expected task ID %d at index %d, got %d", tt.expected[i], i, task.ID)
				}
			}
		})
	}
}

func TestFilter_Combined(t *testing.T) {
	today := time.Now().Truncate(24 * time.Hour)
	yesterday := today.AddDate(0, 0, -1)
	tomorrow := today.AddDate(0, 0, 1)

	tasks := []Task{
		{ID: 1, Content: "Important task", TagsAsText: "important", Status: StatusOpen, DueDate: &yesterday},
		{ID: 2, Content: "Important closed", TagsAsText: "important", Status: StatusClosed, DueDate: &yesterday},
		{ID: 3, Content: "Regular task", TagsAsText: "regular", Status: StatusOpen, DueDate: &yesterday},
		{ID: 4, Content: "Important future", TagsAsText: "important", Status: StatusOpen, DueDate: &tomorrow},
	}

	// Filter: important AND open AND due before today
	result := NewFilter(tasks).
		WithTag("important").
		WithStatus(StatusOpen).
		WithDueBefore(today).
		Apply()

	if len(result) != 1 {
		t.Errorf("expected 1 task, got %d", len(result))
		return
	}
	if result[0].ID != 1 {
		t.Errorf("expected task ID 1, got %d", result[0].ID)
	}
}

func TestFilter_EmptyFilters(t *testing.T) {
	tasks := []Task{
		{ID: 1, Content: "Task 1"},
		{ID: 2, Content: "Task 2"},
	}

	result := NewFilter(tasks).Apply()
	if len(result) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(result))
	}
}

func TestFilter_EmptyTasks(t *testing.T) {
	result := NewFilter([]Task{}).WithTag("test").Apply()
	if len(result) != 0 {
		t.Errorf("expected 0 tasks, got %d", len(result))
	}
}

func TestFilter_Performance_1000Tasks(t *testing.T) {
	// Create 1000 tasks with various attributes
	tasks := make([]Task, 1000)
	now := time.Now()

	for i := 0; i < 1000; i++ {
		due := now.AddDate(0, 0, i-500) // -500 to +499 days from now
		tasks[i] = Task{
			ID:         i + 1,
			Content:    "Task with some content to search through",
			TagsAsText: "tag1, tag2, tag3",
			Status:     TaskStatus(i % 3),
			DueDate:    &due,
		}
	}

	// Mark important tasks
	for i := 0; i < 100; i++ {
		tasks[i].TagsAsText = "important, urgent"
	}

	start := time.Now()

	// Run a complex filter
	result := NewFilter(tasks).
		WithTag("important").
		WithStatus(StatusOpen).
		WithDueBefore(now).
		WithSearch("content").
		Apply()

	elapsed := time.Since(start)

	// Performance target: <10ms for 1000 tasks
	if elapsed > 10*time.Millisecond {
		t.Errorf("filter took %v, expected <10ms", elapsed)
	}

	// Verify we got some results
	if len(result) == 0 {
		t.Log("No matching tasks found, but performance test passed")
	}
}

func TestFilter_TagsMap(t *testing.T) {
	// Test that Tags map is also checked
	tasks := []Task{
		{ID: 1, Content: "Task 1", Tags: Tags{"important": true}, TagsAsText: ""},
		{ID: 2, Content: "Task 2", Tags: nil, TagsAsText: "important"},
	}

	result := NewFilter(tasks).WithTag("important").Apply()
	if len(result) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(result))
	}
}
