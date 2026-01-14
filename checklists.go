package checkvist

import (
	"context"
	"fmt"
)

// checklists.go contains the ChecklistService for CRUD operations on checklists.

// ChecklistService provides operations on Checkvist checklists.
type ChecklistService struct {
	client *Client
}

// Checklists returns a ChecklistService for performing checklist operations.
func (c *Client) Checklists() *ChecklistService {
	return &ChecklistService{client: c}
}

// ListOptions configures the List operation.
type ListOptions struct {
	// Archived filters to show only archived checklists when true.
	Archived bool
}

// List returns all checklists accessible to the authenticated user.
func (s *ChecklistService) List(ctx context.Context) ([]Checklist, error) {
	return s.ListWithOptions(ctx, ListOptions{})
}

// ListWithOptions returns checklists with the specified options.
func (s *ChecklistService) ListWithOptions(ctx context.Context, opts ListOptions) ([]Checklist, error) {
	path := "/checklists.json"
	if opts.Archived {
		path += "?archived=true"
	}

	var checklists []Checklist
	if err := s.client.doGet(ctx, path, &checklists); err != nil {
		return nil, err
	}
	return checklists, nil
}

// Get returns a single checklist by ID.
func (s *ChecklistService) Get(ctx context.Context, id int) (*Checklist, error) {
	path := fmt.Sprintf("/checklists/%d.json", id)

	var checklist Checklist
	if err := s.client.doGet(ctx, path, &checklist); err != nil {
		return nil, err
	}
	return &checklist, nil
}

// createChecklistRequest is the request body for creating a checklist.
type createChecklistRequest struct {
	Name string `json:"name"`
}

// Create creates a new checklist with the given name.
func (s *ChecklistService) Create(ctx context.Context, name string) (*Checklist, error) {
	body := createChecklistRequest{Name: name}

	var checklist Checklist
	if err := s.client.doPost(ctx, "/checklists.json", body, &checklist); err != nil {
		return nil, err
	}
	return &checklist, nil
}

// updateChecklistRequest is the request body for updating a checklist.
type updateChecklistRequest struct {
	Name string `json:"name"`
}

// Update updates the name of an existing checklist.
func (s *ChecklistService) Update(ctx context.Context, id int, name string) (*Checklist, error) {
	path := fmt.Sprintf("/checklists/%d.json", id)
	body := updateChecklistRequest{Name: name}

	var checklist Checklist
	if err := s.client.doPut(ctx, path, body, &checklist); err != nil {
		return nil, err
	}
	return &checklist, nil
}

// Delete permanently deletes a checklist by ID.
func (s *ChecklistService) Delete(ctx context.Context, id int) error {
	path := fmt.Sprintf("/checklists/%d.json", id)
	return s.client.doDelete(ctx, path)
}

// archiveRequest is the request body for archiving/unarchiving a checklist.
type archiveRequest struct {
	Archived bool `json:"archived"`
}

// Archive archives a checklist by ID.
func (s *ChecklistService) Archive(ctx context.Context, id int) (*Checklist, error) {
	path := fmt.Sprintf("/checklists/%d.json", id)
	body := archiveRequest{Archived: true}

	var checklist Checklist
	if err := s.client.doPut(ctx, path, body, &checklist); err != nil {
		return nil, err
	}
	return &checklist, nil
}

// Unarchive unarchives a checklist by ID.
func (s *ChecklistService) Unarchive(ctx context.Context, id int) (*Checklist, error) {
	path := fmt.Sprintf("/checklists/%d.json", id)
	body := archiveRequest{Archived: false}

	var checklist Checklist
	if err := s.client.doPut(ctx, path, body, &checklist); err != nil {
		return nil, err
	}
	return &checklist, nil
}
