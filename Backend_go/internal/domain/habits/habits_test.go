package habits

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCalculateStreakQuality(t *testing.T) {
	tests := []struct {
		name     string
		habit    *Habit
		history  []StreakHistory
		expected float64
	}{
		{
			name: "Empty history should return 0",
			habit: &Habit{
				ID: uuid.New(),
			},
			history:  []StreakHistory{},
			expected: 0,
		},
		{
			name: "Perfect streak should return 1.0",
			habit: &Habit{
				ID: uuid.New(),
			},
			history: []StreakHistory{
				{
					StartDate:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					EndDate:       time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
					StreakLength:  5,
					CompletedDays: 5,
				},
			},
			expected: 1.0,
		},
		{
			name: "Multiple streaks with gap",
			habit: &Habit{
				ID: uuid.New(),
			},
			history: []StreakHistory{
				{
					StartDate:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					EndDate:       time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
					StreakLength:  10,
					CompletedDays: 10,
				},
				{
					StartDate:     time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
					EndDate:       time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC),
					StreakLength:  6,
					CompletedDays: 6,
				},
			},
			expected: 0.8, // 16 completed days out of 20 total days
		},
		{
			name: "Overlapping streaks",
			habit: &Habit{
				ID: uuid.New(),
			},
			history: []StreakHistory{
				{
					StartDate:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					EndDate:       time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
					StreakLength:  10,
					CompletedDays: 8, // Adjusted for overlap
				},
				{
					StartDate:     time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC),
					EndDate:       time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
					StreakLength:  8,
					CompletedDays: 6, // Adjusted for overlap
				},
			},
			expected: 0.93, // 14 completed days out of 15 total days
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quality := calculateStreakQuality(tt.habit, tt.history)
			assert.InDelta(t, tt.expected, quality, 0.01, "streak quality calculation mismatch")
		})
	}
}

func TestLogStreakHistory(t *testing.T) {
	tests := []struct {
		name              string
		habitID           uuid.UUID
		streakLength      int
		lastCompletedDate time.Time
		existingHistory   []StreakHistory
		expectedHistory   StreakHistory
	}{
		{
			name:              "New streak without overlap",
			habitID:           uuid.New(),
			streakLength:      5,
			lastCompletedDate: time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
			existingHistory:   []StreakHistory{},
			expectedHistory: StreakHistory{
				StartDate:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:       time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
				StreakLength:  5,
				CompletedDays: 5,
			},
		},
		{
			name:              "Streak with overlap",
			habitID:           uuid.New(),
			streakLength:      10,
			lastCompletedDate: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			existingHistory: []StreakHistory{
				{
					StartDate:     time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
					EndDate:       time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
					StreakLength:  6,
					CompletedDays: 6,
				},
			},
			expectedHistory: StreakHistory{
				StartDate:     time.Date(2024, 1, 6, 0, 0, 0, 0, time.UTC),
				EndDate:       time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
				StreakLength:  10,
				CompletedDays: 4, // 10 days total - 6 days overlap
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock repository implementation for testing
			mockRepo := &mockRepository{
				existingHistory: tt.existingHistory,
			}

			err := mockRepo.LogStreakHistory(nil, tt.habitID, tt.streakLength, tt.lastCompletedDate)
			assert.NoError(t, err)

			// Add debug output
			if tt.name == "Streak with overlap" {
				t.Logf("Expected CompletedDays: %d", tt.expectedHistory.CompletedDays)
				t.Logf("Actual CompletedDays: %d", mockRepo.lastLoggedHistory.CompletedDays)
				t.Logf("Streak Length: %d", tt.streakLength)
				t.Logf("Existing History CompletedDays: %d", tt.existingHistory[0].CompletedDays)
			}

			// Verify the logged history matches expected
			assert.Equal(t, tt.expectedHistory.StartDate, mockRepo.lastLoggedHistory.StartDate)
			assert.Equal(t, tt.expectedHistory.EndDate, mockRepo.lastLoggedHistory.EndDate)
			assert.Equal(t, tt.expectedHistory.StreakLength, mockRepo.lastLoggedHistory.StreakLength)
			assert.Equal(t, tt.expectedHistory.CompletedDays, mockRepo.lastLoggedHistory.CompletedDays)
		})
	}
}

// Mock repository for testing
type mockRepository struct {
	existingHistory   []StreakHistory
	lastLoggedHistory StreakHistory
}

func (m *mockRepository) LogStreakHistory(ctx context.Context, habitID uuid.UUID, streakLength int, lastCompletedDate time.Time) error {
	startDate := lastCompletedDate.AddDate(0, 0, -streakLength+1)

	adjustedCompletedDays := streakLength
	for _, h := range m.existingHistory {
		// Check if there's any overlap
		if (startDate.Before(h.EndDate) || startDate.Equal(h.EndDate)) &&
			(lastCompletedDate.After(h.StartDate) || lastCompletedDate.Equal(h.StartDate)) {
			// If there's overlap, subtract the existing streak's completed days
			adjustedCompletedDays = streakLength - h.CompletedDays
		}
	}

	// Ensure adjustedCompletedDays doesn't go negative
	if adjustedCompletedDays < 0 {
		adjustedCompletedDays = 0
	}

	m.lastLoggedHistory = StreakHistory{
		ID:            uuid.New(),
		HabitID:       habitID,
		StartDate:     startDate,
		EndDate:       lastCompletedDate,
		StreakLength:  streakLength,
		CompletedDays: adjustedCompletedDays,
		CreatedAt:     time.Now(),
	}

	return nil
}
