# Feature Voting Platform

A full-stack feature voting platform where users can propose features, list all features, and vote on features they think would be good to have. Built with Go backend API, PostgreSQL database, and designed for native Android frontend.

## 🏗️ Architecture

- **Database**: PostgreSQL with optimized schema and vote counting
- **Backend**: Go REST API with Gin framework, JWT authentication
- **Frontend**: Native Android (Stage 3)
- **Infrastructure**: Docker-based development environment

## 🚀 Quick Start

### Prerequisites
- Docker and Docker Compose
- Make (usually pre-installed on macOS/Linux)

### Start Development Environment

```bash
# Clone and navigate to project
git clone <repository-url>
cd feature-voting-platform

# Start complete development environment
make dev

# This will:
# 1. Start PostgreSQL database
# 2. Run database migrations
# 3. Create database users
# 4. Start the API server
```

### API Endpoints

The API will be available at: `http://localhost:8080`

- **Health Check**: `GET /health`
- **Swagger Documentation**: `GET /swagger/index.html`
- **API Base**: `/api/v1`

## 📋 API Documentation

### Authentication Endpoints

```bash
# Register a new user
POST /api/v1/auth/register
{
  "username": "john_doe",
  "email": "john@example.com", 
  "password": "securepassword"
}

# Login
POST /api/v1/auth/login
{
  "email": "john@example.com",
  "password": "securepassword"
}

# Get user profile (requires JWT token)
GET /api/v1/auth/profile
Authorization: Bearer <token>
```

### Feature Endpoints

```bash
# Create a feature (requires authentication)
POST /api/v1/features
Authorization: Bearer <token>
{
  "title": "Dark Mode",
  "description": "Add dark mode theme to the application"
}

# Get all features (optional authentication for vote status)
GET /api/v1/features?page=1&per_page=10

# Get specific feature
GET /api/v1/features/1

# Update feature (only by creator)
PUT /api/v1/features/1
Authorization: Bearer <token>
{
  "title": "Updated title",
  "description": "Updated description"
}

# Delete feature (only by creator)
DELETE /api/v1/features/1
Authorization: Bearer <token>

# Get user's features
GET /api/v1/features/my
Authorization: Bearer <token>
```

### Voting Endpoints

```bash
# Vote for a feature
POST /api/v1/features/1/vote
Authorization: Bearer <token>

# Remove vote from a feature
DELETE /api/v1/features/1/vote
Authorization: Bearer <token>

# Toggle vote (add if not voted, remove if voted)
POST /api/v1/features/1/toggle-vote
Authorization: Bearer <token>

# Get user's votes
GET /api/v1/votes/my
Authorization: Bearer <token>
```

## 🛠️ Development Commands

```bash
# Infrastructure management
make infra          # Start database and run migrations
make infra-up       # Start database only
make infra-down     # Stop infrastructure
make infra-logs     # Show database logs
make infra-clean    # Clean up (removes data!)

# Database migrations
make migrate-up     # Run migrations
make migrate-down   # Rollback last migration
make migrate-status # Check migration status
make migration name=migration_name  # Create new migration

# Database connections
make db-connect     # Connect as admin user
make db-connect-app # Connect as application user

# API management
make api-up         # Start API service
make api-down       # Stop API service  
make api-logs       # Show API logs

# Development workflow
make dev            # Start complete development environment
make dev-reset      # Reset development environment

# Utilities
make env            # Show environment variables
make help           # Show all commands
```

## 🗄️ Database Schema

### Tables

- **users**: User accounts with authentication
- **features**: Feature requests with vote counts
- **votes**: User votes on features (unique constraint)

### Key Features

- **Automatic vote counting**: Database triggers maintain vote_count
- **Unique vote constraint**: One vote per user per feature
- **Optimized queries**: Indexes for performance
- **User management**: Separate admin/app database users

## 🔧 Configuration

Environment variables (see `.env.example`):

```bash
# Database
POSTGRES_HOSTNAME=localhost
POSTGRES_PORT=5432
POSTGRES_ADMIN_USERNAME=postgres
POSTGRES_ADMIN_PASSWORD=postgres_admin_pass
POSTGRES_STANDARD_USERNAME=voting_app
POSTGRES_STANDARD_PASSWORD=voting_app_pass
POSTGRES_DB=feature_voting_platform

# API
API_PORT=8080
APP_ENV=development
JWT_SECRET=your_jwt_secret_change_in_production
```

## 🏗️ Project Structure

```
├── backend/
│   ├── cmd/
│   │   ├── api/main.go           # API server
│   │   └── migrate/main.go       # Migration tool
│   ├── internal/
│   │   ├── config/               # Configuration
│   │   ├── handlers/             # HTTP handlers
│   │   ├── middleware/           # HTTP middleware
│   │   ├── models/               # Data models
│   │   └── repository/           # Database layer
│   ├── pkg/utils/                # Utilities
│   └── docs/                     # Swagger docs
├── migrations/                   # Database migrations
├── Dockerfile                    # Multi-stage Docker build
├── docker-compose.yaml           # Infrastructure setup
└── Makefile                      # Development commands
```

## 🔒 Security Features

- **JWT Authentication**: Secure token-based auth
- **Password hashing**: bcrypt with salt
- **CORS support**: Configurable CORS middleware
- **Input validation**: Request validation and sanitization
- **Non-root containers**: Security-first Docker images
- **Environment-based config**: No hardcoded secrets

## 📊 Key Features Implemented

### ✅ Functional Requirements
1. ✅ Users can propose features
2. ✅ Users can list all features 
3. ✅ Users can vote on features
4. ✅ Features track unique user vote counts

### ✅ Non-functional Requirements
1. ✅ PostgreSQL database with optimized schema
2. ✅ Go backend API with comprehensive endpoints
3. 🔄 Native Android frontend (Stage 3)

## 🚀 Next Steps

**Stage 3**: Native Android Frontend
- Android project with modern architecture (MVVM)
- User authentication flows
- Feature listing and creation UI
- Voting functionality with real-time updates
- API integration with proper error handling

## 📝 API Response Examples

### Feature List Response
```json
{
  "features": [
    {
      "id": 1,
      "title": "Dark Mode",
      "description": "Add dark mode theme",
      "created_by": 1,
      "created_by_username": "john_doe",
      "vote_count": 5,
      "has_user_voted": true,
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    }
  ],
  "total": 10,
  "page": 1,
  "per_page": 10
}
```

### Vote Response
```json
{
  "message": "Vote added successfully",
  "feature_id": 1,
  "vote_count": 6,
  "has_voted": true
}
```