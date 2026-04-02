CREATE TABLE schedules (
    id VARCHAR(36) PRIMARY KEY,
    scheduled_at DATETIME NOT NULL,
    service_name VARCHAR(255) NOT NULL,
    environment ENUM('production', 'staging', 'development') NOT NULL,
    description TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_scheduled_at (scheduled_at),
    INDEX idx_environment (environment)
);
