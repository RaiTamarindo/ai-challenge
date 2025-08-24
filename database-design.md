# Database Design for Feature Voting Platform

## Technology Choice: PostgreSQL

**Rationale:**
- **ACID Compliance**: Ensures data integrity for vote operations
- **Relational Model**: Perfect for user-feature-vote relationships
- **Constraint Support**: Unique constraints for preventing duplicate votes
- **Performance**: Efficient counting and aggregation operations
- **Scalability**: Good horizontal and vertical scaling options
- **Community Support**: Mature ecosystem and extensive documentation

## Data Model

### Tables

#### 1. users
```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### 2. features
```sql
CREATE TABLE features (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    created_by INTEGER NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### 3. votes
```sql
CREATE TABLE votes (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    feature_id INTEGER NOT NULL REFERENCES features(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, feature_id) -- Ensures one vote per user per feature
);
```

### Indexes

```sql
-- Performance indexes
CREATE INDEX idx_features_created_by ON features(created_by);
CREATE INDEX idx_votes_feature_id ON votes(feature_id);
CREATE INDEX idx_votes_user_id ON votes(user_id);
CREATE INDEX idx_features_created_at ON features(created_at DESC);
```

### Views

```sql
-- View for features with vote counts
CREATE VIEW features_with_votes AS
SELECT 
    f.id,
    f.title,
    f.description,
    f.created_by,
    u.username as created_by_username,
    f.created_at,
    f.updated_at,
    COALESCE(v.vote_count, 0) as vote_count
FROM features f
LEFT JOIN users u ON f.created_by = u.id
LEFT JOIN (
    SELECT feature_id, COUNT(*) as vote_count
    FROM votes
    GROUP BY feature_id
) v ON f.id = v.feature_id
ORDER BY f.created_at DESC;
```

## Key Design Decisions

1. **Unique Constraint on Votes**: `UNIQUE(user_id, feature_id)` prevents duplicate votes
2. **Foreign Key Constraints**: Ensure referential integrity
3. **Denormalized View**: `features_with_votes` for efficient querying
4. **Serial IDs**: Auto-incrementing primary keys for simplicity
5. **Timestamps**: Track creation and modification times
6. **Indexes**: Optimize common query patterns

## Sample Queries

```sql
-- Get all features with vote counts
SELECT * FROM features_with_votes;

-- Get features created by specific user
SELECT * FROM features_with_votes WHERE created_by = ?;

-- Get vote count for specific feature
SELECT vote_count FROM features_with_votes WHERE id = ?;

-- Check if user has voted for feature
SELECT EXISTS(SELECT 1 FROM votes WHERE user_id = ? AND feature_id = ?);

-- Add a vote (will fail if duplicate due to unique constraint)
INSERT INTO votes (user_id, feature_id) VALUES (?, ?);
```

## Database Setup Script

```sql
-- Create database
CREATE DATABASE feature_voting_platform;

-- Connect to database and create tables
\c feature_voting_platform;

-- Execute table creation, indexes, and views from above
```