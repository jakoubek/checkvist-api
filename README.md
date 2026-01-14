# checkvist-api

A type-safe, idiomatic Go client library for the [Checkvist](https://checkvist.com/) API.

## Installation

```bash
go get code.beautifulmachines.dev/jakoubek/checkvist-api
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "code.beautifulmachines.dev/jakoubek/checkvist-api"
)

func main() {
    // Create a new client
    client := checkvist.NewClient("your-email@example.com", "your-api-key")

    ctx := context.Background()

    // List all checklists
    checklists, err := client.Checklists().List(ctx)
    if err != nil {
        log.Fatal(err)
    }

    for _, cl := range checklists {
        fmt.Printf("Checklist: %s (%d tasks)\n", cl.Name, cl.TaskCount)
    }

    // Create a new task
    task, err := client.Tasks(checklists[0].ID).Create(ctx,
        checkvist.NewTask("Buy groceries").
            WithDueDate(checkvist.DueTomorrow).
            WithPriority(1).
            WithTags("shopping", "personal"),
    )
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Created task: %s (ID: %d)\n", task.Content, task.ID)
}
```

## API Overview

### Checklists

```go
// List all checklists
checklists, err := client.Checklists().List(ctx)

// List archived checklists
archived, err := client.Checklists().ListWithOptions(ctx, checkvist.ListOptions{Archived: true})

// Get a single checklist
checklist, err := client.Checklists().Get(ctx, checklistID)

// Create a new checklist
checklist, err := client.Checklists().Create(ctx, "My New List")

// Update a checklist
checklist, err := client.Checklists().Update(ctx, checklistID, "New Name")

// Delete a checklist
err := client.Checklists().Delete(ctx, checklistID)
```

### Tasks

```go
// List all tasks in a checklist
tasks, err := client.Tasks(checklistID).List(ctx)

// Get a single task
task, err := client.Tasks(checklistID).Get(ctx, taskID)

// Create a task with the builder pattern
task, err := client.Tasks(checklistID).Create(ctx,
    checkvist.NewTask("Task content").
        WithParent(parentID).       // Create as subtask
        WithDueDate(checkvist.DueAt(time.Now().AddDate(0, 0, 7))).
        WithPriority(1).            // 1 = highest, 2 = high, 0 = normal
        WithTags("work", "urgent"),
)

// Update a task
content := "Updated content"
task, err := client.Tasks(checklistID).Update(ctx, taskID, checkvist.UpdateTaskRequest{
    Content: &content,
})

// Close/Reopen/Invalidate tasks
task, err := client.Tasks(checklistID).Close(ctx, taskID)
task, err := client.Tasks(checklistID).Reopen(ctx, taskID)
task, err := client.Tasks(checklistID).Invalidate(ctx, taskID)

// Delete a task
err := client.Tasks(checklistID).Delete(ctx, taskID)
```

### Notes (Comments)

```go
// List notes on a task
notes, err := client.Notes(checklistID, taskID).List(ctx)

// Create a note
note, err := client.Notes(checklistID, taskID).Create(ctx, "My comment")

// Update a note
note, err := client.Notes(checklistID, taskID).Update(ctx, noteID, "Updated comment")

// Delete a note
err := client.Notes(checklistID, taskID).Delete(ctx, noteID)
```

### Due Dates

The library provides convenient due date helpers:

```go
// Predefined constants
checkvist.DueToday
checkvist.DueTomorrow
checkvist.DueNextWeek
checkvist.DueNextMonth

// From time.Time
checkvist.DueAt(time.Now().AddDate(0, 0, 7))

// Relative days
checkvist.DueInDays(5)

// Raw smart syntax
checkvist.DueString("^friday")
```

## Error Handling

The library provides structured error types for API errors:

```go
task, err := client.Tasks(checklistID).Get(ctx, taskID)
if err != nil {
    // Check for specific error types
    if errors.Is(err, checkvist.ErrNotFound) {
        fmt.Println("Task not found")
    } else if errors.Is(err, checkvist.ErrUnauthorized) {
        fmt.Println("Invalid credentials")
    } else if errors.Is(err, checkvist.ErrRateLimited) {
        fmt.Println("Too many requests, please slow down")
    }

    // Get detailed error information
    var apiErr *checkvist.APIError
    if errors.As(err, &apiErr) {
        fmt.Printf("Status: %d, Message: %s\n", apiErr.StatusCode, apiErr.Message)
    }
}
```

Available sentinel errors:
- `ErrUnauthorized` - Invalid credentials (HTTP 401)
- `ErrNotFound` - Resource not found (HTTP 404)
- `ErrRateLimited` - Rate limit exceeded (HTTP 429)
- `ErrBadRequest` - Invalid request (HTTP 400)
- `ErrServerError` - Server error (HTTP 5xx)

## Configuration Options

Customize the client with functional options:

```go
client := checkvist.NewClient("email", "api-key",
    // Custom HTTP client
    checkvist.WithHTTPClient(customHTTPClient),

    // Custom timeout
    checkvist.WithTimeout(60 * time.Second),

    // Custom retry configuration
    checkvist.WithRetryConfig(checkvist.RetryConfig{
        MaxRetries: 5,
        BaseDelay:  2 * time.Second,
        MaxDelay:   60 * time.Second,
        Jitter:     true,
    }),

    // Custom logger
    checkvist.WithLogger(slog.New(slog.NewJSONHandler(os.Stdout, nil))),

    // Custom base URL (for testing)
    checkvist.WithBaseURL("https://custom-api.example.com"),
)
```

## Authentication

The library handles authentication automatically:
- Authenticates on first API call
- Stores token securely
- Automatically refreshes token before expiry
- Thread-safe for concurrent use

You can also authenticate explicitly:

```go
// Standard authentication
err := client.Authenticate(ctx)

// With 2FA
err := client.AuthenticateWith2FA(ctx, "123456")

// Get current user info
user, err := client.CurrentUser(ctx)
```

## Documentation

Full API documentation is available on [pkg.go.dev](https://pkg.go.dev/code.beautifulmachines.dev/jakoubek/checkvist-api).

## License

MIT License - see [LICENSE](LICENSE) file for details.
