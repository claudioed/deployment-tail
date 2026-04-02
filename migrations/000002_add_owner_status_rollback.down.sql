-- Remove owner, status, and rollback_plan columns from schedules table
DROP INDEX idx_status ON schedules;
DROP INDEX idx_owner ON schedules;

ALTER TABLE schedules
    DROP COLUMN rollback_plan,
    DROP COLUMN status,
    DROP COLUMN owner;
