-- +migrate Up
UPDATE features 
SET vote_count = (
    SELECT COALESCE(COUNT(*), 0) 
    FROM votes 
    WHERE votes.feature_id = features.id
);

-- +migrate Down
SELECT 'do not rollback' AS message;