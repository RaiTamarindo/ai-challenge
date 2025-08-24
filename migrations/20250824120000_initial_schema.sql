-- +migrate Up
-- Initial schema for feature voting platform

-- Users table
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Features table (with vote_count column for performance)
CREATE TABLE features (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    created_by INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    vote_count INTEGER DEFAULT 0 NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Votes table (junction table between users and features)
CREATE TABLE votes (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    feature_id INTEGER NOT NULL REFERENCES features(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, feature_id) -- Ensures one vote per user per feature
);

-- Performance indexes
CREATE INDEX idx_features_created_by ON features(created_by);
CREATE INDEX idx_features_vote_count ON features(vote_count DESC);
CREATE INDEX idx_votes_feature_id ON votes(feature_id);
CREATE INDEX idx_votes_user_id ON votes(user_id);
CREATE INDEX idx_features_created_at ON features(created_at DESC);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = CURRENT_TIMESTAMP;
   RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers for auto-updating updated_at
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_features_updated_at BEFORE UPDATE ON features
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Function to increment vote count
CREATE OR REPLACE FUNCTION increment_feature_vote_count()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE features 
    SET vote_count = vote_count + 1 
    WHERE id = NEW.feature_id;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Function to decrement vote count
CREATE OR REPLACE FUNCTION decrement_feature_vote_count()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE features 
    SET vote_count = vote_count - 1 
    WHERE id = OLD.feature_id;
    RETURN OLD;
END;
$$ language 'plpgsql';

-- Triggers to maintain vote_count consistency
CREATE TRIGGER vote_count_increment AFTER INSERT ON votes
    FOR EACH ROW EXECUTE FUNCTION increment_feature_vote_count();

CREATE TRIGGER vote_count_decrement AFTER DELETE ON votes
    FOR EACH ROW EXECUTE FUNCTION decrement_feature_vote_count();

-- +migrate Down
-- Drop triggers first
DROP TRIGGER IF EXISTS vote_count_decrement ON votes;
DROP TRIGGER IF EXISTS vote_count_increment ON votes;
DROP TRIGGER IF EXISTS update_features_updated_at ON features;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop functions
DROP FUNCTION IF EXISTS decrement_feature_vote_count();
DROP FUNCTION IF EXISTS increment_feature_vote_count();
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_features_created_at;
DROP INDEX IF EXISTS idx_votes_user_id;
DROP INDEX IF EXISTS idx_votes_feature_id;
DROP INDEX IF EXISTS idx_features_vote_count;
DROP INDEX IF EXISTS idx_features_created_by;

-- Drop tables (in reverse order due to foreign keys)
DROP TABLE IF EXISTS votes;
DROP TABLE IF EXISTS features;
DROP TABLE IF EXISTS users;