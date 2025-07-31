package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"seta-training/internal/middleware"
	"seta-training/internal/services"
)

type AssetHandler struct {
	folderService services.FolderServiceInterface
	noteService   services.NoteServiceInterface
	teamService   services.TeamServiceInterface
}

func NewAssetHandler(folderService services.FolderServiceInterface, noteService services.NoteServiceInterface, teamService services.TeamServiceInterface) *AssetHandler {
	return &AssetHandler{
		folderService: folderService,
		noteService:   noteService,
		teamService:   teamService,
	}
}

// GetUserAssets gets all assets owned by or shared with a user
func (h *AssetHandler) GetUserAssets(c *gin.Context) {
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

	// Only managers can view other users' assets, or users can view their own
	if claims.UserID != userID && claims.Role != "manager" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Insufficient permissions",
		})
		return
	}

	// Get user's folders
	folders, err := h.folderService.GetUserFolders(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get user folders: " + err.Error(),
		})
		return
	}

	// Get user's notes
	notes, err := h.noteService.GetUserNotes(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get user notes: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"folders": folders,
		"notes":   notes,
	})
}

// GetTeamAssets gets all assets that team members own or can access (managers only)
func (h *AssetHandler) GetTeamAssets(c *gin.Context) {
	teamIDStr := c.Param("teamId")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid team ID",
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

	// Only managers can view team assets
	if claims.Role != "manager" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Only managers can view team assets",
		})
		return
	}

	// Verify user is a manager of this team
	team, err := h.teamService.GetTeam(teamID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Team not found",
		})
		return
	}

	// Check if current user is a manager of this team
	isManager := false
	for _, manager := range team.Managers {
		if manager.ID == claims.UserID {
			isManager = true
			break
		}
	}

	if !isManager {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "You are not a manager of this team",
		})
		return
	}

	// Get all team members (including managers)
	allMembers := append(team.Members, team.Managers...)
	
	// Collect all assets from team members
	var allFolders []interface{}
	var allNotes []interface{}

	for _, member := range allMembers {
		// Get member's folders
		folders, err := h.folderService.GetUserFolders(member.ID)
		if err != nil {
			continue // Skip on error, don't fail the entire request
		}
		
		for _, folder := range folders {
			allFolders = append(allFolders, gin.H{
				"folder": folder,
				"owner":  member,
			})
		}

		// Get member's notes
		notes, err := h.noteService.GetUserNotes(member.ID)
		if err != nil {
			continue // Skip on error, don't fail the entire request
		}
		
		for _, note := range notes {
			allNotes = append(allNotes, gin.H{
				"note":  note,
				"owner": member,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"team_id": teamID,
		"team_name": team.Name,
		"folders": allFolders,
		"notes":   allNotes,
		"total_folders": len(allFolders),
		"total_notes":   len(allNotes),
	})
}
