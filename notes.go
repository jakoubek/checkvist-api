package checkvist

import (
	"context"
	"fmt"
)

// notes.go contains the NoteService for CRUD operations on notes (comments) attached to tasks.

// NoteService provides operations on notes (comments) attached to a specific task.
type NoteService struct {
	client      *Client
	checklistID int
	taskID      int
}

// Notes returns a NoteService for performing note operations on the specified task.
func (c *Client) Notes(checklistID, taskID int) *NoteService {
	return &NoteService{
		client:      c,
		checklistID: checklistID,
		taskID:      taskID,
	}
}

// List returns all notes (comments) attached to the task.
func (s *NoteService) List(ctx context.Context) ([]Note, error) {
	path := fmt.Sprintf("/checklists/%d/tasks/%d/comments.json", s.checklistID, s.taskID)

	var notes []Note
	if err := s.client.doGet(ctx, path, &notes); err != nil {
		return nil, err
	}
	return notes, nil
}

// createNoteRequest is the request body for creating a note.
type createNoteRequest struct {
	Comment string `json:"comment"`
}

// Create creates a new note (comment) on the task.
func (s *NoteService) Create(ctx context.Context, comment string) (*Note, error) {
	path := fmt.Sprintf("/checklists/%d/tasks/%d/comments.json", s.checklistID, s.taskID)
	body := createNoteRequest{Comment: comment}

	var note Note
	if err := s.client.doPost(ctx, path, body, &note); err != nil {
		return nil, err
	}
	return &note, nil
}

// updateNoteRequest is the request body for updating a note.
type updateNoteRequest struct {
	Comment string `json:"comment"`
}

// Update updates an existing note's comment text.
func (s *NoteService) Update(ctx context.Context, noteID int, comment string) (*Note, error) {
	path := fmt.Sprintf("/checklists/%d/tasks/%d/comments/%d.json", s.checklistID, s.taskID, noteID)
	body := updateNoteRequest{Comment: comment}

	var note Note
	if err := s.client.doPut(ctx, path, body, &note); err != nil {
		return nil, err
	}
	return &note, nil
}

// Delete permanently deletes a note.
func (s *NoteService) Delete(ctx context.Context, noteID int) error {
	path := fmt.Sprintf("/checklists/%d/tasks/%d/comments/%d.json", s.checklistID, s.taskID, noteID)
	return s.client.doDelete(ctx, path)
}
