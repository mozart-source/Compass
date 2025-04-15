package habits

import (
	"context"
	"fmt"

	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/notification"
	"github.com/google/uuid"
)

// HabitNotificationService handles notifications for habits
type HabitNotificationService struct {
	notificationService notification.Service
	domainNotifier      notification.DomainNotifier
}

// NewHabitNotificationService creates a new habit notification service
func NewHabitNotificationService(notificationService notification.Service) *HabitNotificationService {
	return &HabitNotificationService{
		notificationService: notificationService,
	}
}

// WithDomainNotifier adds a domain notifier to the habit notification service
// This provides enhanced functionality while maintaining backward compatibility
func (s *HabitNotificationService) WithDomainNotifier(notifier notification.DomainNotifier) *HabitNotificationService {
	s.domainNotifier = notifier
	return s
}

// NotifyHabitCompleted sends a notification when a habit is completed
func (s *HabitNotificationService) NotifyHabitCompleted(ctx context.Context, userID uuid.UUID, habit *Habit) error {
	title := "Habit Completed"
	content := fmt.Sprintf("You've completed your habit: %s", habit.Title)
	data := map[string]string{
		"habitID": habit.ID.String(),
		"title":   habit.Title,
	}

	// Use domain notifier if available
	if s.domainNotifier != nil {
		// Use async delivery with multiple channels if available
		return s.domainNotifier.NotifyUserWithDelivery(
			ctx,
			userID,
			notification.HabitCompleted,
			title,
			content,
			data,
			"habits",
			habit.ID,
			[]notification.DeliveryMethod{notification.InApp, notification.Email},
		)
	}

	// Fall back to direct service if notifier not configured
	return s.notificationService.CreateForUser(
		ctx,
		userID,
		notification.HabitCompleted,
		title,
		content,
		data,
		"habits",
		habit.ID,
	)
}

// NotifyHabitStreak sends a notification when a habit streak reaches a milestone
func (s *HabitNotificationService) NotifyHabitStreak(ctx context.Context, userID uuid.UUID, habit *Habit) error {
	title := "Habit Streak"
	content := fmt.Sprintf("Amazing! You've maintained a %d day streak for your habit: %s", habit.CurrentStreak, habit.Title)
	data := map[string]string{
		"habitID":       habit.ID.String(),
		"title":         habit.Title,
		"currentStreak": fmt.Sprintf("%d", habit.CurrentStreak),
	}

	// Use domain notifier if available
	if s.domainNotifier != nil {
		// For streaks, we use supported channels only
		return s.domainNotifier.NotifyUserWithDelivery(
			ctx,
			userID,
			notification.HabitStreak,
			title,
			content,
			data,
			"habits",
			habit.ID,
			[]notification.DeliveryMethod{notification.InApp, notification.Email},
		)
	}

	// Fall back to direct service
	return s.notificationService.CreateForUser(
		ctx,
		userID,
		notification.HabitStreak,
		title,
		content,
		data,
		"habits",
		habit.ID,
	)
}

// NotifyHabitStreakBroken sends a notification when a habit streak is broken
func (s *HabitNotificationService) NotifyHabitStreakBroken(ctx context.Context, userID uuid.UUID, habit *Habit, streakLength int) error {
	title := "Habit Streak Broken"
	content := fmt.Sprintf("Your %d day streak for habit \"%s\" has been broken. Don't worry, you can start a new streak today!", streakLength, habit.Title)
	data := map[string]string{
		"habitID":      habit.ID.String(),
		"title":        habit.Title,
		"streakLength": fmt.Sprintf("%d", streakLength),
	}

	// Use domain notifier if available
	if s.domainNotifier != nil {
		return s.domainNotifier.NotifyUserWithDelivery(
			ctx,
			userID,
			notification.HabitBroken,
			title,
			content,
			data,
			"habits",
			habit.ID,
			[]notification.DeliveryMethod{notification.InApp, notification.Email},
		)
	}

	// Fall back to direct service
	return s.notificationService.CreateForUser(
		ctx,
		userID,
		notification.HabitBroken,
		title,
		content,
		data,
		"habits",
		habit.ID,
	)
}

// NotifyHabitReminder sends a reminder notification for a habit
func (s *HabitNotificationService) NotifyHabitReminder(ctx context.Context, userID uuid.UUID, habit *Habit) error {
	title := "Habit Reminder"
	content := fmt.Sprintf("Don't forget to complete your habit: %s", habit.Title)
	data := map[string]string{
		"habitID": habit.ID.String(),
		"title":   habit.Title,
	}

	// Use domain notifier if available
	if s.domainNotifier != nil {
		// For reminders, we only use InApp for now
		return s.domainNotifier.NotifyUserWithDelivery(
			ctx,
			userID,
			notification.HabitReminder,
			title,
			content,
			data,
			"habits",
			habit.ID,
			[]notification.DeliveryMethod{notification.InApp},
		)
	}

	// Fall back to direct service
	return s.notificationService.CreateForUser(
		ctx,
		userID,
		notification.HabitReminder,
		title,
		content,
		data,
		"habits",
		habit.ID,
	)
}

// NotifyHabitMilestone sends a notification when a habit reaches a significant milestone
func (s *HabitNotificationService) NotifyHabitMilestone(ctx context.Context, userID uuid.UUID, habit *Habit, milestone string, achievementDesc string) error {
	title := fmt.Sprintf("Habit Milestone: %s", milestone)
	content := fmt.Sprintf("Congratulations! %s for habit \"%s\"", achievementDesc, habit.Title)
	data := map[string]string{
		"habitID":   habit.ID.String(),
		"title":     habit.Title,
		"milestone": milestone,
	}

	// Use domain notifier if available
	if s.domainNotifier != nil {
		// For milestones, we use supported channels only
		return s.domainNotifier.NotifyUserWithDelivery(
			ctx,
			userID,
			notification.HabitMilestone,
			title,
			content,
			data,
			"habits",
			habit.ID,
			[]notification.DeliveryMethod{notification.InApp, notification.Email},
		)
	}

	// Fall back to direct service
	return s.notificationService.CreateForUser(
		ctx,
		userID,
		notification.HabitMilestone,
		title,
		content,
		data,
		"habits",
		habit.ID,
	)
}

// ShouldSendStreakNotification determines if a streak notification should be sent
// sent for milestones (3 days, 7 days, 14 days, 30 days, etc)
func (s *HabitNotificationService) ShouldSendStreakNotification(streak int) bool {
	milestones := []int{3, 7, 14, 21, 30, 60, 90, 100, 180, 365}

	for _, milestone := range milestones {
		if streak == milestone {
			return true
		}
	}

	return false
}
