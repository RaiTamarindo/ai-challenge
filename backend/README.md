# Feature Voting Platform Backend

A Go-based REST API backend for a feature voting platform, built with clean architecture principles using domain-driven design and the adapter pattern.

## Architecture

This project follows clean architecture principles with a clear separation between domain logic and infrastructure concerns:

- **Domain Layer** (`domain/`): Contains business entities and repository interfaces
  - `domain/users/`: User entities and repository interface
  - `domain/features/`: Feature entities and repository interface  
  - `domain/votes/`: Vote entities and repository interface

- **Adapter Layer** (`adapters/`): Contains infrastructure implementations
  - `adapters/postgres/`: Database implementations of repositories
  - `adapters/auth/`: Authentication and password services
  - `adapters/rest/`: HTTP handlers and middleware
  - `adapters/logs/`: Structured logging implementation

## Features

- User registration and authentication with JWT tokens
- Feature creation, reading, updating, and deletion
- Feature voting system
- User vote history
- Structured JSON logging
- Database migrations
- Comprehensive unit testing with mocks
- Docker support

## Tech Stack

- **Language**: Go 1.21+
- **Web Framework**: Gin
- **Database**: PostgreSQL
- **Authentication**: JWT tokens with bcrypt password hashing
- **Logging**: Structured JSON logging with logrus
- **Testing**: Testify with mockery for mocking
- **Database**: sql-migrate for migrations
- **Containerization**: Docker

## Quick Start

### Prerequisites

- Go 1.21 or higher
- PostgreSQL 13+
- Docker (optional)
- migrate CLI tool for database migrations

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd backend
```

2. Install dependencies:
```bash
make deps
```

3. Set up environment variables:
```bash
export DATABASE_URL="postgres://username:password@localhost/feature_voting?sslmode=disable"
export JWT_SECRET="your-secret-key"
export PORT="8080"
```

4. Run database migrations:
```bash
make migrate-up
```

5. Build and run the application:
```bash
make build
./bin/api
```

Or run directly:
```bash
make run
```

### Docker Setup

Build and run with Docker:
```bash
make docker-build
make docker-run
```

## Development

### Running Tests

The project includes comprehensive unit tests with mocking. Use the following commands:

```bash
# Run all tests
make test

# Run tests with verbose output and race detection
make test-verbose

# Run tests with coverage report
make test-coverage
```

The coverage report will be generated as `coverage.html` in the project root.

### Mock Generation

This project uses [mockery](https://github.com/vektra/mockery) to generate mocks from interfaces. The configuration is defined in `.mockery.yaml`.

To regenerate mocks:
```bash
make generate-mocks
```

### Adding New Interfaces for Mocking

When you add new interfaces that need mocking:

1. Add the interface to the appropriate section in `.mockery.yaml`
2. Run `make generate-mocks` to generate the new mocks
3. The mocks will be generated in the `mocks/` subdirectory of each package

Example `.mockery.yaml` configuration:
```yaml
packages:
  github.com/feature-voting-platform/backend/domain/users:
    interfaces:
      Repository:
        config:
          dir: "{{.InterfaceDir}}/mocks"
```

### Project Structure

```
.
├── cmd/api/                    # Application entry point
│   └── main.go
├── domain/                     # Domain layer (business logic)
│   ├── features/              # Feature domain
│   │   ├── feature.go         # Feature entity
│   │   ├── repository.go      # Repository interface
│   │   └── mocks/             # Generated mocks
│   ├── users/                 # User domain
│   │   ├── user.go           # User entity
│   │   ├── repository.go     # Repository interface
│   │   └── mocks/            # Generated mocks
│   └── votes/                 # Vote domain
│       ├── vote.go           # Vote entity
│       ├── repository.go     # Repository interface
│       └── mocks/            # Generated mocks
├── adapters/                   # Adapter layer (infrastructure)
│   ├── auth/                  # Authentication services
│   │   ├── jwt.go            # JWT token service
│   │   ├── jwt_test.go       # JWT tests
│   │   └── mocks/            # Generated mocks
│   ├── logs/                  # Logging adapter
│   │   ├── logger.go         # Logger implementation
│   │   └── mocks/            # Generated mocks
│   ├── postgres/              # Database adapters
│   │   ├── database.go       # Database connection
│   │   ├── user_repository.go           # User repository implementation
│   │   ├── user_repository_test.go      # User repository tests
│   │   ├── feature_repository.go       # Feature repository implementation
│   │   └── feature_repository_test.go  # Feature repository tests
│   └── rest/                  # HTTP adapters
│       ├── auth_handler.go       # Authentication handlers
│       ├── auth_handler_test.go  # Auth handler tests
│       ├── feature_handler.go    # Feature handlers
│       ├── feature_handler_test.go # Feature handler tests
│       ├── vote_handler.go       # Vote handlers
│       ├── vote_handler_test.go  # Vote handler tests
│       └── middleware.go         # HTTP middleware
├── migrations/                 # Database migrations
├── Makefile                   # Build and development commands
├── .mockery.yaml             # Mockery configuration
├── go.mod                    # Go modules
├── go.sum                    # Go modules checksums
└── README.md                 # This file
```

### Testing Strategy

The project uses a comprehensive testing strategy:

1. **Unit Tests**: Each component is tested in isolation using mocks
2. **Repository Tests**: Database repositories are tested using sqlmock
3. **Handler Tests**: HTTP handlers are tested using Gin's test context
4. **Service Tests**: Authentication services are tested with real implementations

Key testing libraries:
- **testify**: Assertions and test suites
- **mockery**: Mock generation from interfaces
- **sqlmock**: SQL database mocking
- **gin**: HTTP testing utilities

### API Endpoints

#### Authentication
- `POST /auth/register` - User registration
- `POST /auth/login` - User login

#### Features
- `GET /features` - List features (with pagination)
- `POST /features` - Create new feature (authenticated)
- `GET /features/:id` - Get feature by ID
- `PUT /features/:id` - Update feature (authenticated, creator only)
- `DELETE /features/:id` - Delete feature (authenticated, creator only)

#### Voting
- `POST /features/:id/vote` - Vote for a feature (authenticated)
- `GET /votes` - Get user's vote history (authenticated)

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | Required |
| `JWT_SECRET` | Secret key for JWT token signing | Required |
| `PORT` | Server port | `8080` |
| `LOG_LEVEL` | Log level (debug, info, warn, error) | `info` |

### Database Schema

The application uses the following main tables:
- `users`: User accounts and authentication
- `features`: Feature requests and descriptions  
- `votes`: User votes for features

See the `migrations/` directory for detailed schema definitions.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass: `make test`
5. Generate mocks if interfaces changed: `make generate-mocks`
6. Submit a pull request

## Development Commands

Use the Makefile for common development tasks:

```bash
make help          # Show available commands
make dev           # Full development setup (deps, mocks, test, build)
make ci            # CI pipeline (deps, test with coverage)  
make clean         # Clean build artifacts
```

## License

This project is licensed under the MIT License.