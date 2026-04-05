-- Drop indexes
DROP INDEX IF EXISTS idx_schedules_deleted_at ON schedules;
DROP INDEX IF EXISTS idx_schedules_updated_by ON schedules;
DROP INDEX IF EXISTS idx_schedules_created_by ON schedules;

-- Drop foreign key constraints
ALTER TABLE schedules
    DROP FOREIGN KEY IF EXISTS fk_schedules_deleted_by,
    DROP FOREIGN KEY IF EXISTS fk_schedules_updated_by,
    DROP FOREIGN KEY IF EXISTS fk_schedules_created_by;

-- Drop audit columns
ALTER TABLE schedules
    DROP COLUMN IF EXISTS deleted_at,
    DROP COLUMN IF EXISTS deleted_by,
    DROP COLUMN IF EXISTS updated_by,
    DROP COLUMN IF EXISTS created_by;
