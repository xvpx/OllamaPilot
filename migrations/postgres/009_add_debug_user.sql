-- Add debug user for testing purposes
INSERT INTO users (id, username, email, password_hash, created_at, updated_at, is_active)
VALUES (
    'debug-user-id',
    'debug-user',
    'debug@example.com',
    '$2a$10$dummy.hash.for.debug.user.only.not.real.password',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP,
    true
) ON CONFLICT (id) DO NOTHING;

-- Also handle conflict on username and email in case they exist
INSERT INTO users (id, username, email, password_hash, created_at, updated_at, is_active)
SELECT
    'debug-user-id-alt',
    'debug-user-alt',
    'debug-alt@example.com',
    '$2a$10$dummy.hash.for.debug.user.only.not.real.password',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP,
    true
WHERE NOT EXISTS (SELECT 1 FROM users WHERE id = 'debug-user-id');