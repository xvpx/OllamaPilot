-- Projects table
CREATE TABLE projects (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Add project_id to sessions table to link sessions to projects
ALTER TABLE sessions ADD COLUMN project_id TEXT;
ALTER TABLE sessions ADD CONSTRAINT fk_sessions_project_id 
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE SET NULL;

-- Create index for better performance
CREATE INDEX idx_sessions_project_id ON sessions(project_id);
CREATE INDEX idx_projects_is_active ON projects(is_active);

-- Trigger to update projects updated_at
CREATE TRIGGER update_projects_updated_at
    BEFORE UPDATE ON projects
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Insert a default project for existing sessions
INSERT INTO projects (id, name, description, is_active) 
VALUES ('default-project', 'Default Project', 'Default project for existing sessions', true);

-- Update existing sessions to belong to the default project
UPDATE sessions SET project_id = 'default-project' WHERE project_id IS NULL;