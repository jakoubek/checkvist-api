package checkvist_test

import (
	"context"
	"fmt"
	"time"

	"code.beautifulmachines.dev/jakoubek/checkvist-api"
)

// This file contains GoDoc examples that appear in the package documentation.
// All examples are runnable but require valid credentials to actually work.

func Example_basicUsage() {
	// Create a new client with your credentials
	client := checkvist.NewClient("user@example.com", "your-api-key")

	// List all checklists
	ctx := context.Background()
	checklists, err := client.Checklists().List(ctx)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	for _, cl := range checklists {
		fmt.Printf("Checklist: %s (%d tasks)\n", cl.Name, cl.TaskCount)
	}
}

func ExampleNewClient() {
	// Basic client creation
	client := checkvist.NewClient("user@example.com", "your-api-key")
	_ = client

	// With custom options
	client = checkvist.NewClient("user@example.com", "your-api-key",
		checkvist.WithTimeout(60*time.Second),
		checkvist.WithRetryConfig(checkvist.RetryConfig{
			MaxRetries: 5,
			BaseDelay:  500 * time.Millisecond,
			MaxDelay:   30 * time.Second,
		}),
	)
	_ = client
}

func ExampleClient_Authenticate() {
	client := checkvist.NewClient("user@example.com", "your-api-key")

	// Explicit authentication (usually not needed - auto-authenticates on first request)
	ctx := context.Background()
	err := client.Authenticate(ctx)
	if err != nil {
		fmt.Println("Authentication failed:", err)
		return
	}
	fmt.Println("Authenticated successfully")
}

func ExampleChecklistService_List() {
	client := checkvist.NewClient("user@example.com", "your-api-key")
	ctx := context.Background()

	// List all active checklists
	checklists, err := client.Checklists().List(ctx)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	for _, cl := range checklists {
		fmt.Printf("- %s (ID: %d)\n", cl.Name, cl.ID)
	}
}

func ExampleTaskService_Create() {
	client := checkvist.NewClient("user@example.com", "your-api-key")
	ctx := context.Background()

	// Create a task using the builder pattern
	task, err := client.Tasks(123).Create(ctx,
		checkvist.NewTask("Buy groceries").
			WithTags("shopping", "urgent").
			WithDueDate(checkvist.DueTomorrow).
			WithPriority(1),
	)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("Created task: %s (ID: %d)\n", task.Content, task.ID)
}

func ExampleNewTask() {
	// Simple task
	task := checkvist.NewTask("Complete project report")

	// Task with all options
	task = checkvist.NewTask("Review pull request").
		WithTags("code-review", "urgent").
		WithDueDate(checkvist.DueTomorrow).
		WithPriority(1).
		WithParent(456). // Makes this a subtask
		WithPosition(1)  // First position among siblings

	_ = task
}

func ExampleNewFilter() {
	// Assume we have a list of tasks from the API
	tasks := []checkvist.Task{
		{ID: 1, Content: "Buy groceries", TagsAsText: "shopping", Status: checkvist.StatusOpen},
		{ID: 2, Content: "Call doctor", TagsAsText: "health", Status: checkvist.StatusClosed},
		{ID: 3, Content: "Review PR", TagsAsText: "work, urgent", Status: checkvist.StatusOpen},
	}

	// Filter to open tasks with "urgent" tag
	filtered := checkvist.NewFilter(tasks).
		WithTag("urgent").
		WithStatus(checkvist.StatusOpen).
		Apply()

	fmt.Printf("Found %d matching tasks\n", len(filtered))
	// Output: Found 1 matching tasks
}

func ExampleFilter_Apply() {
	// Create sample tasks with due dates
	yesterday := time.Now().AddDate(0, 0, -1)
	tomorrow := time.Now().AddDate(0, 0, 1)

	tasks := []checkvist.Task{
		{ID: 1, Content: "Overdue task", DueDate: &yesterday, Status: checkvist.StatusOpen},
		{ID: 2, Content: "Future task", DueDate: &tomorrow, Status: checkvist.StatusOpen},
		{ID: 3, Content: "Closed task", DueDate: &yesterday, Status: checkvist.StatusClosed},
	}

	// Find overdue open tasks
	overdue := checkvist.NewFilter(tasks).
		WithOverdue().
		Apply()

	for _, t := range overdue {
		fmt.Printf("Overdue: %s\n", t.Content)
	}
	// Output: Overdue: Overdue task
}

func ExampleDueAt() {
	// Set due date to a specific date
	deadline := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	task := checkvist.NewTask("Project deadline").
		WithDueDate(checkvist.DueAt(deadline))
	_ = task
}

func ExampleDueInDays() {
	// Set due date to 7 days from now
	task := checkvist.NewTask("Follow up").
		WithDueDate(checkvist.DueInDays(7))
	_ = task
}
