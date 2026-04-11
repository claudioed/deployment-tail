ALTER TABLE schedules
    DROP INDEX idx_status,
    DROP INDEX idx_owner,
    DROP COLUMN rollback_plan,
    DROP COLUMN status,
    DROP COLUMN owner;
