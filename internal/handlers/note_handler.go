package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"seta-training/internal/middleware"
	"seta-training/internal/services"
)

type NoteHandler struct {
	noteService services.NoteServiceInterface
}

func NewNoteHandler(noteService services.NoteServiceInterface) *NoteHandler {
	return &NoteHandler{
		noteService: noteService,
	}
}

// CreateNote creates a new note in a folder
func (h *NoteHandler) CreateNote(c *gin.Context) {
	folderIDStr := c.Param("folderId")
	folderID, err := uuid.Parse(folderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid folder ID",
		})
		return
	}

	var input services.CreateNoteInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid input: " + err.Error(),
		})
		return
	}

	// Get current user from context
	claims, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	note, err := h.noteService.CreateNote(folderID, &input, claims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, note)
}

// GetNote gets note details
func (h *NoteHandler) GetNote(c *gin.Context) {
	noteIDStr := c.Param("noteId")
	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid note ID",
		})
		return
	}

	// Get current user from context
	claims, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	note, err := h.noteService.GetNote(noteID, claims.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, note)
}

// UpdateNote updates note details
func (h *NoteHandler) UpdateNote(c *gin.Context) {
	noteIDStr := c.Param("noteId")
	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid note ID",
		})
		return
	}

	var input services.UpdateNoteInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid input: " + err.Error(),
		})
		return
	}

	// Get current user from context
	claims, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	note, err := h.noteService.UpdateNote(noteID, &input, claims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, note)
}

// DeleteNote deletes a note
func (h *NoteHandler) DeleteNote(c *gin.Context) {
	noteIDStr := c.Param("noteId")
	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid note ID",
		})
		return
	}

	// Get current user from context
	claims, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	err = h.noteService.DeleteNote(noteID, claims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Note deleted successfully",
	})
}

// ShareNote shares a note with another user
func (h *NoteHandler) ShareNote(c *gin.Context) {
	noteIDStr := c.Param("noteId")
	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid note ID",
		})
		return
	}

	var input services.ShareNoteInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid input: " + err.Error(),
		})
		return
	}

	// Get current user from context
	claims, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	err = h.noteService.ShareNote(noteID, &input, claims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Note shared successfully",
	})
}

// RevokeShare revokes note sharing from a user
func (h *NoteHandler) RevokeShare(c *gin.Context) {
	noteIDStr := c.Param("noteId")
	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid note ID",
		})
		return
	}

	userIDStr := c.Param("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	// Get current user from context
	claims, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	err = h.noteService.RevokeShare(noteID, userID, claims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Note sharing revoked successfully",
	})
}
