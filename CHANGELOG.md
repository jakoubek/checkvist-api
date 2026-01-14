# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **Client**: Core API client with functional options pattern
  - `NewClient(username, remoteKey, ...Option)` constructor
  - `WithHTTPClient`, `WithTimeout`, `WithRetryConfig`, `WithLogger`, `WithBaseURL` options
  - Thread-safe token management with automatic renewal

- **Authentication**: Full authentication support
  - `Authenticate(ctx)` for explicit login
  - `AuthenticateWith2FA(ctx, token)` for 2FA support
  - Automatic token refresh before expiry
  - Token sent via `X-Client-Token` header

- **HTTP Request Helper**: Internal request handling with retry logic
  - Exponential backoff for HTTP 429 (rate limiting) and 5xx errors
  - Configurable retry settings via `RetryConfig`
  - Context cancellation support
  - Debug logging via `slog`

- **Data Models**: Type-safe structs for Checkvist entities
  - `Checklist` with ID, Name, Public, Archived, TaskCount, etc.
  - `Task` with Content, Status, Priority, DueDate, Tags, Children
  - `Note` for task comments
  - `User` for user information
  - `TaskStatus` enum (Open, Closed, Invalidated)
  - `DueDate` with smart syntax support (`DueToday`, `DueTomorrow`, `DueAt`, `DueInDays`)

- **Error Handling**: Structured error types
  - `APIError` with StatusCode, Message, RequestID
  - Sentinel errors: `ErrUnauthorized`, `ErrNotFound`, `ErrRateLimited`, `ErrBadRequest`, `ErrServerError`
  - Compatible with `errors.Is()` and `errors.As()`

- **Build Automation**: Mage build targets
  - `mage test` - run tests
  - `mage coverage` - run tests with coverage
  - `mage lint` - run staticcheck
  - `mage fmt` - format code
  - `mage check` - run all quality checks

- **Tests**: Comprehensive unit test suite
  - Client configuration tests
  - Authentication flow tests
  - Retry logic tests
  - httptest.Server-based mocking

- **Filter Builder**: Client-side task filtering (API has no server-side filtering)
  - `NewFilter(tasks)` constructor
  - `WithTag`, `WithTags` for tag filtering (AND logic)
  - `WithStatus` for status filtering (Open, Closed, Invalidated)
  - `WithDueBefore`, `WithDueAfter`, `WithDueOn` for due date filtering
  - `WithOverdue` for finding overdue open tasks
  - `WithSearch` for case-insensitive content search
  - `Apply()` returns filtered tasks
  - Performance: <10ms for 1000+ tasks

- **Checklist Archive**: Archive and unarchive checklists
  - `ChecklistService.Archive(ctx, id)` to archive a checklist
  - `ChecklistService.Unarchive(ctx, id)` to restore an archived checklist

- **Repeating Tasks**: Support for recurring task patterns
  - `TaskBuilder.WithRepeat(pattern)` using Checkvist smart syntax
  - Supports: daily, weekly, monthly, yearly, custom intervals

- **GoDoc Examples**: Runnable examples for documentation
  - `Example_basicUsage` - complete usage flow
  - `ExampleNewClient` - client configuration options
  - `ExampleTaskService_Create` - task creation with builder
  - `ExampleNewFilter`, `ExampleFilter_Apply` - filtering examples
  - `ExampleDueAt`, `ExampleDueInDays` - due date helpers
