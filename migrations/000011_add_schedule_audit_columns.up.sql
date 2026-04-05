-- Add audit columns to schedules table
ALTER TABLE schedules
    ADD COLUMN created_by CHAR(36) NULL AFTER rollback_plan,
    ADD COLUMN updated_by CHAR(36) NULL AFTER created_by,
    ADD COLUMN deleted_by CHAR(36) NULL AFTER updated_by,
    ADD COLUMN deleted_at TIMESTAMP NULL AFTER deleted_by;

-- Add foreign key constraints to users table
-- Note: These are added after the columns to ensure users table exists
ALTER TABLE schedules
    ADD CONSTRAINT fk_schedules_created_by FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL,
    ADD CONSTRAINT fk_schedules_updated_by FOREIGN KEY (updated_by) REFERENCES users(id) ON DELETE SET NULL,
    ADD CONSTRAINT fk_schedules_deleted_by FOREIGN KEY (deleted_by) REFERENCES users(id) ON DELETE SET NULL;

-- Add indexes for audit queries
CREATE INDEX idx_schedules_created_by ON schedules(created_by);
CREATE INDEX idx_schedules_updated_by ON schedules(updated_by);
CREATE INDEX idx_schedules_deleted_at ON schedules(deleted_at);
