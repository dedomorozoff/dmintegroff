-- Migration: 004_cascade_delete_integrations
-- Description: Change foreign key constraint to CASCADE delete integrations when project is deleted (MySQL)
-- Date: 2024-11-30

-- Drop existing foreign key constraint
ALTER TABLE integrations 
DROP FOREIGN KEY integrations_ibfk_2;

-- Add new foreign key constraint with CASCADE delete
ALTER TABLE integrations 
ADD CONSTRAINT integrations_project_id_fkey 
FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE;
