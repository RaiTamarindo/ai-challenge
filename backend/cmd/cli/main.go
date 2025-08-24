package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/feature-voting-platform/backend/adapters/auth"
	"github.com/feature-voting-platform/backend/adapters/postgres"
	"github.com/feature-voting-platform/backend/domain/users"
	"github.com/feature-voting-platform/backend/internal/config"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := postgres.NewDatabase(cfg.Database.URL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize repositories and services
	userRepo := postgres.NewUserRepository(db)
	passwordService := auth.NewBCryptPasswordService()

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
		err := createUser(userRepo, passwordService, *name, *email, *password)
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

func createUser(userRepo users.Repository, passwordService auth.PasswordService, username, email, password string) error {
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

	// Check if user already exists by email
	if _, err := userRepo.GetByEmail(email); err == nil {
		return fmt.Errorf("user with email '%s' already exists", email)
	}

	// Check if user already exists by username
	if _, err := userRepo.GetByUsername(username); err == nil {
		return fmt.Errorf("user with username '%s' already exists", username)
	}

	// Hash password
	hashedPassword, err := passwordService.HashPassword(password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &users.User{
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