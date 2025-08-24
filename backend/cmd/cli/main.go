package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/feature-voting-platform/backend/internal/config"
	"github.com/feature-voting-platform/backend/internal/models"
	"github.com/feature-voting-platform/backend/internal/repository"
	"github.com/feature-voting-platform/backend/pkg/utils"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := repository.NewDatabase(cfg.Database.URL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)

	// Define command line flags
	var (
		command  = flag.String("command", "", "Command to execute (create-user)")
		name     = flag.String("name", "", "Username for create-user command")
		email    = flag.String("email", "", "Email for create-user command")
		password = flag.String("password", "", "Password for create-user command")
	)

	flag.Parse()

	switch *command {
	case "create-user":
		err := createUser(userRepo, *name, *email, *password)
		if err != nil {
			log.Fatalf("Failed to create user: %v", err)
		}
	default:
		fmt.Println("Feature Voting Platform CLI")
		fmt.Println("")
		fmt.Println("Available commands:")
		fmt.Println("  create-user   Create a new user")
		fmt.Println("")
		fmt.Println("Usage:")
		fmt.Println("  create-user -name=<username> -email=<email> -password=<password>")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  ./cli -command=create-user -name=john_doe -email=john@example.com -password=securepass")
		os.Exit(1)
	}
}

func createUser(userRepo *repository.UserRepository, username, email, password string) error {
	// Validate input
	if username == "" {
		return fmt.Errorf("username is required")
	}
	if email == "" {
		return fmt.Errorf("email is required")
	}
	if password == "" {
		return fmt.Errorf("password is required")
	}

	// Basic validation
	username = strings.TrimSpace(username)
	email = strings.ToLower(strings.TrimSpace(email))

	if len(username) < 3 || len(username) > 50 {
		return fmt.Errorf("username must be between 3 and 50 characters")
	}
	if len(password) < 6 {
		return fmt.Errorf("password must be at least 6 characters")
	}
	if !strings.Contains(email, "@") {
		return fmt.Errorf("invalid email format")
	}

	// Check if user already exists
	emailExists, err := userRepo.EmailExists(email)
	if err != nil {
		return fmt.Errorf("failed to check if email exists: %w", err)
	}
	if emailExists {
		return fmt.Errorf("user with email '%s' already exists", email)
	}

	usernameExists, err := userRepo.UsernameExists(username)
	if err != nil {
		return fmt.Errorf("failed to check if username exists: %w", err)
	}
	if usernameExists {
		return fmt.Errorf("user with username '%s' already exists", username)
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{
		Username:     username,
		Email:        email,
		PasswordHash: hashedPassword,
	}

	if err := userRepo.Create(user); err != nil {
		return fmt.Errorf("failed to create user in database: %w", err)
	}

	fmt.Printf("✅ User created successfully!\n")
	fmt.Printf("   ID: %d\n", user.ID)
	fmt.Printf("   Username: %s\n", user.Username)
	fmt.Printf("   Email: %s\n", user.Email)
	fmt.Printf("   Created: %s\n", user.CreatedAt.Format("2006-01-02 15:04:05"))

	return nil
}