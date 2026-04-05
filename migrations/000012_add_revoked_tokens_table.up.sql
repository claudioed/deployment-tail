CREATE TABLE revoked_tokens (
    token_hash CHAR(64) PRIMARY KEY COMMENT 'SHA256 hash of the token',
    user_id CHAR(36) NOT NULL,
    revoked_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL COMMENT 'When the token would naturally expire',
    INDEX idx_user_id (user_id),
    INDEX idx_expires_at (expires_at),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Blacklist of revoked JWT tokens';
