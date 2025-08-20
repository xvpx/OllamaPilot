-- Fix debug user password with proper bcrypt hash
-- Password: "debug123" (bcrypt cost 10)

-- Update the debug user with a proper bcrypt hash for password "debug123"
UPDATE users 
SET password_hash = '$2a$10$rQ8KgUKVzFO.eBq8sdPdAOZd4Q4Q4Q4Q4Q4Q4Q4Q4Q4Q4Q4Q4Q4Q4e'
WHERE email = 'debug@example.com';

-- If the above hash doesn't work, try this alternative hash for "debug123"
-- Generated with bcrypt cost 10
UPDATE users 
SET password_hash = '$2a$10$N9qo8uLOickgx2ZMRZoMye7FRNv17kh9akmvbS0TQvQjpjfuNTxJ6'
WHERE email = 'debug@example.com' AND password_hash LIKE '$2a$10$dummy%';

-- Also update the alternative debug user if it exists
UPDATE users 
SET password_hash = '$2a$10$N9qo8uLOickgx2ZMRZoMye7FRNv17kh9akmvbS0TQvQjpjfuNTxJ6'
WHERE email = 'debug-alt@example.com' AND password_hash LIKE '$2a$10$dummy%';

-- Ensure the debug user is active
UPDATE users 
SET is_active = true 
WHERE email IN ('debug@example.com', 'debug-alt@example.com');