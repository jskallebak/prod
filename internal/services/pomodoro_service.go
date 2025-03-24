package services

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jskallebak/prod/internal/db/sqlc"
)

// PomodoroStatus represents the status of a Pomodoro session
type PomodoroStatus string

const (
	// StatusActive indicates an active Pomodoro session
	StatusActive PomodoroStatus = "active"
	// StatusPaused indicates a paused Pomodoro session
	StatusPaused PomodoroStatus = "paused"
	// StatusCompleted indicates a completed Pomodoro session
	StatusCompleted PomodoroStatus = "completed"
	// StatusCancelled indicates a cancelled Pomodoro session
	StatusCancelled PomodoroStatus = "cancelled"
)

// PomodoroService handles business logic for Pomodoro sessions
type PomodoroService struct {
	queries *sqlc.Queries
}

// NewPomodoroService creates a new PomodoroService
func NewPomodoroService(queries *sqlc.Queries) *PomodoroService {
	return &PomodoroService{
		queries: queries,
	}
}

// PomodoroSession represents a Pomodoro session
type PomodoroSession struct {
	ID                 int32
	UserID             int32
	TaskID             *int32
	Status             PomodoroStatus
	WorkDuration       time.Duration
	BreakDuration      time.Duration
	StartTime          pgtype.Timestamptz
	EndTime            pgtype.Timestamptz
	PauseTime          pgtype.Timestamptz
	TotalPauseDuration time.Duration
	ActualWorkDuration time.Duration
	Note               string
	CreatedAt          pgtype.Timestamptz
	UpdatedAt          pgtype.Timestamptz
}

// PomodoroConfig represents user configuration for Pomodoro sessions
type PomodoroConfig struct {
	UserID             int32
	WorkDuration       int32
	BreakDuration      int32
	LongBreakDuration  int32
	LongBreakInterval  int32
	AutoStartBreaks    bool
	AutoStartPomodoros bool
	CreatedAt          pgtype.Timestamptz
	UpdatedAt          pgtype.Timestamptz
}

// PomodoroReport represents statistics and data for a Pomodoro report
type PomodoroReport struct {
	TotalSessions         int
	CompletedSessions     int
	CancelledSessions     int
	TotalTimeSeconds      int64
	WorkTimeSeconds       int64
	BreakTimeSeconds      int64
	PauseTimeSeconds      int64
	AvgSessionSeconds     int64
	AvgWorkSessionSeconds int64
	AvgBreakSeconds       int64
	DailyStats            []DailyStat
	TopTasks              []TaskStat
}

// DailyStat represents Pomodoro statistics for a single day
type DailyStat struct {
	Date              time.Time
	TotalSessions     int
	CompletedSessions int
	WorkTimeSeconds   int64
}

// TaskStat represents statistics for a single task
type TaskStat struct {
	ID               int32
	Description      string
	SessionCount     int
	TotalTimeSeconds int64
}

// StartSession starts a new Pomodoro session
func (s *PomodoroService) StartSession(
	ctx context.Context,
	userID int32,
	taskID *int32,
	workDuration time.Duration,
	breakDuration time.Duration,
	note string,
) (*PomodoroSession, error) {
	// Check if there's already an active session
	activeSession, err := s.GetActiveSession(ctx, userID)
	if err == nil && activeSession != nil {
		return nil, fmt.Errorf("user already has an active Pomodoro session")
	}

	// Create a new session
	params := sqlc.CreatePomodoroSessionParams{
		UserID: pgtype.Int4{
			Int32: userID,
			Valid: true,
		},
		Status:        string(StatusActive),
		WorkDuration:  int32(workDuration.Minutes()),
		BreakDuration: int32(breakDuration.Minutes()),
		StartTime: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
	}

	// Set optional fields
	if taskID != nil {
		params.TaskID = pgtype.Int4{
			Int32: *taskID,
			Valid: true,
		}
	}

	if note != "" {
		params.Note = pgtype.Text{
			String: note,
			Valid:  true,
		}
	}

	// Call data layer
	session, err := s.queries.CreatePomodoroSession(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create Pomodoro session: %w", err)
	}

	// Convert to service model
	pomodoroSession := &PomodoroSession{
		ID:            session.ID,
		UserID:        session.UserID.Int32,
		Status:        PomodoroStatus(session.Status),
		WorkDuration:  time.Duration(session.WorkDuration) * time.Minute,
		BreakDuration: time.Duration(session.BreakDuration) * time.Minute,
		StartTime:     session.StartTime,
		CreatedAt:     session.CreatedAt,
		Note:          session.Note.String,
	}

	if session.TaskID.Valid {
		taskID := session.TaskID.Int32
		pomodoroSession.TaskID = &taskID
	}

	return pomodoroSession, nil
}

// GetActiveSession retrieves the active Pomodoro session for a user if one exists
func (s *PomodoroService) GetActiveSession(ctx context.Context, userID int32) (*PomodoroSession, error) {
	session, err := s.queries.GetActivePomodoroSession(ctx, pgtype.Int4{
		Int32: userID,
		Valid: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get active session: %w", err)
	}

	// Convert to service model
	pomodoroSession := &PomodoroSession{
		ID:            session.ID,
		UserID:        session.UserID.Int32,
		Status:        PomodoroStatus(session.Status),
		WorkDuration:  time.Duration(session.WorkDuration) * time.Minute,
		BreakDuration: time.Duration(session.BreakDuration) * time.Minute,
		StartTime:     session.StartTime,
		CreatedAt:     session.CreatedAt,
		Note:          session.Note.String,
	}

	if session.TaskID.Valid {
		taskID := session.TaskID.Int32
		pomodoroSession.TaskID = &taskID
	}

	if session.EndTime.Valid {
		pomodoroSession.EndTime = session.EndTime
	}

	if session.PauseTime.Valid {
		pomodoroSession.PauseTime = session.PauseTime
	}

	return pomodoroSession, nil
}

// StopSession stops an active Pomodoro session
func (s *PomodoroService) StopSession(ctx context.Context, userID int32, complete bool) (*PomodoroSession, error) {
	// Get active session
	activeSession, err := s.GetActiveSession(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("no active Pomodoro session found: %w", err)
	}

	// Determine status based on completion
	status := string(StatusCancelled)
	if complete {
		status = string(StatusCompleted)
	}

	// Update the session
	params := sqlc.StopPomodoroSessionParams{
		ID: activeSession.ID,
		UserID: pgtype.Int4{
			Int32: userID,
			Valid: true,
		},
		Status: status,
		EndTime: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
	}

	// Call data layer
	session, err := s.queries.StopPomodoroSession(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to stop Pomodoro session: %w", err)
	}

	// Convert to service model
	pomodoroSession := &PomodoroSession{
		ID:            session.ID,
		UserID:        session.UserID.Int32,
		Status:        PomodoroStatus(session.Status),
		WorkDuration:  time.Duration(session.WorkDuration) * time.Minute,
		BreakDuration: time.Duration(session.BreakDuration) * time.Minute,
		StartTime:     session.StartTime,
		EndTime:       session.EndTime,
		CreatedAt:     session.CreatedAt,
		Note:          session.Note.String,
	}

	if session.TaskID.Valid {
		taskID := session.TaskID.Int32
		pomodoroSession.TaskID = &taskID
	}

	return pomodoroSession, nil
}

// PauseSession pauses an active Pomodoro session
func (s *PomodoroService) PauseSession(ctx context.Context, userID int32) (*PomodoroSession, error) {
	// Get active session
	activeSession, err := s.GetActiveSession(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("no active Pomodoro session found: %w", err)
	}

	// Ensure session is not already paused
	if activeSession.Status == StatusPaused {
		return nil, fmt.Errorf("session is already paused")
	}

	// Update the session
	params := sqlc.PausePomodoroSessionParams{
		ID: activeSession.ID,
		UserID: pgtype.Int4{
			Int32: userID,
			Valid: true,
		},
		Status: string(StatusPaused),
		PauseTime: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
	}

	// Call data layer
	session, err := s.queries.PausePomodoroSession(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to pause Pomodoro session: %w", err)
	}

	// Convert to service model
	pomodoroSession := &PomodoroSession{
		ID:            session.ID,
		UserID:        session.UserID.Int32,
		Status:        PomodoroStatus(session.Status),
		WorkDuration:  time.Duration(session.WorkDuration) * time.Minute,
		BreakDuration: time.Duration(session.BreakDuration) * time.Minute,
		StartTime:     session.StartTime,
		PauseTime:     session.PauseTime,
		CreatedAt:     session.CreatedAt,
		Note:          session.Note.String,
	}

	if session.TaskID.Valid {
		taskID := session.TaskID.Int32
		pomodoroSession.TaskID = &taskID
	}

	return pomodoroSession, nil
}

// ResumeSession resumes a paused Pomodoro session
func (s *PomodoroService) ResumeSession(ctx context.Context, userID int32) (*PomodoroSession, error) {
	// Get active session
	activeSession, err := s.GetActiveSession(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("no active Pomodoro session found: %w", err)
	}

	// Ensure session is paused
	if activeSession.Status != StatusPaused {
		return nil, fmt.Errorf("session is not paused")
	}

	// Calculate pause duration
	now := time.Now()
	pauseDuration := now.Sub(activeSession.PauseTime.Time)

	// Update the session
	params := sqlc.ResumePomodoroSessionParams{
		ID: activeSession.ID,
		UserID: pgtype.Int4{
			Int32: userID,
			Valid: true,
		},
		Status: string(StatusActive),
		TotalPauseDuration: pgtype.Int4{
			Int32: int32(pauseDuration.Seconds()),
			Valid: true,
		},
	}

	// Call data layer
	session, err := s.queries.ResumePomodoroSession(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to resume Pomodoro session: %w", err)
	}

	// Convert to service model
	pomodoroSession := &PomodoroSession{
		ID:                 session.ID,
		UserID:             session.UserID.Int32,
		Status:             PomodoroStatus(session.Status),
		WorkDuration:       time.Duration(session.WorkDuration) * time.Minute,
		BreakDuration:      time.Duration(session.BreakDuration) * time.Minute,
		StartTime:          session.StartTime,
		TotalPauseDuration: time.Duration(session.TotalPauseDuration.Int32) * time.Second,
		CreatedAt:          session.CreatedAt,
		Note:               session.Note.String,
	}

	if session.TaskID.Valid {
		taskID := session.TaskID.Int32
		pomodoroSession.TaskID = &taskID
	}

	return pomodoroSession, nil
}

// AttachTask attaches a task to an active Pomodoro session
func (s *PomodoroService) AttachTask(ctx context.Context, userID int32, taskID int32) (*PomodoroSession, error) {
	// Get active session
	activeSession, err := s.GetActiveSession(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("no active Pomodoro session found: %w", err)
	}

	// Update the session
	params := sqlc.AttachTaskToPomodoroParams{
		ID: activeSession.ID,
		UserID: pgtype.Int4{
			Int32: userID,
			Valid: true,
		},
		TaskID: pgtype.Int4{
			Int32: taskID,
			Valid: true,
		},
	}

	// Call data layer
	session, err := s.queries.AttachTaskToPomodoro(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to attach task to Pomodoro session: %w", err)
	}

	// Convert to service model
	pomodoroSession := &PomodoroSession{
		ID:            session.ID,
		UserID:        session.UserID.Int32,
		Status:        PomodoroStatus(session.Status),
		WorkDuration:  time.Duration(session.WorkDuration) * time.Minute,
		BreakDuration: time.Duration(session.BreakDuration) * time.Minute,
		StartTime:     session.StartTime,
		CreatedAt:     session.CreatedAt,
		Note:          session.Note.String,
	}

	if session.TaskID.Valid {
		taskID := session.TaskID.Int32
		pomodoroSession.TaskID = &taskID
	}

	if session.EndTime.Valid {
		pomodoroSession.EndTime = session.EndTime
	}

	if session.PauseTime.Valid {
		pomodoroSession.PauseTime = session.PauseTime
	}

	return pomodoroSession, nil
}

// DetachTask removes a task attachment from an active Pomodoro session
func (s *PomodoroService) DetachTask(ctx context.Context, userID int32) (*PomodoroSession, error) {
	// Get active session
	activeSession, err := s.GetActiveSession(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("no active Pomodoro session found: %w", err)
	}

	// Ensure session has a task attached
	if activeSession.TaskID == nil {
		return nil, fmt.Errorf("no task attached to current session")
	}

	// Update the session
	params := sqlc.DetachTaskFromPomodoroParams{
		ID: activeSession.ID,
		UserID: pgtype.Int4{
			Int32: userID,
			Valid: true,
		},
	}

	// Call data layer
	session, err := s.queries.DetachTaskFromPomodoro(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to detach task from Pomodoro session: %w", err)
	}

	// Convert to service model
	pomodoroSession := &PomodoroSession{
		ID:            session.ID,
		UserID:        session.UserID.Int32,
		Status:        PomodoroStatus(session.Status),
		WorkDuration:  time.Duration(session.WorkDuration) * time.Minute,
		BreakDuration: time.Duration(session.BreakDuration) * time.Minute,
		StartTime:     session.StartTime,
		CreatedAt:     session.CreatedAt,
		Note:          session.Note.String,
	}

	if session.EndTime.Valid {
		pomodoroSession.EndTime = session.EndTime
	}

	if session.PauseTime.Valid {
		pomodoroSession.PauseTime = session.PauseTime
	}

	return pomodoroSession, nil
}

// GetUserConfig gets the user's Pomodoro configuration
func (s *PomodoroService) GetUserConfig(ctx context.Context, userID int32) (*PomodoroConfig, error) {
	config, err := s.queries.GetPomodoroConfig(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get Pomodoro config: %w", err)
	}

	// Convert to service model
	pomodoroConfig := &PomodoroConfig{
		UserID:             config.UserID,
		WorkDuration:       config.WorkDuration,
		BreakDuration:      config.BreakDuration,
		LongBreakDuration:  config.LongBreakDuration,
		LongBreakInterval:  config.LongBreakInterval,
		AutoStartBreaks:    config.AutoStartBreaks,
		AutoStartPomodoros: config.AutoStartPomodoros,
		CreatedAt:          pgtype.Timestamptz{Time: config.CreatedAt, Valid: true},
		UpdatedAt:          pgtype.Timestamptz{Time: config.UpdatedAt, Valid: true},
	}

	return pomodoroConfig, nil
}

// UpdateUserConfig updates the user's Pomodoro configuration
func (s *PomodoroService) UpdateUserConfig(
	ctx context.Context,
	userID int32,
	workDuration int32,
	breakDuration int32,
	longBreakDuration int32,
	longBreakInterval int32,
	autoStartBreaks bool,
	autoStartPomodoros bool,
) (*PomodoroConfig, error) {
	params := sqlc.UpsertPomodoroConfigParams{
		UserID:             userID,
		WorkDuration:       workDuration,
		BreakDuration:      breakDuration,
		LongBreakDuration:  longBreakDuration,
		LongBreakInterval:  longBreakInterval,
		AutoStartBreaks:    autoStartBreaks,
		AutoStartPomodoros: autoStartPomodoros,
	}

	config, err := s.queries.UpsertPomodoroConfig(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update Pomodoro config: %w", err)
	}

	// Convert to service model
	pomodoroConfig := &PomodoroConfig{
		UserID:             config.UserID,
		WorkDuration:       config.WorkDuration,
		BreakDuration:      config.BreakDuration,
		LongBreakDuration:  config.LongBreakDuration,
		LongBreakInterval:  config.LongBreakInterval,
		AutoStartBreaks:    config.AutoStartBreaks,
		AutoStartPomodoros: config.AutoStartPomodoros,
		CreatedAt:          pgtype.Timestamptz{Time: config.CreatedAt, Valid: true},
		UpdatedAt:          pgtype.Timestamptz{Time: config.UpdatedAt, Valid: true},
	}

	return pomodoroConfig, nil
}

// ListSessions lists completed Pomodoro sessions with optional filtering
func (s *PomodoroService) ListSessions(
	ctx context.Context,
	userID int32,
	taskID *int32,
	startDate *time.Time,
	endDate *time.Time,
	status string,
	limit int32,
) ([]PomodoroSession, error) {
	params := sqlc.ListPomodoroSessionsParams{
		UserID: pgtype.Int4{
			Int32: userID,
			Valid: true,
		},
		Limit: limit,
	}

	if taskID != nil {
		params.TaskID = pgtype.Int4{
			Int32: *taskID,
			Valid: true,
		}
	}

	if startDate != nil {
		params.StartTime = pgtype.Timestamptz{
			Time:  *startDate,
			Valid: true,
		}
	}

	if endDate != nil {
		params.StartTime_2 = pgtype.Timestamptz{
			Time:  *endDate,
			Valid: true,
		}
	}

	sessions, err := s.queries.ListPomodoroSessions(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list Pomodoro sessions: %w", err)
	}

	// Filter by status if provided (do this in memory since we don't have a SQL parameter for it)
	var filteredSessions []sqlc.PomodoroSession
	if status != "" {
		for _, session := range sessions {
			if session.Status == status {
				filteredSessions = append(filteredSessions, session)
			}
		}
	} else {
		filteredSessions = sessions
	}

	// Convert to service model
	pomodoroSessions := make([]PomodoroSession, len(filteredSessions))
	for i, session := range filteredSessions {
		pomodoroSessions[i] = PomodoroSession{
			ID:            session.ID,
			UserID:        session.UserID.Int32,
			Status:        PomodoroStatus(session.Status),
			WorkDuration:  time.Duration(session.WorkDuration) * time.Minute,
			BreakDuration: time.Duration(session.BreakDuration) * time.Minute,
			StartTime:     session.StartTime,
			CreatedAt:     session.CreatedAt,
			Note:          session.Note.String,
		}

		if session.TaskID.Valid {
			taskID := session.TaskID.Int32
			pomodoroSessions[i].TaskID = &taskID
		}

		if session.EndTime.Valid {
			pomodoroSessions[i].EndTime = session.EndTime
		}

		if session.PauseTime.Valid {
			pomodoroSessions[i].PauseTime = session.PauseTime
		}
	}

	return pomodoroSessions, nil
}

// GetSessionStats gets statistics for Pomodoro sessions
func (s *PomodoroService) GetSessionStats(
	ctx context.Context,
	userID int32,
	taskID *int32,
	startDate *time.Time,
	endDate *time.Time,
) (map[string]interface{}, error) {
	params := sqlc.GetPomodoroStatsParams{
		UserID: pgtype.Int4{
			Int32: userID,
			Valid: true,
		},
	}

	if taskID != nil {
		params.TaskID = pgtype.Int4{
			Int32: *taskID,
			Valid: true,
		}
	}

	if startDate != nil {
		params.StartTime = pgtype.Timestamptz{
			Time:  *startDate,
			Valid: true,
		}
	}

	if endDate != nil {
		params.StartTime_2 = pgtype.Timestamptz{
			Time:  *endDate,
			Valid: true,
		}
	}

	stats, err := s.queries.GetPomodoroStats(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get Pomodoro stats: %w", err)
	}

	// Convert to map for easier consumption by clients
	result := map[string]interface{}{
		"total_sessions":      stats.TotalSessions,
		"completed_sessions":  stats.CompletedSessions,
		"cancelled_sessions":  stats.CancelledSessions,
		"total_work_mins":     stats.TotalWorkMins,
		"total_break_mins":    stats.TotalBreakMins,
		"total_duration_mins": stats.TotalDurationMins,
		"avg_duration_mins":   stats.AvgDurationMins,
	}

	if stats.MostProductiveDay.Valid {
		result["most_productive_day"] = stats.MostProductiveDay.Time
	}

	if stats.MostProductiveHour >= 0 {
		result["most_productive_hour"] = stats.MostProductiveHour
	}

	return result, nil
}

// GenerateReport generates a report of Pomodoro usage
func (s *PomodoroService) GenerateReport(ctx context.Context, userID int32, taskID *int32, startDate, endDate *time.Time) (*PomodoroReport, error) {
	// Get all sessions within the period
	sessions, err := s.ListSessions(ctx, userID, taskID, startDate, endDate, "", 1000)
	if err != nil {
		return nil, fmt.Errorf("error getting sessions for report: %w", err)
	}

	// Initialize the report
	report := &PomodoroReport{
		TotalSessions: len(sessions),
	}

	// Early return if no sessions
	if len(sessions) == 0 {
		return report, nil
	}

	// Keep track of daily stats
	dailyStats := make(map[string]*DailyStat)

	// Keep track of task stats
	taskStats := make(map[int32]*TaskStat)

	// Process each session
	for _, session := range sessions {
		// Count by status
		if session.Status == StatusCompleted {
			report.CompletedSessions++
		} else if session.Status == StatusCancelled {
			report.CancelledSessions++
		}

		// Calculate session time
		var sessionDuration time.Duration
		if session.EndTime.Valid {
			sessionDuration = session.EndTime.Time.Sub(session.StartTime.Time)
		} else {
			// For active sessions, use current time
			sessionDuration = time.Now().Sub(session.StartTime.Time)
		}

		// Subtract pause time
		sessionDuration -= session.TotalPauseDuration

		// Add to total time
		totalSeconds := int64(sessionDuration.Seconds())
		report.TotalTimeSeconds += totalSeconds

		// Add to work time
		report.WorkTimeSeconds += int64(session.WorkDuration.Seconds())

		// Break time is the difference
		breakSeconds := totalSeconds - int64(session.WorkDuration.Seconds())
		if breakSeconds > 0 {
			report.BreakTimeSeconds += breakSeconds
		}

		// Add pause time
		report.PauseTimeSeconds += int64(session.TotalPauseDuration.Seconds())

		// Track daily stats
		dateKey := session.StartTime.Time.Format("2006-01-02")
		daily, exists := dailyStats[dateKey]
		if !exists {
			daily = &DailyStat{
				Date:          session.StartTime.Time,
				TotalSessions: 0,
			}
			dailyStats[dateKey] = daily
		}

		daily.TotalSessions++
		if session.Status == StatusCompleted {
			daily.CompletedSessions++
		}
		daily.WorkTimeSeconds += int64(session.WorkDuration.Seconds())

		// Track task stats if a task is attached
		if session.TaskID != nil {
			task, exists := taskStats[*session.TaskID]
			if !exists {
				// Get task information
				taskService := NewTaskService(s.queries)
				taskInfo, err := taskService.GetTask(ctx, *session.TaskID, userID)
				if err == nil {
					task = &TaskStat{
						ID:           taskInfo.ID,
						Description:  taskInfo.Description,
						SessionCount: 0,
					}
					taskStats[*session.TaskID] = task
				}
			}

			if task != nil {
				task.SessionCount++
				task.TotalTimeSeconds += int64(session.WorkDuration.Seconds())
			}
		}
	}

	// Calculate averages
	if report.TotalSessions > 0 {
		report.AvgSessionSeconds = report.TotalTimeSeconds / int64(report.TotalSessions)
	}

	if report.CompletedSessions > 0 {
		report.AvgWorkSessionSeconds = report.WorkTimeSeconds / int64(report.CompletedSessions)
	}

	if report.CompletedSessions > 0 {
		report.AvgBreakSeconds = report.BreakTimeSeconds / int64(report.CompletedSessions)
	}

	// Convert daily stats map to slice
	for _, stat := range dailyStats {
		report.DailyStats = append(report.DailyStats, *stat)
	}

	// Sort daily stats by date
	// (In a real implementation, we would sort here)

	// Convert task stats map to slice
	for _, stat := range taskStats {
		report.TopTasks = append(report.TopTasks, *stat)
	}

	// Sort task stats by session count (descending)
	// (In a real implementation, we would sort here)

	return report, nil
}
