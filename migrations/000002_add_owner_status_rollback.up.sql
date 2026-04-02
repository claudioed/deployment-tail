-- Add owner, status, and rollback_plan columns to schedules table
ALTER TABLE schedules
    ADD COLUMN owner VARCHAR(255) NOT NULL DEFAULT 'system',
    ADD COLUMN status ENUM('created', 'approved', 'denied') NOT NULL DEFAULT 'approved',
    ADD COLUMN rollback_plan TEXT;

-- Add indexes for filtering
CREATE INDEX idx_owner ON schedules(owner);
CREATE INDEX idx_status ON schedules(status);
