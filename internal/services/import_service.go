package services

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"seta-training/internal/models"
	"seta-training/pkg/logger"
)

// ImportService handles CSV user imports with concurrent processing
type ImportService struct {
	userService UserServiceInterface
	logger      logger.Logger
}

// NewImportService creates a new import service
func NewImportService(userService UserServiceInterface, logger logger.Logger) *ImportService {
	return &ImportService{
		userService: userService,
		logger:      logger,
	}
}

// UserImportRecord represents a single user record from CSV
type UserImportRecord struct {
	Username string `csv:"username"`
	Email    string `csv:"email"`
	Password string `csv:"password"`
	Role     string `csv:"role"`
	LineNum  int    `csv:"-"` // Track line number for error reporting
}

// ImportResult represents the result of importing a single user
type ImportResult struct {
	Record  UserImportRecord `json:"record"`
	Success bool             `json:"success"`
	Error   string           `json:"error,omitempty"`
	UserID  string           `json:"user_id,omitempty"`
}

// ImportSummary represents the overall import summary
type ImportSummary struct {
	TotalRecords    int            `json:"total_records"`
	SuccessCount    int            `json:"success_count"`
	FailureCount    int            `json:"failure_count"`
	ProcessingTime  string         `json:"processing_time"`
	Results         []ImportResult `json:"results"`
	Errors          []string       `json:"errors,omitempty"`
}

// ImportConfig holds configuration for the import process
type ImportConfig struct {
	WorkerCount     int           `json:"worker_count"`
	BatchSize       int           `json:"batch_size"`
	Timeout         time.Duration `json:"timeout"`
	MaxRecords      int           `json:"max_records"`
	SkipDuplicates  bool          `json:"skip_duplicates"`
}

// DefaultImportConfig returns default configuration
func DefaultImportConfig() ImportConfig {
	return ImportConfig{
		WorkerCount:    5,  // Number of concurrent workers
		BatchSize:      100, // Records per batch
		Timeout:        30 * time.Second,
		MaxRecords:     1000, // Maximum records to process
		SkipDuplicates: true,
	}
}

// ImportUsersFromCSV processes CSV data concurrently using worker pools
func (s *ImportService) ImportUsersFromCSV(ctx context.Context, csvReader io.Reader, config ImportConfig) (*ImportSummary, error) {
	startTime := time.Now()
	
	s.logger.Info("Starting CSV user import",
		logger.Int("worker_count", config.WorkerCount),
		logger.Int("batch_size", config.BatchSize),
		logger.Int("max_records", config.MaxRecords),
	)

	// Parse CSV records
	records, err := s.parseCSVRecords(csvReader, config.MaxRecords)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	if len(records) == 0 {
		return &ImportSummary{
			TotalRecords:   0,
			SuccessCount:   0,
			FailureCount:   0,
			ProcessingTime: time.Since(startTime).String(),
			Results:        []ImportResult{},
		}, nil
	}

	s.logger.Info("Parsed CSV records", logger.Int("count", len(records)))

	// Create channels for worker communication
	recordChan := make(chan UserImportRecord, config.BatchSize)
	resultChan := make(chan ImportResult, len(records))

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, config.Timeout)
	defer cancel()

	// Start worker pool
	var wg sync.WaitGroup
	for i := 0; i < config.WorkerCount; i++ {
		wg.Add(1)
		go s.worker(ctx, i+1, recordChan, resultChan, &wg)
	}

	// Send records to workers
	go func() {
		defer close(recordChan)
		for _, record := range records {
			select {
			case recordChan <- record:
			case <-ctx.Done():
				s.logger.Warn("Context cancelled while sending records")
				return
			}
		}
	}()

	// Wait for all workers to complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	results := make([]ImportResult, 0, len(records))
	successCount := 0
	failureCount := 0

	for result := range resultChan {
		results = append(results, result)
		if result.Success {
			successCount++
		} else {
			failureCount++
		}
	}

	processingTime := time.Since(startTime)
	
	s.logger.Info("CSV import completed",
		logger.Int("total", len(records)),
		logger.Int("success", successCount),
		logger.Int("failed", failureCount),
		logger.Duration("duration", processingTime),
	)

	return &ImportSummary{
		TotalRecords:   len(records),
		SuccessCount:   successCount,
		FailureCount:   failureCount,
		ProcessingTime: processingTime.String(),
		Results:        results,
	}, nil
}

// parseCSVRecords parses CSV data into UserImportRecord structs
func (s *ImportService) parseCSVRecords(reader io.Reader, maxRecords int) ([]UserImportRecord, error) {
	csvReader := csv.NewReader(reader)
	csvReader.TrimLeadingSpace = true

	// Read header
	header, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Validate header
	expectedHeaders := []string{"username", "email", "password", "role"}
	if !s.validateHeader(header, expectedHeaders) {
		return nil, fmt.Errorf("invalid CSV header. Expected: %v, Got: %v", expectedHeaders, header)
	}

	var records []UserImportRecord
	lineNum := 2 // Start from line 2 (after header)

	for {
		if maxRecords > 0 && len(records) >= maxRecords {
			s.logger.Warn("Reached maximum record limit", logger.Int("max_records", maxRecords))
			break
		}

		row, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			s.logger.Error("Error reading CSV row", 
				logger.Int("line", lineNum),
				logger.Error(err),
			)
			lineNum++
			continue
		}

		if len(row) < 4 {
			s.logger.Warn("Skipping incomplete row", 
				logger.Int("line", lineNum),
				logger.Int("columns", len(row)),
			)
			lineNum++
			continue
		}

		record := UserImportRecord{
			Username: strings.TrimSpace(row[0]),
			Email:    strings.TrimSpace(row[1]),
			Password: strings.TrimSpace(row[2]),
			Role:     strings.TrimSpace(row[3]),
			LineNum:  lineNum,
		}

		// Basic validation
		if record.Username == "" || record.Email == "" || record.Password == "" {
			s.logger.Warn("Skipping row with empty required fields", logger.Int("line", lineNum))
			lineNum++
			continue
		}

		records = append(records, record)
		lineNum++
	}

	return records, nil
}

// validateHeader checks if CSV header matches expected format
func (s *ImportService) validateHeader(header, expected []string) bool {
	if len(header) < len(expected) {
		return false
	}
	
	for i, expectedCol := range expected {
		if strings.ToLower(strings.TrimSpace(header[i])) != expectedCol {
			return false
		}
	}
	return true
}

// worker processes user import records concurrently
func (s *ImportService) worker(ctx context.Context, workerID int, recordChan <-chan UserImportRecord, resultChan chan<- ImportResult, wg *sync.WaitGroup) {
	defer wg.Done()
	
	s.logger.Debug("Worker started", logger.Int("worker_id", workerID))
	
	for {
		select {
		case record, ok := <-recordChan:
			if !ok {
				s.logger.Debug("Worker finished - channel closed", logger.Int("worker_id", workerID))
				return
			}
			
			result := s.processUserRecord(ctx, record, workerID)
			
			select {
			case resultChan <- result:
			case <-ctx.Done():
				s.logger.Warn("Context cancelled while sending result", logger.Int("worker_id", workerID))
				return
			}
			
		case <-ctx.Done():
			s.logger.Warn("Worker cancelled by context", logger.Int("worker_id", workerID))
			return
		}
	}
}

// processUserRecord processes a single user record
func (s *ImportService) processUserRecord(ctx context.Context, record UserImportRecord, workerID int) ImportResult {
	s.logger.Debug("Processing user record",
		logger.Int("worker_id", workerID),
		logger.Int("line", record.LineNum),
		logger.String("username", record.Username),
		logger.String("email", record.Email),
	)

	// Validate role
	var role models.UserRole
	switch strings.ToLower(record.Role) {
	case "manager":
		role = models.RoleManager
	case "member":
		role = models.RoleMember
	default:
		return ImportResult{
			Record:  record,
			Success: false,
			Error:   fmt.Sprintf("invalid role '%s'. Must be 'manager' or 'member'", record.Role),
		}
	}

	// Create user input
	input := &CreateUserInput{
		Username: record.Username,
		Email:    record.Email,
		Password: record.Password,
		Role:     role,
	}

	// Create user via GraphQL mutation (through service)
	user, err := s.userService.CreateUser(input)
	if err != nil {
		s.logger.Error("Failed to create user",
			logger.Int("worker_id", workerID),
			logger.Int("line", record.LineNum),
			logger.String("email", record.Email),
			logger.Error(err),
		)
		
		return ImportResult{
			Record:  record,
			Success: false,
			Error:   err.Error(),
		}
	}

	s.logger.Debug("User created successfully",
		logger.Int("worker_id", workerID),
		logger.Int("line", record.LineNum),
		logger.String("user_id", user.ID.String()),
		logger.String("email", user.Email),
	)

	return ImportResult{
		Record:  record,
		Success: true,
		UserID:  user.ID.String(),
	}
}
