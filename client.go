// Package checkvist provides a type-safe, idiomatic Go client for the Checkvist API.
//
// This package allows Go applications to interact with Checkvist checklists,
// tasks, and notes. It handles authentication, automatic token renewal,
// and provides fluent interfaces for task creation and filtering.
//
// Basic usage:
//
//	client := checkvist.NewClient(username, remoteKey)
//	checklists, err := client.Checklists().List(ctx)
package checkvist

// client.go contains the Client struct, constructor, and authentication logic.
