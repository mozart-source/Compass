package migrations

import (
	"fmt"
	"time"

	"errors"

	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/calendar"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/habits"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/organization"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/project"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/roles"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/task"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/todos"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/user"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/workflow"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/notification"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/infrastructure/persistence/postgres/connection"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// MigrationRecord tracks the migration history
type MigrationRecord struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"not null;unique"`
	Version   int       `gorm:"not null"`
	AppliedAt time.Time `gorm:"not null"`
}

// TableName specifies the table name for migration records
func (MigrationRecord) TableName() string {
	return "schema_migrations"
}

// Try to enable the pgvector extension if it exists
// This will silently continue if the extension doesn't exist
func tryEnablePgVector(db *gorm.DB, logger *zap.Logger) {
	// Try to create the extension, but don't fail if it doesn't exist
	err := db.Exec("CREATE EXTENSION IF NOT EXISTS vector").Error
	if err != nil {
		logger.Warn("Could not enable pgvector extension", zap.Error(err))
		logger.Info("Vector operations will use text representation instead of native vector types")
	} else {
		logger.Info("Successfully enabled pgvector extension")
	}
}

// AutoMigrate runs database migrations for all models
func AutoMigrate(db *connection.Database, logger *zap.Logger) error {
	logger.Info("Starting automatic database migration...")

	// Try to enable pgvector if available
	tryEnablePgVector(db.DB, logger)

	// Enable UUID extension for PostgreSQL
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`).Error; err != nil {
		logger.Error("Failed to create UUID extension", zap.Error(err))
		return fmt.Errorf("failed to create UUID extension: %v", err)
	}

	// Create migrations table if it doesn't exist
	if err := db.AutoMigrate(&MigrationRecord{}); err != nil {
		logger.Error("Failed to create migrations table", zap.Error(err))
		return fmt.Errorf("failed to create migrations table: %v", err)
	}

	// Start a transaction for the entire migration process
	return db.Transaction(func(tx *gorm.DB) error {
		// Create a wrapped database connection for the transaction
		txDB := &connection.Database{DB: tx}

		// Get the current highest version number
		var lastVersion int
		if err := txDB.Model(&MigrationRecord{}).Select("COALESCE(MAX(version), 0)").Scan(&lastVersion).Error; err != nil {
			return fmt.Errorf("failed to get last version: %v", err)
		}

		// Define the models in the order they should be migrated
		// This order matters due to foreign key relationships
		models := []interface{}{
			&notification.Notification{},
			&roles.Role{},
			&roles.Permission{},
			&user.User{}, // Users should be first as they're referenced by other tables
			&roles.UserRole{},
			&roles.RolePermission{},
			&organization.Organization{}, // Organizations depend on users
			&project.Project{},           // Projects depend on organizations
			&task.Task{},                 // Tasks depend on projects, users, and organizations
			&habits.Habit{},
			&habits.StreakHistory{},
			&habits.HabitCompletionLog{},
			&calendar.CalendarEvent{},
			&calendar.RecurrenceRule{},
			&calendar.EventOccurrence{},
			&calendar.EventException{},
			&calendar.EventReminder{},
			&calendar.EventCollaborator{},
			&workflow.Workflow{},
			&workflow.WorkflowStep{},
			&workflow.WorkflowExecution{},
			&workflow.WorkflowStepExecution{},
			&workflow.WorkflowAgentLink{},
			&workflow.WorkflowTransition{},
			&todos.Todo{},
			&user.UserAnalytics{},
			&user.SessionAnalytics{},
			&task.TaskAnalytics{},
			&calendar.EventAnalytics{},
			&habits.HabitAnalytics{},
		}

		// Migrate each model
		for i, model := range models {
			modelName := fmt.Sprintf("%T", model)

			// Check if this model has been migrated
			var record MigrationRecord
			err := txDB.Where("name = ?", modelName).First(&record).Error
			isNewMigration := err == gorm.ErrRecordNotFound

			// Run the migration
			if err := txDB.AutoMigrate(model); err != nil {
				logger.Error("Failed to migrate model",
					zap.String("model", modelName),
					zap.Error(err),
				)
				return fmt.Errorf("failed to migrate %s: %v", modelName, err)
			}

			// Record the migration if it's new
			if isNewMigration {
				record = MigrationRecord{
					Name:      modelName,
					Version:   lastVersion + i + 1, // Increment version for each new migration
					AppliedAt: time.Now(),
				}
				if err := txDB.Create(&record).Error; err != nil {
					logger.Error("Failed to record migration",
						zap.String("model", modelName),
						zap.Error(err),
					)
					return fmt.Errorf("failed to record migration for %s: %v", modelName, err)
				}
				logger.Info("Applied new migration",
					zap.String("model", modelName),
					zap.Int("version", record.Version),
				)
			}
		}

		// Create default roles and permissions
		if err := createDefaultRolesAndPermissions(tx); err != nil {
			return err
		}

		logger.Info("Database migration completed successfully")
		return nil
	})
}

// createDefaultRolesAndPermissions creates default roles and permissions
func createDefaultRolesAndPermissions(db *gorm.DB) error {
	// Create default permissions
	permissions := []roles.Permission{
		{Name: "organizations:create", Description: "Create organizations"},
		{Name: "organizations:read", Description: "Read organizations"},
		{Name: "organizations:update", Description: "Update organizations"},
		{Name: "organizations:delete", Description: "Delete organizations"},

		{Name: "projects:create", Description: "Create projects"},
		{Name: "projects:read", Description: "Read projects"},
		{Name: "projects:update", Description: "Update projects"},
		{Name: "projects:delete", Description: "Delete projects"},

		{Name: "tasks:create", Description: "Create tasks"},
		{Name: "tasks:read", Description: "Read tasks"},
		{Name: "tasks:update", Description: "Update tasks"},
		{Name: "tasks:delete", Description: "Delete tasks"},

		{Name: "roles:create", Description: "Create roles"},
		{Name: "roles:read", Description: "Read roles"},
		{Name: "roles:update", Description: "Update roles"},
		{Name: "roles:delete", Description: "Delete roles"},
		{Name: "roles:assign", Description: "Assign roles to users"},
	}

	// Create permissions if they don't exist
	for _, perm := range permissions {
		if err := db.Where("name = ?", perm.Name).FirstOrCreate(&perm).Error; err != nil {
			return fmt.Errorf("failed to create permission %s: %w", perm.Name, err)
		}
	}

	// Create default roles
	defaultRoles := []struct {
		Role        roles.Role
		Permissions []string
	}{
		{
			Role: roles.Role{
				Name:        "admin",
				Description: "Administrator role with full access",
			},
			Permissions: []string{
				"organizations:create", "organizations:read", "organizations:update", "organizations:delete",
				"projects:create", "projects:read", "projects:update", "projects:delete",
				"tasks:create", "tasks:read", "tasks:update", "tasks:delete",
				"roles:create", "roles:read", "roles:update", "roles:delete", "roles:assign",
			},
		},
		{
			Role: roles.Role{
				Name:        "user",
				Description: "Regular user role with basic access",
			},
			Permissions: []string{
				"organizations:read",
				"projects:read",
				"tasks:read", "tasks:create", "tasks:update",
			},
		},
	}

	for _, r := range defaultRoles {
		// Create role if it doesn't exist
		var existingRole roles.Role
		if err := db.Where("name = ?", r.Role.Name).FirstOrCreate(&existingRole, r.Role).Error; err != nil {
			return fmt.Errorf("failed to create role %s: %w", r.Role.Name, err)
		}

		// Get all permissions for this role
		for _, permName := range r.Permissions {
			var perm roles.Permission
			if err := db.Where("name = ?", permName).First(&perm).Error; err != nil {
				return fmt.Errorf("permission %s not found: %w", permName, err)
			}

			// Check if role-permission association already exists
			var rolePermission roles.RolePermission
			err := db.Where("role_id = ? AND permission_id = ?", existingRole.ID, perm.ID).
				First(&rolePermission).Error

			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					// Create new association only if it doesn't exist
					rolePermission = roles.RolePermission{
						RoleID:       existingRole.ID,
						PermissionID: perm.ID,
						CreatedAt:    time.Now(),
						UpdatedAt:    time.Now(),
					}
					if err := db.Create(&rolePermission).Error; err != nil {
						return fmt.Errorf("failed to assign permission %s to role %s: %w",
							permName, r.Role.Name, err)
					}
				} else {
					return fmt.Errorf("error checking role-permission association: %w", err)
				}
			}
		}
	}

	return nil
}

// GetMigrationHistory returns the history of applied migrations
func GetMigrationHistory(db *connection.Database) ([]MigrationRecord, error) {
	var records []MigrationRecord
	err := db.Order("version ASC").Find(&records).Error
	return records, err
}
