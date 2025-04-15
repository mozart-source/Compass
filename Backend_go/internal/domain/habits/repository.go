package habits

import (
	"context"
	"errors"
	"time"

	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/infrastructure/persistence/postgres/connection"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrHabitNotFound = errors.New("habit not found")
	ErrInvalidInput  = errors.New("invalid input")
)

// HabitFilter defines the filtering options for habits
type HabitFilter struct {
	UserID   *uuid.UUID
	Title    *string
	Page     int
	PageSize int
}

// Repository defines the interface for habit persistence operations
type Repository interface {
	Create(ctx context.Context, habit *Habit) error
	FindByID(ctx context.Context, id uuid.UUID) (*Habit, error)
	FindAll(ctx context.Context, filter HabitFilter) ([]Habit, int64, error)
	Update(ctx context.Context, habit *Habit) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByTitle(ctx context.Context, title string, userID uuid.UUID) (*Habit, error)
	MarkCompleted(ctx context.Context, id uuid.UUID, userID uuid.UUID, completionDate *time.Time) error
	UnmarkCompleted(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	ResetDailyCompletions(ctx context.Context) (int64, error)
	CheckAndResetBrokenStreaks(ctx context.Context) (int64, error)
	GetTopStreaks(ctx context.Context, userID uuid.UUID, limit int) ([]Habit, error)
	GetHabitsDueToday(ctx context.Context, userID uuid.UUID) ([]Habit, error)
	GetUncompletedHabitsDueToday(ctx context.Context) ([]Habit, error)
	FindCompletedHabits(ctx context.Context, habits *[]Habit) error
	GetActiveStreaks(ctx context.Context) ([]Habit, error)
	LogStreakHistory(ctx context.Context, habitID uuid.UUID, streakLength int, lastCompletedDate time.Time) error
	ResetStreak(ctx context.Context, habitID uuid.UUID) error
	GetStreakHistory(ctx context.Context, habitID uuid.UUID) ([]StreakHistory, error)
	UpdateStreakQuality(ctx context.Context, habitID uuid.UUID) error
	IsStreakBroken(ctx context.Context, lastCompletedDate *time.Time) (bool, error)

	// Heatmap related methods
	LogHabitCompletion(ctx context.Context, habitID uuid.UUID, userID uuid.UUID, date time.Time) error
	RemoveHabitCompletion(ctx context.Context, habitID uuid.UUID, userID uuid.UUID, date time.Time) error
	GetHeatmapData(ctx context.Context, userID uuid.UUID, startDate time.Time, endDate time.Time) (map[string]int, error)

	// Analytics methods
	RecordHabitActivity(ctx context.Context, analytics *HabitAnalytics) error
	GetHabitAnalytics(ctx context.Context, filter AnalyticsFilter) ([]HabitAnalytics, int64, error)
	GetHabitActivitySummary(ctx context.Context, habitID uuid.UUID, startTime, endTime time.Time) (map[string]int, error)
	GetUserHabitActivitySummary(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time) (map[string]int, error)
}

type repository struct {
	db *connection.Database
}

func NewRepository(db *connection.Database) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, habit *Habit) error {
	return r.db.WithContext(ctx).Create(habit).Error
}

func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*Habit, error) {
	var habit Habit
	result := r.db.WithContext(ctx).First(&habit, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrHabitNotFound
		}
		return nil, result.Error
	}
	return &habit, nil
}

func (r *repository) FindAll(ctx context.Context, filter HabitFilter) ([]Habit, int64, error) {
	var habits []Habit
	var total int64
	query := r.db.WithContext(ctx).Model(&Habit{})

	if filter.UserID != nil {
		query = query.Where("user_id = ?", *filter.UserID)
	}

	if filter.Title != nil {
		query = query.Where("title LIKE ?", "%"+*filter.Title+"%")
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Set default PageSize if not set
	if filter.PageSize == 0 {
		filter.PageSize = 10000
	}

	err = query.Offset(filter.Page * filter.PageSize).
		Limit(filter.PageSize).
		Find(&habits).Error
	if err != nil {
		return nil, 0, err
	}

	return habits, total, nil
}

func (r *repository) Update(ctx context.Context, habit *Habit) error {
	result := r.db.WithContext(ctx).Save(habit)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrHabitNotFound
	}
	return nil
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&Habit{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrHabitNotFound
	}
	return nil
}

func (r *repository) FindByTitle(ctx context.Context, title string, userID uuid.UUID) (*Habit, error) {
	var habit Habit
	result := r.db.WithContext(ctx).
		Where("title = ? AND user_id = ?", title, userID).
		First(&habit)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrHabitNotFound
		}
		return nil, result.Error
	}
	return &habit, nil
}

func (r *repository) MarkCompleted(ctx context.Context, id uuid.UUID, userID uuid.UUID, completionDate *time.Time) error {
	now := time.Now()
	if completionDate == nil {
		completionDate = &now
	}

	result := r.db.WithContext(ctx).Model(&Habit{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(map[string]interface{}{
			"is_completed":        true,
			"last_completed_date": completionDate,
			"current_streak":      gorm.Expr("current_streak + 1"),
			"longest_streak":      gorm.Expr("GREATEST(longest_streak, current_streak + 1)"),
		})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrHabitNotFound
	}
	return nil
}

func (r *repository) UnmarkCompleted(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	result := r.db.WithContext(ctx).Model(&Habit{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(map[string]interface{}{
			"is_completed":   false,
			"current_streak": gorm.Expr("current_streak - 1"),
		})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrHabitNotFound
	}
	return nil
}

func (r *repository) ResetDailyCompletions(ctx context.Context) (int64, error) {
	// Use TIMEZONE function in postgres to ensure dates are compared in the user's timezone
	result := r.db.WithContext(ctx).Model(&Habit{}).
		Where("is_completed = ? AND DATE(last_completed_date AT TIME ZONE 'UTC') < DATE(NOW() AT TIME ZONE 'UTC')", true).
		Update("is_completed", false)

	return result.RowsAffected, result.Error
}

func (r *repository) CheckAndResetBrokenStreaks(ctx context.Context) (int64, error) {
	// Use TIMEZONE function in postgres to ensure dates are compared in the user's timezone
	// Align the timezone handling exactly like in ResetDailyCompletions
	result := r.db.WithContext(ctx).Model(&Habit{}).
		Where("current_streak > 0 AND (last_completed_date IS NULL OR DATE(last_completed_date AT TIME ZONE 'UTC') < DATE(NOW() AT TIME ZONE 'UTC' - INTERVAL '1 day'))").
		Updates(map[string]interface{}{
			"current_streak": 0,
		})

	return result.RowsAffected, result.Error
}

func (r *repository) GetTopStreaks(ctx context.Context, userID uuid.UUID, limit int) ([]Habit, error) {
	var habits []Habit
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("current_streak desc").
		Limit(limit).
		Find(&habits).Error

	return habits, err
}

func (r *repository) GetHabitsDueToday(ctx context.Context, userID uuid.UUID) ([]Habit, error) {
	var habits []Habit
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_completed = ? AND start_day <= ? AND (end_day IS NULL OR end_day >= ?)",
			userID, false, today, today).
		Find(&habits).Error

	return habits, err
}

func (r *repository) FindCompletedHabits(ctx context.Context, habits *[]Habit) error {
	return r.db.WithContext(ctx).
		Where("is_completed = ?", true).
		Find(habits).Error
}

func (r *repository) GetActiveStreaks(ctx context.Context) ([]Habit, error) {
	var habits []Habit
	err := r.db.WithContext(ctx).
		Where("current_streak > 0").
		Find(&habits).Error
	return habits, err
}

func (r *repository) LogStreakHistory(ctx context.Context, habitID uuid.UUID, streakLength int, lastCompletedDate time.Time) error {
	// Calculate the actual start date by going back streakLength-1 days
	startDate := lastCompletedDate.AddDate(0, 0, -streakLength+1)

	// Get any existing streak history that might overlap
	var existingHistory []StreakHistory
	if err := r.db.WithContext(ctx).
		Where("habit_id = ? AND ((start_date BETWEEN ? AND ?) OR (end_date BETWEEN ? AND ?))",
			habitID, startDate, lastCompletedDate, startDate, lastCompletedDate).
		Find(&existingHistory).Error; err != nil {
		return err
	}

	// If there's overlap, adjust the completed days
	adjustedCompletedDays := streakLength
	for _, h := range existingHistory {
		if h.StartDate.After(startDate) && h.StartDate.Before(lastCompletedDate) ||
			h.EndDate.After(startDate) && h.EndDate.Before(lastCompletedDate) {
			// Subtract any overlapping days to avoid double counting
			overlap := int(h.EndDate.Sub(h.StartDate).Hours()/24) + 1
			adjustedCompletedDays = streakLength - overlap
		}
	}

	history := StreakHistory{
		ID:            uuid.New(),
		HabitID:       habitID,
		StartDate:     startDate,
		EndDate:       lastCompletedDate,
		StreakLength:  streakLength,
		CompletedDays: adjustedCompletedDays,
		CreatedAt:     time.Now(),
	}

	return r.db.WithContext(ctx).Create(&history).Error
}

func (r *repository) ResetStreak(ctx context.Context, habitID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&Habit{}).
		Where("id = ?", habitID).
		Updates(map[string]interface{}{
			"current_streak":    0,
			"streak_start_date": nil,
		}).Error
}

func (r *repository) GetStreakHistory(ctx context.Context, habitID uuid.UUID) ([]StreakHistory, error) {
	var history []StreakHistory
	err := r.db.WithContext(ctx).
		Where("habit_id = ?", habitID).
		Order("end_date DESC").
		Find(&history).Error
	return history, err
}

func (r *repository) UpdateStreakQuality(ctx context.Context, habitID uuid.UUID) error {
	var history []StreakHistory
	var totalDays, completedDays int

	// Fetch history
	if err := r.db.WithContext(ctx).
		Where("habit_id = ?", habitID).
		Find(&history).Error; err != nil {
		return err
	}

	// Calculate quality
	for _, h := range history {
		days := int(h.EndDate.Sub(h.StartDate).Hours() / 24)
		totalDays += days
		completedDays += h.CompletedDays
	}

	quality := 0.0
	if totalDays > 0 {
		quality = float64(completedDays) / float64(totalDays)
	}

	// Update habit
	return r.db.WithContext(ctx).Model(&Habit{}).
		Where("id = ?", habitID).
		Update("streak_quality", quality).Error
}

func (r *repository) IsStreakBroken(ctx context.Context, lastCompletedDate *time.Time) (bool, error) {
	if lastCompletedDate == nil {
		return true, nil
	}

	var isBroken bool
	query := `SELECT DATE(? AT TIME ZONE 'UTC') < DATE(NOW() AT TIME ZONE 'UTC' - INTERVAL '1 day')`
	err := r.db.WithContext(ctx).Raw(query, lastCompletedDate).Scan(&isBroken).Error
	return isBroken, err
}

func (r *repository) LogHabitCompletion(ctx context.Context, habitID uuid.UUID, userID uuid.UUID, date time.Time) error {
	// Create a new habit completion log entry
	log := HabitCompletionLog{
		ID:        uuid.New(),
		HabitID:   habitID,
		UserID:    userID,
		Date:      date,
		CreatedAt: time.Now(),
	}

	return r.db.WithContext(ctx).Create(&log).Error
}

func (r *repository) RemoveHabitCompletion(ctx context.Context, habitID uuid.UUID, userID uuid.UUID, date time.Time) error {
	// Delete the completion log for the specific habit, user and date
	result := r.db.WithContext(ctx).
		Where("habit_id = ? AND user_id = ? AND DATE(date) = DATE(?)",
			habitID, userID, date).
		Delete(&HabitCompletionLog{})

	return result.Error
}

func (r *repository) GetHeatmapData(ctx context.Context, userID uuid.UUID, startDate time.Time, endDate time.Time) (map[string]int, error) {
	// Query to get counts of completed habits per day
	var results []struct {
		Date           string
		CompletedCount int
	}

	// Format the date as YYYY-MM-DD string in the database query
	query := `
		SELECT 
			TO_CHAR(date, 'YYYY-MM-DD') AS date, 
			COUNT(*) AS completed_count
		FROM 
			habit_completion_logs
		WHERE 
			user_id = ? 
			AND date BETWEEN ? AND ?
		GROUP BY 
			TO_CHAR(date, 'YYYY-MM-DD')
		ORDER BY 
			date;
	`

	err := r.db.WithContext(ctx).Raw(query, userID, startDate, endDate).Scan(&results).Error
	if err != nil {
		return nil, err
	}

	// Convert the results to a map for easier access
	heatmapData := make(map[string]int)
	for _, result := range results {
		heatmapData[result.Date] = result.CompletedCount
	}

	return heatmapData, nil
}

// GetUncompletedHabitsDueToday returns all habits from all users that are due today and not yet completed
func (r *repository) GetUncompletedHabitsDueToday(ctx context.Context) ([]Habit, error) {
	var habits []Habit
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	err := r.db.WithContext(ctx).
		Where("is_completed = ? AND start_day <= ? AND (end_day IS NULL OR end_day >= ?)",
			false, today, today).
		Find(&habits).Error

	return habits, err
}

// Analytics implementation
func (r *repository) RecordHabitActivity(ctx context.Context, analytics *HabitAnalytics) error {
	return r.db.WithContext(ctx).Create(analytics).Error
}

func (r *repository) GetHabitAnalytics(ctx context.Context, filter AnalyticsFilter) ([]HabitAnalytics, int64, error) {
	var analytics []HabitAnalytics
	var total int64
	query := r.db.WithContext(ctx).Model(&HabitAnalytics{})

	if filter.HabitID != nil {
		query = query.Where("habit_id = ?", *filter.HabitID)
	}
	if filter.UserID != nil {
		query = query.Where("user_id = ?", *filter.UserID)
	}
	if filter.Action != nil {
		query = query.Where("action = ?", *filter.Action)
	}
	if filter.StartTime != nil && filter.EndTime != nil {
		query = query.Where("timestamp BETWEEN ? AND ?", *filter.StartTime, *filter.EndTime)
	} else if filter.StartTime != nil {
		query = query.Where("timestamp >= ?", *filter.StartTime)
	} else if filter.EndTime != nil {
		query = query.Where("timestamp <= ?", *filter.EndTime)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Order("timestamp DESC").
		Offset(filter.Page * filter.PageSize).
		Limit(filter.PageSize).
		Find(&analytics).Error
	if err != nil {
		return nil, 0, err
	}

	return analytics, total, nil
}

func (r *repository) GetHabitActivitySummary(ctx context.Context, habitID uuid.UUID, startTime, endTime time.Time) (map[string]int, error) {
	var results []struct {
		Action string
		Count  int
	}

	err := r.db.WithContext(ctx).Model(&HabitAnalytics{}).
		Select("action, count(*) as count").
		Where("habit_id = ? AND timestamp BETWEEN ? AND ?", habitID, startTime, endTime).
		Group("action").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	summary := make(map[string]int)
	for _, result := range results {
		summary[result.Action] = result.Count
	}

	return summary, nil
}

func (r *repository) GetUserHabitActivitySummary(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time) (map[string]int, error) {
	var results []struct {
		Action string
		Count  int
	}

	err := r.db.WithContext(ctx).Model(&HabitAnalytics{}).
		Select("action, count(*) as count").
		Where("user_id = ? AND timestamp BETWEEN ? AND ?", userID, startTime, endTime).
		Group("action").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	summary := make(map[string]int)
	for _, result := range results {
		summary[result.Action] = result.Count
	}

	return summary, nil
}
