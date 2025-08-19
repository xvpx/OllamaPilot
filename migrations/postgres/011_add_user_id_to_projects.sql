-- Add user_id to projects table for ownership
ALTER TABLE projects ADD COLUMN user_id TEXT;

-- Add foreign key constraint
ALTER TABLE projects ADD CONSTRAINT fk_projects_user_id 
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Create index for better performance
CREATE INDEX idx_projects_user_id ON projects(user_id);

-- Update existing projects to belong to the debug user
UPDATE projects SET user_id = 'debug-user-id' WHERE user_id IS NULL;

-- Make user_id NOT NULL after setting default values
ALTER TABLE projects ALTER COLUMN user_id SET NOT NULL;