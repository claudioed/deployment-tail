ALTER TABLE schedules
    ADD COLUMN owner VARCHAR(255) NOT NULL DEFAULT 'system',
    ADD COLUMN status ENUM('created', 'approved', 'denied') NOT NULL DEFAULT 'approved',
    ADD COLUMN rollback_plan TEXT,
    ADD INDEX idx_owner (owner),
    ADD INDEX idx_status (status);
