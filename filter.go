package checkvist

import (
	"strings"
	"time"
)

// filter.go contains the Filter builder for client-side task filtering.
// The Checkvist API does not support server-side filtering, so filtering
// is performed locally after fetching all tasks.

// Filter provides a builder pattern for filtering tasks client-side.
type Filter struct {
	tasks   []Task
	filters []func(Task) bool
}

// NewFilter creates a new Filter with the given tasks.
func NewFilter(tasks []Task) *Filter {
	return &Filter{tasks: tasks}
}

// WithTag filters tasks that have the specified tag.
func (f *Filter) WithTag(tag string) *Filter {
	f.filters = append(f.filters, func(t Task) bool {
		return taskHasTag(t, tag)
	})
	return f
}

// WithTags filters tasks that have all of the specified tags (AND logic).
func (f *Filter) WithTags(tags ...string) *Filter {
	f.filters = append(f.filters, func(t Task) bool {
		for _, tag := range tags {
			if !taskHasTag(t, tag) {
				return false
			}
		}
		return true
	})
	return f
}

// WithStatus filters tasks by their status.
func (f *Filter) WithStatus(status TaskStatus) *Filter {
	f.filters = append(f.filters, func(t Task) bool {
		return t.Status == status
	})
	return f
}

// WithDueBefore filters tasks with due dates before the specified time.
func (f *Filter) WithDueBefore(deadline time.Time) *Filter {
	f.filters = append(f.filters, func(t Task) bool {
		if t.DueDate == nil {
			return false
		}
		return t.DueDate.Before(deadline)
	})
	return f
}

// WithDueAfter filters tasks with due dates after the specified time.
func (f *Filter) WithDueAfter(after time.Time) *Filter {
	f.filters = append(f.filters, func(t Task) bool {
		if t.DueDate == nil {
			return false
		}
		return t.DueDate.After(after)
	})
	return f
}

// WithDueOn filters tasks with due dates on the specified day.
func (f *Filter) WithDueOn(day time.Time) *Filter {
	year, month, d := day.Date()
	f.filters = append(f.filters, func(t Task) bool {
		if t.DueDate == nil {
			return false
		}
		ty, tm, td := t.DueDate.Date()
		return ty == year && tm == month && td == d
	})
	return f
}

// WithOverdue filters tasks that are overdue (due date is before today).
func (f *Filter) WithOverdue() *Filter {
	today := time.Now().Truncate(24 * time.Hour)
	f.filters = append(f.filters, func(t Task) bool {
		if t.DueDate == nil {
			return false
		}
		return t.DueDate.Before(today) && t.Status == StatusOpen
	})
	return f
}

// WithSearch filters tasks whose content contains the search query (case-insensitive).
func (f *Filter) WithSearch(query string) *Filter {
	lowerQuery := strings.ToLower(query)
	f.filters = append(f.filters, func(t Task) bool {
		return strings.Contains(strings.ToLower(t.Content), lowerQuery)
	})
	return f
}

// Apply applies all filters and returns the filtered tasks.
func (f *Filter) Apply() []Task {
	if len(f.filters) == 0 {
		result := make([]Task, len(f.tasks))
		copy(result, f.tasks)
		return result
	}

	result := make([]Task, 0, len(f.tasks))
	for _, task := range f.tasks {
		if f.matches(task) {
			result = append(result, task)
		}
	}
	return result
}

// matches checks if a task matches all filters.
func (f *Filter) matches(task Task) bool {
	for _, filter := range f.filters {
		if !filter(task) {
			return false
		}
	}
	return true
}

// taskHasTag checks if a task has a specific tag.
func taskHasTag(t Task, tag string) bool {
	// Check parsed Tags map first
	if t.Tags != nil && t.Tags[tag] {
		return true
	}
	// Fall back to checking TagsAsText
	if t.TagsAsText == "" {
		return false
	}
	lowerTag := strings.ToLower(tag)
	for _, part := range strings.Split(t.TagsAsText, ",") {
		if strings.ToLower(strings.TrimSpace(part)) == lowerTag {
			return true
		}
	}
	return false
}
