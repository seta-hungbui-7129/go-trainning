package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"seta-training/internal/middleware"
	"seta-training/internal/services"
	"seta-training/pkg/logger"
	"seta-training/pkg/metrics"
)

// ImportHandler handles CSV import operations
type ImportHandler struct {
	importService services.ImportServiceInterface
	logger        logger.Logger
	metrics       *metrics.Metrics
}

// NewImportHandler creates a new import handler
func NewImportHandler(importService services.ImportServiceInterface, logger logger.Logger, metrics *metrics.Metrics) *ImportHandler {
	return &ImportHandler{
		importService: importService,
		logger:        logger,
		metrics:       metrics,
	}
}

// ImportUsersRequest represents the request structure for import configuration
type ImportUsersRequest struct {
	WorkerCount    int  `form:"worker_count" json:"worker_count"`
	BatchSize      int  `form:"batch_size" json:"batch_size"`
	MaxRecords     int  `form:"max_records" json:"max_records"`
	SkipDuplicates bool `form:"skip_duplicates" json:"skip_duplicates"`
	TimeoutSeconds int  `form:"timeout_seconds" json:"timeout_seconds"`
}

// ImportUsers handles POST /import-users endpoint
func (h *ImportHandler) ImportUsers(c *gin.Context) {
	startTime := time.Now()
	
	// Get current user from context (only managers can import users)
	claims, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	// Only managers can import users
	if claims.Role != "manager" {
		h.logger.Warn("Non-manager attempted user import",
			logger.String("user_id", claims.UserID.String()),
			logger.String("role", string(claims.Role)),
		)
		h.metrics.RecordError("authorization", "import_handler")
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Only managers can import users",
		})
		return
	}

	h.logger.Info("User import request started",
		logger.String("manager_id", claims.UserID.String()),
		logger.String("client_ip", c.ClientIP()),
	)

	// Parse multipart form
	err := c.Request.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		h.logger.Error("Failed to parse multipart form", logger.Error(err))
		h.metrics.RecordError("validation", "import_handler")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to parse form data: " + err.Error(),
		})
		return
	}

	// Get CSV file from form
	file, header, err := c.Request.FormFile("csv_file")
	if err != nil {
		h.logger.Error("Failed to get CSV file from form", logger.Error(err))
		h.metrics.RecordError("validation", "import_handler")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "CSV file is required. Please upload a file with key 'csv_file'",
		})
		return
	}
	defer file.Close()

	// Validate file type
	if header.Header.Get("Content-Type") != "text/csv" && 
	   !isCSVFile(header.Filename) {
		h.logger.Warn("Invalid file type uploaded",
			logger.String("filename", header.Filename),
			logger.String("content_type", header.Header.Get("Content-Type")),
		)
		h.metrics.RecordError("validation", "import_handler")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "File must be a CSV file (.csv extension or text/csv content type)",
		})
		return
	}

	// Validate file size (max 5MB)
	const maxFileSize = 5 << 20 // 5 MB
	if header.Size > maxFileSize {
		h.logger.Warn("File too large",
			logger.String("filename", header.Filename),
			logger.Int("size_bytes", int(header.Size)),
			logger.Int("max_size_bytes", maxFileSize),
		)
		h.metrics.RecordError("validation", "import_handler")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("File size too large. Maximum allowed: %d MB", maxFileSize/(1<<20)),
		})
		return
	}

	h.logger.Info("CSV file received",
		logger.String("filename", header.Filename),
		logger.Int("size_bytes", int(header.Size)),
		logger.String("content_type", header.Header.Get("Content-Type")),
	)

	// Parse import configuration from form or use defaults
	config := h.parseImportConfig(c)
	
	h.logger.Info("Import configuration",
		logger.Int("worker_count", config.WorkerCount),
		logger.Int("batch_size", config.BatchSize),
		logger.Int("max_records", config.MaxRecords),
		logger.Duration("timeout", config.Timeout),
		logger.Any("skip_duplicates", config.SkipDuplicates),
	)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	// Process CSV import
	summary, err := h.importService.ImportUsersFromCSV(ctx, file, config)
	if err != nil {
		h.logger.Error("CSV import failed", logger.Error(err))
		h.metrics.RecordError("processing", "import_handler")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process CSV import: " + err.Error(),
		})
		return
	}

	// Record metrics
	h.metrics.RecordDatabaseQuery("bulk_insert", "users")
	
	// Log summary
	h.logger.Info("CSV import completed",
		logger.String("manager_id", claims.UserID.String()),
		logger.String("filename", header.Filename),
		logger.Int("total_records", summary.TotalRecords),
		logger.Int("success_count", summary.SuccessCount),
		logger.Int("failure_count", summary.FailureCount),
		logger.String("processing_time", summary.ProcessingTime),
		logger.Duration("total_time", time.Since(startTime)),
	)

	// Return success response with summary
	response := gin.H{
		"message": "CSV import completed",
		"summary": summary,
		"file_info": gin.H{
			"filename":     header.Filename,
			"size_bytes":   header.Size,
			"content_type": header.Header.Get("Content-Type"),
		},
		"config": gin.H{
			"worker_count":    config.WorkerCount,
			"batch_size":      config.BatchSize,
			"max_records":     config.MaxRecords,
			"timeout_seconds": int(config.Timeout.Seconds()),
			"skip_duplicates": config.SkipDuplicates,
		},
		"processed_by": gin.H{
			"manager_id": claims.UserID.String(),
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
		},
	}

	// Set appropriate status code based on results
	statusCode := http.StatusOK
	if summary.FailureCount > 0 && summary.SuccessCount == 0 {
		statusCode = http.StatusBadRequest // All failed
	} else if summary.FailureCount > 0 {
		statusCode = http.StatusPartialContent // Some failed
	}

	c.JSON(statusCode, response)
}

// parseImportConfig parses import configuration from request or returns defaults
func (h *ImportHandler) parseImportConfig(c *gin.Context) services.ImportConfig {
	config := services.DefaultImportConfig()

	// Parse worker count
	if workerCountStr := c.PostForm("worker_count"); workerCountStr != "" {
		if workerCount, err := strconv.Atoi(workerCountStr); err == nil && workerCount > 0 && workerCount <= 20 {
			config.WorkerCount = workerCount
		}
	}

	// Parse batch size
	if batchSizeStr := c.PostForm("batch_size"); batchSizeStr != "" {
		if batchSize, err := strconv.Atoi(batchSizeStr); err == nil && batchSize > 0 && batchSize <= 1000 {
			config.BatchSize = batchSize
		}
	}

	// Parse max records
	if maxRecordsStr := c.PostForm("max_records"); maxRecordsStr != "" {
		if maxRecords, err := strconv.Atoi(maxRecordsStr); err == nil && maxRecords > 0 && maxRecords <= 10000 {
			config.MaxRecords = maxRecords
		}
	}

	// Parse timeout
	if timeoutStr := c.PostForm("timeout_seconds"); timeoutStr != "" {
		if timeoutSecs, err := strconv.Atoi(timeoutStr); err == nil && timeoutSecs > 0 && timeoutSecs <= 300 {
			config.Timeout = time.Duration(timeoutSecs) * time.Second
		}
	}

	// Parse skip duplicates
	if skipDuplicatesStr := c.PostForm("skip_duplicates"); skipDuplicatesStr != "" {
		config.SkipDuplicates = skipDuplicatesStr == "true" || skipDuplicatesStr == "1"
	}

	return config
}

// isCSVFile checks if filename has CSV extension
func isCSVFile(filename string) bool {
	return len(filename) > 4 && filename[len(filename)-4:] == ".csv"
}

// GetImportTemplate returns a CSV template for user import
func (h *ImportHandler) GetImportTemplate(c *gin.Context) {
	// Only authenticated users can download template
	_, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	// CSV template content
	template := `username,email,password,role
john.doe,john.doe@example.com,password123,manager
jane.smith,jane.smith@example.com,password456,member
bob.wilson,bob.wilson@example.com,password789,member`

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=user_import_template.csv")
	c.String(http.StatusOK, template)
}

// GetImportStatus returns the status of recent imports (could be extended for async processing)
func (h *ImportHandler) GetImportStatus(c *gin.Context) {
	// Only managers can check import status
	claims, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	if string(claims.Role) != "manager" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Only managers can check import status",
		})
		return
	}

	// For now, return basic info about import capabilities
	// This could be extended to track async import jobs
	c.JSON(http.StatusOK, gin.H{
		"import_capabilities": gin.H{
			"max_file_size_mb":     5,
			"max_records":          10000,
			"max_workers":          20,
			"max_timeout_seconds":  300,
			"supported_formats":    []string{"CSV"},
			"required_columns":     []string{"username", "email", "password", "role"},
			"supported_roles":      []string{"manager", "member"},
		},
		"current_limits": gin.H{
			"concurrent_imports": 1, // Currently synchronous
			"queue_size":        0,
		},
	})
}
