package database

import (
	"fmt"
	"log"

	"github.com/barannkoca/banking-backend/internal/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// SeedData creates initial data for development
func SeedData() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	// Check if users already exist
	var userCount int64
	DB.Model(&models.User{}).Count(&userCount)
	if userCount > 0 {
		log.Println("ℹ️ Database already seeded, skipping...")
		return nil
	}

	log.Println("🌱 Seeding database...")

	// Create admin user
	adminPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	adminUser := &models.User{
		ID:           uuid.New(),
		Username:     "admin",
		Email:        "admin@banking.com",
		PasswordHash: string(adminPassword),
		Role:         models.RoleAdmin,
	}

	if err := DB.Create(adminUser).Error; err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	// Create admin balance
	adminBalance := &models.Balance{
		UserID: adminUser.ID,
		Amount: 1000000.00, // 1M initial balance for admin
	}
	if err := DB.Create(adminBalance).Error; err != nil {
		return fmt.Errorf("failed to create admin balance: %w", err)
	}

	// Create test customer
	customerPassword, _ := bcrypt.GenerateFromPassword([]byte("customer123"), bcrypt.DefaultCost)
	customerUser := &models.User{
		ID:           uuid.New(),
		Username:     "johndoe",
		Email:        "john@example.com",
		PasswordHash: string(customerPassword),
		Role:         models.RoleCustomer,
	}

	if err := DB.Create(customerUser).Error; err != nil {
		return fmt.Errorf("failed to create customer user: %w", err)
	}

	// Create customer balance
	customerBalance := &models.Balance{
		UserID: customerUser.ID,
		Amount: 1000.00, // 1K initial balance for customer
	}
	if err := DB.Create(customerBalance).Error; err != nil {
		return fmt.Errorf("failed to create customer balance: %w", err)
	}

	// Create another test customer
	customer2Password, _ := bcrypt.GenerateFromPassword([]byte("customer456"), bcrypt.DefaultCost)
	customer2User := &models.User{
		ID:           uuid.New(),
		Username:     "janedoe",
		Email:        "jane@example.com",
		PasswordHash: string(customer2Password),
		Role:         models.RoleCustomer,
	}

	if err := DB.Create(customer2User).Error; err != nil {
		return fmt.Errorf("failed to create customer2 user: %w", err)
	}

	// Create customer2 balance
	customer2Balance := &models.Balance{
		UserID: customer2User.ID,
		Amount: 500.00, // 500 initial balance for customer2
	}
	if err := DB.Create(customer2Balance).Error; err != nil {
		return fmt.Errorf("failed to create customer2 balance: %w", err)
	}

	// Create sample transaction
	sampleTx := &models.Transaction{
		FromUserID: &adminUser.ID,
		ToUserID:   &customerUser.ID,
		Amount:     100.00,
		Type:       models.TransactionTypeTransfer,
		Status:     models.TransactionStatusCompleted,
		Reference:  "Initial deposit",
	}
	if err := DB.Create(sampleTx).Error; err != nil {
		return fmt.Errorf("failed to create sample transaction: %w", err)
	}

	// Create audit log for seeding
	auditLog := &models.AuditLog{
		EntityType: string(models.EntityTypeUser),
		EntityID:   "system",
		Action:     string(models.AuditActionCreate),
		Details:    "Database seeded with initial data",
		IPAddress:  "127.0.0.1",
		UserAgent:  "System Seeder",
	}
	if err := DB.Create(auditLog).Error; err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	log.Println("✅ Database seeding completed successfully")
	log.Println("👤 Admin user: admin@banking.com / admin123")
	log.Println("👤 Customer 1: john@example.com / customer123")
	log.Println("👤 Customer 2: jane@example.com / customer456")

	return nil
}
