package habits

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/events"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/infrastructure/cache"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var (
	ErrInvalidTransition = errors.New("invalid status transition")
	ErrDependencyFailed  = errors.New("dependencies not completed")
)

type Service interface {
	CreateHabit(ctx context.Context, input CreateHabitInput) (*Habit, error)
	GetHabit(ctx context.Context, id uuid.UUID) (*Habit, error)
	ListHabits(ctx context.Context, filter HabitFilter) ([]Habit, int64, error)
	UpdateHabit(ctx context.Context, id uuid.UUID, input UpdateHabitInput) (*Habit, error)
	DeleteHabit(ctx context.Context, id uuid.UUID) error
	MarkCompleted(ctx context.Context, id uuid.UUID, userID uuid.UUID, completionDate *time.Time) error
	UnmarkCompleted(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	ResetDailyCompletions(ctx context.Context) (int64, error)
	CheckAndResetBrokenStreaks(ctx context.Context) (int64, error)
	GetTopStreaks(ctx context.Context, userID uuid.UUID, limit int) ([]Habit, error)
	GetStreakHistory(ctx context.Context, id uuid.UUID) ([]StreakHistory, error)
	GetHabitsDueToday(ctx context.Context, userID uuid.UUID) ([]Habit, error)

	// Heatmap related methods
	LogHabitCompletion(ctx context.Context, habitID uuid.UUID, userID uuid.UUID, date time.Time) error
	GetHeatmapData(ctx context.Context, userID uuid.UUID, period string) (map[string]int, error)

	// Notification related methods
	SendHabitReminders(ctx context.Context) error

	// Analytics methods
	RecordHabitActivity(ctx context.Context, input RecordHabitActivityInput) error
	GetHabitAnalytics(ctx context.Context, habitID uuid.UUID, startTime, endTime time.Time, page, pageSize int) ([]HabitAnalytics, int64, error)
	GetUserHabitAnalytics(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time, page, pageSize int) ([]HabitAnalytics, int64, error)
	GetHabitActivitySummary(ctx context.Context, habitID uuid.UUID, startTime, endTime time.Time) (*HabitActivitySummary, error)
	GetUserHabitActivitySummary(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time) (*UserHabitActivitySummary, error)
	GetDashboardMetrics(userID uuid.UUID) (HabitsDashboardMetrics, error)
}

type service struct {
	repo      Repository
	notifySvc *HabitNotificationService
	redis     *cache.RedisClient
	logger    *zap.Logger
}

func NewService(repo Repository, notifySvc *HabitNotificationService, redis *cache.RedisClient, logger *zap.Logger) Service {
	return &service{
		repo:      repo,
		notifySvc: notifySvc,
		redis:     redis,
		logger:    logger,
	}
}

func (s *service) CreateHabit(ctx context.Context, input CreateHabitInput) (*Habit, error) {
	habit := &Habit{
		ID:          uuid.New(),
		UserID:      input.UserID,
		Title:       input.Title,
		Description: input.Description,
		StartDay:    input.StartDay,
		EndDay:      input.EndDay,
	}

	err := s.repo.Create(ctx, habit)
	if err != nil {
		return nil, err
	}

	s.recordHabitActivity(ctx, habit, habit.UserID, "habit_created", map[string]interface{}{
		"title": habit.Title,
	})

	// Publish dashboard event
	event := &events.DashboardEvent{
		EventType: events.DashboardEventCacheInvalidate,
		UserID:    input.UserID,
		Timestamp: time.Now().UTC(),
		Details: map[string]interface{}{
			"action":    "habit_created",
			"habit_id":  habit.ID,
			"title":     habit.Title,
			"start_day": habit.StartDay.Format(time.RFC3339),
		},
	}
	if err := s.redis.PublishDashboardEvent(ctx, event); err != nil {
		s.logger.Error("Failed to publish dashboard event", zap.Error(err))
	}

	return habit, nil
}

// Helper to record habit creation
func (s *service) recordHabitCreation(ctx context.Context, habit *Habit) {
	s.RecordHabitActivity(ctx, RecordHabitActivityInput{
		HabitID: habit.ID,
		UserID:  habit.UserID,
		Action:  ActionHabitCreated,
		Metadata: map[string]interface{}{
			"title":       habit.Title,
			"description": habit.Description,
			"start_day":   habit.StartDay.Format(time.RFC3339),
		},
		Timestamp: time.Now(),
	})
}

func (s *service) GetHabit(ctx context.Context, id uuid.UUID) (*Habit, error) {
	habit, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if habit == nil {
		return nil, ErrHabitNotFound
	}
	return habit, nil
}

func (s *service) ListHabits(ctx context.Context, filter HabitFilter) ([]Habit, int64, error) {
	return s.repo.FindAll(ctx, filter)
}

func (s *service) UpdateHabit(ctx context.Context, id uuid.UUID, input UpdateHabitInput) (*Habit, error) {
	habit, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if habit == nil {
		return nil, ErrHabitNotFound
	}

	// Track if anything changed
	changed := false

	if input.Title != nil {
		if habit.Title != *input.Title {
			habit.Title = *input.Title
			changed = true
		}
	}
	if input.Description != nil {
		if habit.Description != *input.Description {
			habit.Description = *input.Description
			changed = true
		}
	}
	if input.StartDay != nil {
		if !habit.StartDay.Equal(*input.StartDay) {
			habit.StartDay = *input.StartDay
			changed = true
		}
	}
	if input.EndDay != nil {
		if (habit.EndDay == nil) || !habit.EndDay.Equal(*input.EndDay) {
			habit.EndDay = input.EndDay
			changed = true
		}
	}

	if !changed {
		return habit, nil
	}

	err = s.repo.Update(ctx, habit)
	if err != nil {
		return nil, err
	}
	s.recordHabitActivity(ctx, habit, habit.UserID, "habit_updated", map[string]interface{}{
		"title": habit.Title,
	})
	return habit, nil
}

// Helper to record habit update
func (s *service) recordHabitUpdate(ctx context.Context, habit *Habit, originalTitle, originalDesc string,
	originalStartDay time.Time, originalEndDay *time.Time) {

	metadata := map[string]interface{}{
		"habit_id":             habit.ID.String(),
		"new_title":            habit.Title,
		"original_title":       originalTitle,
		"new_description":      habit.Description,
		"original_description": originalDesc,
		"new_start_day":        habit.StartDay.Format(time.RFC3339),
		"original_start_day":   originalStartDay.Format(time.RFC3339),
	}

	if habit.EndDay != nil {
		metadata["new_end_day"] = habit.EndDay.Format(time.RFC3339)
	}
	if originalEndDay != nil {
		metadata["original_end_day"] = originalEndDay.Format(time.RFC3339)
	}

	s.RecordHabitActivity(ctx, RecordHabitActivityInput{
		HabitID:   habit.ID,
		UserID:    habit.UserID,
		Action:    ActionHabitUpdated,
		Metadata:  metadata,
		Timestamp: time.Now(),
	})
}

func (s *service) DeleteHabit(ctx context.Context, id uuid.UUID) error {
	habit, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if habit == nil {
		return ErrHabitNotFound
	}

	s.recordHabitActivity(ctx, habit, habit.UserID, "habit_deleted", map[string]interface{}{
		"title": habit.Title,
	})

	return s.repo.Delete(ctx, id)
}

// Helper to record habit deletion
func (s *service) recordHabitDeletion(ctx context.Context, habit *Habit) {
	s.RecordHabitActivity(ctx, RecordHabitActivityInput{
		HabitID: habit.ID,
		UserID:  habit.UserID,
		Action:  ActionHabitDeleted,
		Metadata: map[string]interface{}{
			"title":          habit.Title,
			"description":    habit.Description,
			"current_streak": habit.CurrentStreak,
			"longest_streak": habit.LongestStreak,
		},
		Timestamp: time.Now(),
	})
}

func (s *service) MarkCompleted(ctx context.Context, id uuid.UUID, userID uuid.UUID, completionDate *time.Time) error {
	habit, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if habit == nil {
		return ErrHabitNotFound
	}

	if err := s.repo.MarkCompleted(ctx, id, userID, completionDate); err != nil {
		return err
	}

	// Update streak quality after marking completed
	if err := s.repo.UpdateStreakQuality(ctx, id); err != nil {
		log.Printf("failed to update streak quality for habit %s: %v", id, err)
	}

	// Log the habit completion for heatmap
	completionTime := time.Now()
	if completionDate != nil {
		completionTime = *completionDate
	}

	if err := s.repo.LogHabitCompletion(ctx, id, userID, completionTime); err != nil {
		log.Printf("failed to log habit completion for heatmap: %v", err)
	}

	// Get updated habit with new streak information
	updatedHabit, err := s.repo.FindByID(ctx, id)
	if err != nil {
		log.Printf("failed to fetch updated habit data: %v", err)
		return nil
	}

	// Record habit completion activity
	s.recordHabitCompletion(ctx, updatedHabit, completionTime)

	// Invalidate dashboard cache for this user
	s.recordHabitActivity(ctx, updatedHabit, userID, "habit_completed", nil)

	// Check if this completion created a streak milestone
	if updatedHabit.CurrentStreak > 0 && (updatedHabit.CurrentStreak == 7 ||
		updatedHabit.CurrentStreak == 30 || updatedHabit.CurrentStreak == 100 ||
		updatedHabit.CurrentStreak == 365) {
		s.recordStreakMilestone(ctx, updatedHabit)
	}

	// Send habit completion notification
	if s.notifySvc != nil {
		if err := s.notifySvc.NotifyHabitCompleted(ctx, userID, updatedHabit); err != nil {
			log.Printf("failed to send habit completion notification: %v", err)
		}

		// Check if we should send a streak notification
		if s.notifySvc.ShouldSendStreakNotification(updatedHabit.CurrentStreak) {
			if err := s.notifySvc.NotifyHabitStreak(ctx, userID, updatedHabit); err != nil {
				log.Printf("failed to send habit streak notification: %v", err)
			}
		}
	}

	// After successful completion, publish event
	event := &events.DashboardEvent{
		EventType: events.DashboardEventCacheInvalidate,
		UserID:    userID,
		Timestamp: time.Now().UTC(),
		Details: map[string]interface{}{
			"action":          "habit_completed",
			"habit_id":        id,
			"completion_time": completionTime.Format(time.RFC3339),
		},
	}
	if err := s.redis.PublishDashboardEvent(ctx, event); err != nil {
		s.logger.Error("Failed to publish dashboard event", zap.Error(err))
	}

	return nil
}

// Helper to record habit completion
func (s *service) recordHabitCompletion(ctx context.Context, habit *Habit, completionTime time.Time) {
	metadata := map[string]interface{}{
		"title":           habit.Title,
		"current_streak":  habit.CurrentStreak,
		"completion_time": completionTime.Format(time.RFC3339),
	}

	if habit.LastCompletedDate != nil {
		metadata["last_completed_date"] = habit.LastCompletedDate.Format(time.RFC3339)
	}

	s.RecordHabitActivity(ctx, RecordHabitActivityInput{
		HabitID:   habit.ID,
		UserID:    habit.UserID,
		Action:    ActionHabitCompleted,
		Metadata:  metadata,
		Timestamp: time.Now(),
	})
}

// Helper to record streak milestone
func (s *service) recordStreakMilestone(ctx context.Context, habit *Habit) {
	s.RecordHabitActivity(ctx, RecordHabitActivityInput{
		HabitID: habit.ID,
		UserID:  habit.UserID,
		Action:  ActionStreakMilestone,
		Metadata: map[string]interface{}{
			"title":       habit.Title,
			"streak_days": habit.CurrentStreak,
			"milestone":   fmt.Sprintf("%d-day streak", habit.CurrentStreak),
		},
		Timestamp: time.Now(),
	})
}

func (s *service) UnmarkCompleted(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	habit, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if habit == nil {
		return ErrHabitNotFound
	}

	// Store current streak before unmarking
	currentStreak := habit.CurrentStreak

	lastCompletedDate := time.Now()
	if habit.LastCompletedDate != nil {
		lastCompletedDate = *habit.LastCompletedDate
	}

	// First remove the completion log for heatmap
	if err := s.repo.RemoveHabitCompletion(ctx, id, userID, lastCompletedDate); err != nil {
		log.Printf("failed to remove habit completion log: %v", err)
		// Don't return here as we still want to unmark the habit
	}

	// Then unmark the habit as completed
	if err := s.repo.UnmarkCompleted(ctx, id, userID); err != nil {
		return err
	}

	// Update streak quality after unmarking completed
	if err := s.repo.UpdateStreakQuality(ctx, id); err != nil {
		log.Printf("failed to update streak quality for habit %s: %v", id, err)
	}

	// Record habit uncompletion activity
	s.recordHabitUncompletion(ctx, habit, currentStreak)

	// Invalidate dashboard cache for this user
	s.recordHabitActivity(ctx, habit, userID, "habit_uncompleted", nil)

	// Publish dashboard event
	event := &events.DashboardEvent{
		EventType: events.DashboardEventCacheInvalidate,
		UserID:    userID,
		Timestamp: time.Now().UTC(),
		Details: map[string]interface{}{
			"action":   "habit_uncompleted",
			"habit_id": id,
		},
	}
	if err := s.redis.PublishDashboardEvent(ctx, event); err != nil {
		s.logger.Error("Failed to publish dashboard event", zap.Error(err))
	}

	return nil
}

// Helper to record habit uncompletion
func (s *service) recordHabitUncompletion(ctx context.Context, habit *Habit, previousStreak int) {
	s.RecordHabitActivity(ctx, RecordHabitActivityInput{
		HabitID: habit.ID,
		UserID:  habit.UserID,
		Action:  ActionHabitUncompleted,
		Metadata: map[string]interface{}{
			"title":           habit.Title,
			"previous_streak": previousStreak,
			"new_streak":      previousStreak - 1,
		},
		Timestamp: time.Now(),
	})
}

func (s *service) ResetDailyCompletions(ctx context.Context) (int64, error) {
	affected, err := s.repo.ResetDailyCompletions(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to reset daily completions: %w", err)
	}
	return affected, nil
}

func (s *service) CheckAndResetBrokenStreaks(ctx context.Context) (int64, error) {
	// Get habits with active streaks
	activeStreaks, err := s.repo.GetActiveStreaks(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch active streaks: %w", err)
	}

	var totalReset int64
	for _, habit := range activeStreaks {
		// Check if streak is broken using timezone-aware database function
		isBroken, err := s.repo.IsStreakBroken(ctx, habit.LastCompletedDate)
		if err != nil {
			log.Printf("failed to check if streak is broken for habit %s: %v", habit.ID, err)
			continue
		}

		if isBroken {
			lastDate := time.Now()
			if habit.LastCompletedDate != nil {
				lastDate = *habit.LastCompletedDate
			}

			// Store previous streak for notification
			previousStreak := habit.CurrentStreak
			habitCopy := habit

			// Before resetting, store the streak history
			if err := s.repo.LogStreakHistory(ctx, habit.ID, habit.CurrentStreak, lastDate); err != nil {
				log.Printf("failed to log streak history for habit %s: %v", habit.ID, err)
			}

			// Update streak quality after logging history
			if err := s.repo.UpdateStreakQuality(ctx, habit.ID); err != nil {
				log.Printf("failed to update streak quality for habit %s: %v", habit.ID, err)
			}

			// Reset the streak
			if err := s.repo.ResetStreak(ctx, habit.ID); err != nil {
				log.Printf("failed to reset streak for habit %s: %v", habit.ID, err)
				continue
			}

			// Record streak broken analytics event
			s.recordStreakBroken(ctx, &habitCopy, previousStreak)

			// Send streak broken notification if streak was significant
			if s.notifySvc != nil && previousStreak >= 3 {
				if err := s.notifySvc.NotifyHabitStreakBroken(ctx, habit.UserID, &habitCopy, previousStreak); err != nil {
					log.Printf("failed to send habit streak broken notification: %v", err)
				}
			}

			totalReset++
		}
	}

	return totalReset, nil
}

// Helper to record streak broken
func (s *service) recordStreakBroken(ctx context.Context, habit *Habit, previousStreak int) {
	lastCompleted := "unknown"
	if habit.LastCompletedDate != nil {
		lastCompleted = habit.LastCompletedDate.Format(time.RFC3339)
	}

	s.RecordHabitActivity(ctx, RecordHabitActivityInput{
		HabitID: habit.ID,
		UserID:  habit.UserID,
		Action:  ActionStreakBroken,
		Metadata: map[string]interface{}{
			"title":          habit.Title,
			"broken_streak":  previousStreak,
			"last_completed": lastCompleted,
		},
		Timestamp: time.Now(),
	})
}

func (s *service) GetTopStreaks(ctx context.Context, userID uuid.UUID, limit int) ([]Habit, error) {
	// Get habits with additional streak metadata
	habits, err := s.repo.GetTopStreaks(ctx, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch top streaks: %w", err)
	}

	// Enrich habits with additional streak information
	for i := range habits {
		// Get streak history
		history, err := s.repo.GetStreakHistory(ctx, habits[i].ID)
		if err != nil {
			log.Printf("failed to fetch streak history for habit %s: %v", habits[i].ID, err)
			continue
		}

		// Calculate streak quality (consistency)
		streakQuality := calculateStreakQuality(&habits[i], history)
		habits[i].StreakQuality = streakQuality
	}

	// Sort by streak quality if available
	sort.Slice(habits, func(i, j int) bool {
		if habits[i].CurrentStreak == habits[j].CurrentStreak {
			return habits[i].StreakQuality > habits[j].StreakQuality
		}
		return habits[i].CurrentStreak > habits[j].CurrentStreak
	})

	return habits, nil
}

// Helper function to calculate streak quality
func calculateStreakQuality(habit *Habit, history []StreakHistory) float64 {
	if len(history) == 0 {
		return 0
	}

	// Sort history by start date to ensure correct calculation
	sort.Slice(history, func(i, j int) bool {
		return history[i].StartDate.Before(history[j].StartDate)
	})

	// Find earliest and latest dates
	earliest := history[0].StartDate
	latest := history[0].EndDate
	totalCompleted := 0

	for _, h := range history {
		if h.StartDate.Before(earliest) {
			earliest = h.StartDate
		}
		if h.EndDate.After(latest) {
			latest = h.EndDate
		}
		totalCompleted += h.CompletedDays
	}

	// Calculate total days in entire period
	totalDays := int(latest.Sub(earliest).Hours()/24) + 1
	if totalDays == 0 {
		return 0
	}

	quality := float64(totalCompleted) / float64(totalDays)

	// Ensure quality doesn't exceed 1.0
	if quality > 1.0 {
		quality = 1.0
	}

	return quality
}

func (s *service) GetStreakHistory(ctx context.Context, id uuid.UUID) ([]StreakHistory, error) {
	habit, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if habit == nil {
		return nil, ErrHabitNotFound
	}

	return s.repo.GetStreakHistory(ctx, id)
}

func (s *service) GetHabitsDueToday(ctx context.Context, userID uuid.UUID) ([]Habit, error) {
	// Modified to return all active habits for the user
	habits, _, err := s.repo.FindAll(ctx, HabitFilter{UserID: &userID})
	if err != nil {
		return nil, err
	}

	// Filter to only include active (non-completed) habits
	var activeHabits []Habit
	for _, habit := range habits {
		if !habit.IsCompleted {
			activeHabits = append(activeHabits, habit)
		}
	}

	s.logger.Info("GetHabitsDueToday results",
		zap.String("user_id", userID.String()),
		zap.Int("total_found", len(activeHabits)))

	return activeHabits, nil
}

// LogHabitCompletion records a habit completion for the heatmap
func (s *service) LogHabitCompletion(ctx context.Context, habitID uuid.UUID, userID uuid.UUID, date time.Time) error {
	return s.repo.LogHabitCompletion(ctx, habitID, userID, date)
}

// GetHeatmapData retrieves habit completion data for the heatmap visualization
func (s *service) GetHeatmapData(ctx context.Context, userID uuid.UUID, period string) (map[string]int, error) {
	now := time.Now()
	var startDate time.Time

	// Calculate start date based on the requested period
	switch period {
	case "year":
		startDate = now.AddDate(-1, 0, 0)
	case "month":
		startDate = now.AddDate(0, -1, 0)
	case "week":
		startDate = now.AddDate(0, 0, -7)
	default:
		// Default to last year
		startDate = now.AddDate(-1, 0, 0)
	}

	return s.repo.GetHeatmapData(ctx, userID, startDate, now)
}

// SendHabitReminders sends reminder notifications for habits due today
func (s *service) SendHabitReminders(ctx context.Context) error {
	// Get all habits due today that haven't been completed
	habits, err := s.repo.GetUncompletedHabitsDueToday(ctx)
	if err != nil {
		return fmt.Errorf("failed to get habits due today: %w", err)
	}

	var sent int
	for _, habit := range habits {
		// Only send reminders if the notification service is available
		if s.notifySvc != nil {
			if err := s.notifySvc.NotifyHabitReminder(ctx, habit.UserID, &habit); err != nil {
				log.Printf("failed to send habit reminder notification for habit %s: %v", habit.ID, err)
				continue
			}
			sent++
		}
	}

	log.Printf("sent %d habit reminders", sent)
	return nil
}

// Helper to marshal metadata
func marshalHabitMetadata(data map[string]interface{}) string {
	b, err := json.Marshal(data)
	if err != nil {
		return "{}"
	}
	return string(b)
}

// Analytics implementation
func (s *service) RecordHabitActivity(ctx context.Context, input RecordHabitActivityInput) error {
	timestamp := input.Timestamp
	if timestamp.IsZero() {
		timestamp = time.Now()
	}

	metadata := ""
	if input.Metadata != nil {
		metadataJSON, err := json.Marshal(input.Metadata)
		if err == nil {
			metadata = string(metadataJSON)
		}
	}

	analytics := &HabitAnalytics{
		ID:        uuid.New(),
		HabitID:   input.HabitID,
		UserID:    input.UserID,
		Action:    input.Action,
		Timestamp: timestamp,
		Metadata:  metadata,
	}

	return s.repo.RecordHabitActivity(ctx, analytics)
}

func (s *service) GetHabitAnalytics(ctx context.Context, habitID uuid.UUID, startTime, endTime time.Time, page, pageSize int) ([]HabitAnalytics, int64, error) {
	filter := AnalyticsFilter{
		HabitID:   &habitID,
		StartTime: &startTime,
		EndTime:   &endTime,
		Page:      page,
		PageSize:  pageSize,
	}

	return s.repo.GetHabitAnalytics(ctx, filter)
}

func (s *service) GetUserHabitAnalytics(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time, page, pageSize int) ([]HabitAnalytics, int64, error) {
	filter := AnalyticsFilter{
		UserID:    &userID,
		StartTime: &startTime,
		EndTime:   &endTime,
		Page:      page,
		PageSize:  pageSize,
	}

	return s.repo.GetHabitAnalytics(ctx, filter)
}

func (s *service) GetHabitActivitySummary(ctx context.Context, habitID uuid.UUID, startTime, endTime time.Time) (*HabitActivitySummary, error) {
	// Verify the habit exists
	habit, err := s.repo.FindByID(ctx, habitID)
	if err != nil {
		return nil, err
	}
	if habit == nil {
		return nil, ErrHabitNotFound
	}

	actionCounts, err := s.repo.GetHabitActivitySummary(ctx, habitID, startTime, endTime)
	if err != nil {
		return nil, err
	}

	// Calculate total actions
	totalActions := 0
	for _, count := range actionCounts {
		totalActions += count
	}

	return &HabitActivitySummary{
		HabitID:      habitID,
		ActionCounts: actionCounts,
		StartTime:    startTime,
		EndTime:      endTime,
		TotalActions: totalActions,
	}, nil
}

func (s *service) GetUserHabitActivitySummary(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time) (*UserHabitActivitySummary, error) {
	actionCounts, err := s.repo.GetUserHabitActivitySummary(ctx, userID, startTime, endTime)
	if err != nil {
		return nil, err
	}

	// Calculate total actions
	totalActions := 0
	for _, count := range actionCounts {
		totalActions += count
	}

	return &UserHabitActivitySummary{
		UserID:       userID,
		ActionCounts: actionCounts,
		StartTime:    startTime,
		EndTime:      endTime,
		TotalActions: totalActions,
	}, nil
}

// Define HabitsDashboardMetrics struct for dashboard metrics aggregation
// HabitsDashboardMetrics represents summary metrics for the dashboard
// Used by GetDashboardMetrics
type HabitsDashboardMetrics struct {
	Total     int
	Active    int
	Completed int
	Streak    int
}

func (s *service) GetDashboardMetrics(userID uuid.UUID) (HabitsDashboardMetrics, error) {
	ctx := context.Background()
	filter := HabitFilter{UserID: &userID}
	habits, _, err := s.repo.FindAll(ctx, filter)
	if err != nil {
		return HabitsDashboardMetrics{}, err
	}
	total := len(habits)
	active := 0
	completed := 0
	streak := 0
	for _, h := range habits {
		if h.IsCompleted {
			completed++
		} else {
			active++
		}
		if h.CurrentStreak > streak {
			streak = h.CurrentStreak
		}
	}
	return HabitsDashboardMetrics{
		Total:     total,
		Active:    active,
		Completed: completed,
		Streak:    streak,
	}, nil
}

func (s *service) recordHabitActivity(ctx context.Context, habit *Habit, userID uuid.UUID, action string, metadata map[string]interface{}) {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadata["action"] = action

	// Publish dashboard event for cache invalidation
	event := &events.DashboardEvent{
		EventType: events.DashboardEventCacheInvalidate,
		UserID:    userID,
		EntityID:  habit.ID,
		Timestamp: time.Now().UTC(),
		Details:   metadata,
	}
	if err := s.redis.PublishDashboardEvent(ctx, event); err != nil {
		s.logger.Error("Failed to publish dashboard event", zap.Error(err))
	}
}
