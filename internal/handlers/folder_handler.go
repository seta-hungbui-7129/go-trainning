package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"seta-training/internal/middleware"
	"seta-training/internal/services"
)

type FolderHandler struct {
	folderService services.FolderServiceInterface
}

func NewFolderHandler(folderService services.FolderServiceInterface) *FolderHandler {
	return &FolderHandler{
		folderService: folderService,
	}
}

// CreateFolder creates a new folder
func (h *FolderHandler) CreateFolder(c *gin.Context) {
	var input services.CreateFolderInput
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

	folder, err := h.folderService.CreateFolder(&input, claims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, folder)
}

// GetFolder gets folder details
func (h *FolderHandler) GetFolder(c *gin.Context) {
	folderIDStr := c.Param("folderId")
	folderID, err := uuid.Parse(folderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid folder ID",
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

	folder, err := h.folderService.GetFolder(folderID, claims.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, folder)
}

// UpdateFolder updates folder details
func (h *FolderHandler) UpdateFolder(c *gin.Context) {
	folderIDStr := c.Param("folderId")
	folderID, err := uuid.Parse(folderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid folder ID",
		})
		return
	}

	var input services.UpdateFolderInput
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

	folder, err := h.folderService.UpdateFolder(folderID, &input, claims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, folder)
}

// DeleteFolder deletes a folder
func (h *FolderHandler) DeleteFolder(c *gin.Context) {
	folderIDStr := c.Param("folderId")
	folderID, err := uuid.Parse(folderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid folder ID",
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

	err = h.folderService.DeleteFolder(folderID, claims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Folder deleted successfully",
	})
}

// ShareFolder shares a folder with another user
func (h *FolderHandler) ShareFolder(c *gin.Context) {
	folderIDStr := c.Param("folderId")
	folderID, err := uuid.Parse(folderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid folder ID",
		})
		return
	}

	var input services.ShareFolderInput
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

	err = h.folderService.ShareFolder(folderID, &input, claims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Folder shared successfully",
	})
}

// RevokeShare revokes folder sharing from a user
func (h *FolderHandler) RevokeShare(c *gin.Context) {
	folderIDStr := c.Param("folderId")
	folderID, err := uuid.Parse(folderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid folder ID",
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

	err = h.folderService.RevokeShare(folderID, userID, claims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Folder sharing revoked successfully",
	})
}
