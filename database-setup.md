# Database Setup Guide

## Quick Start (Recommended)

The easiest way to set up the database is using Docker and the provided Makefile:

```bash
# Start complete infrastructure (database + migrations)
make infra

# Check status
make migrate-status

# Connect to database
make db-connect
```

## Prerequisites

- Docker and Docker Compose
- Make (usually pre-installed on macOS/Linux)

## Setup Steps

1. **Clone repository and navigate to project directory**
   ```bash
   cd /path/to/feature-voting-platform
   ```

2. **Configure environment variables (optional)**
   ```bash
   cp .env.example .env
   # Edit .env file if you want to change default values
   ```

3. **Start infrastructure**
   ```bash
   # Start database, create users, and run migrations
   make infra
   ```

4. **Verify setup**
   ```bash
   # Check migration status
   make migrate-status
   
   # Connect to database and verify data
   make db-connect
   ```

   ```sql
   -- Check tables
   \dt
   
   -- Check sample data with vote counts
   SELECT * FROM features;
   
   -- Exit
   \q
   ```

## Database Connection Details

- **Host**: localhost
- **Port**: 5432 (default)
- **Database**: feature_voting_platform
- **Username**: postgres (or voting_app if created)
- **Password**: (your password)

## Connection String Examples

- **Node.js**: `postgresql://username:password@localhost:5432/feature_voting_platform`
- **Python**: `postgresql://username:password@localhost:5432/feature_voting_platform`
- **Java**: `jdbc:postgresql://localhost:5432/feature_voting_platform`

## Useful Commands

```sql
-- View all features with vote counts
SELECT * FROM features_with_votes;

-- Get user's votes
SELECT f.title, v.created_at 
FROM votes v 
JOIN features f ON v.feature_id = f.id 
WHERE v.user_id = 1;

-- Reset database (drops all data)
DROP SCHEMA public CASCADE;
CREATE SCHEMA public;
-- Then run schema.sql again
```