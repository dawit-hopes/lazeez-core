-- Add front_desk_agent and room_service_staff to users.role CHECK constraint.

ALTER TABLE users DROP CONSTRAINT IF EXISTS chk_users_role;

ALTER TABLE users ADD CONSTRAINT chk_users_role CHECK (role IN (
    'super_admin',
    'branch_manager',
    'super_branch_admin',
    'front_desk_agent',
    'room_service_staff'
));

COMMENT ON COLUMN users.role IS 'User role: super_admin, branch_manager, super_branch_admin, front_desk_agent, or room_service_staff';

ALTER TABLE users ADD COLUMN branch_type VARCHAR(50) NOT NULL DEFAULT 'restaurant';

COMMENT ON COLUMN users.branch_type IS 'Branch type: restaurant, hotel, or other';