-- Migration script to convert existing users to RBAC system
-- This script should be run after the RBAC tables are created

-- Step 1: Create user_roles table if it doesn't exist
CREATE TABLE IF NOT EXISTS user_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Step 2: Create role_permissions table if it doesn't exist
CREATE TABLE IF NOT EXISTS role_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role VARCHAR(20) NOT NULL,
    permission VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Step 3: Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_role ON user_roles(role);
CREATE INDEX IF NOT EXISTS idx_user_roles_user_role ON user_roles(user_id, role);
CREATE INDEX IF NOT EXISTS idx_user_roles_deleted_at ON user_roles(deleted_at);

CREATE INDEX IF NOT EXISTS idx_role_permissions_role ON role_permissions(role);
CREATE INDEX IF NOT EXISTS idx_role_permissions_permission ON role_permissions(permission);
CREATE INDEX IF NOT EXISTS idx_role_permissions_role_permission ON role_permissions(role, permission);
CREATE INDEX IF NOT EXISTS idx_role_permissions_deleted_at ON role_permissions(deleted_at);

-- Step 4: Insert default role permissions
INSERT INTO role_permissions (role, permission) VALUES
    ('admin', 'user:create'),
    ('admin', 'user:read'),
    ('admin', 'user:update'),
    ('admin', 'user:delete'),
    ('admin', 'user:manage'),
    ('admin', 'system:read'),
    ('admin', 'system:update'),
    ('admin', 'system:manage'),
    ('admin', 'audit:read'),
    ('admin', 'audit:export'),
    ('admin', 'audit:manage'),
    ('admin', 'workspace:create'),
    ('admin', 'workspace:read'),
    ('admin', 'workspace:update'),
    ('admin', 'workspace:delete'),
    ('admin', 'workspace:manage'),
    ('admin', 'provider:read'),
    ('admin', 'provider:manage'),
    ('user', 'workspace:create'),
    ('user', 'workspace:read'),
    ('user', 'workspace:update'),
    ('user', 'provider:read'),
    ('viewer', 'workspace:read'),
    ('viewer', 'provider:read')
ON CONFLICT (role, permission) DO NOTHING;

-- Step 5: Migrate existing users to RBAC system
-- First user becomes admin, others become users
WITH user_rankings AS (
    SELECT 
        id,
        ROW_NUMBER() OVER (ORDER BY created_at ASC) as user_rank
    FROM users 
    WHERE deleted_at IS NULL
)
INSERT INTO user_roles (user_id, role)
SELECT 
    id,
    CASE 
        WHEN user_rank = 1 THEN 'admin'
        ELSE 'user'
    END as role
FROM user_rankings
ON CONFLICT (user_id, role) DO NOTHING;

-- Step 6: Remove the old role column from users table (optional)
-- Uncomment the following line if you want to remove the old role column
-- ALTER TABLE users DROP COLUMN IF EXISTS role;

-- Step 7: Verify migration
SELECT 
    u.username,
    u.email,
    ur.role,
    ur.created_at as role_assigned_at
FROM users u
JOIN user_roles ur ON u.id = ur.user_id
WHERE u.deleted_at IS NULL
ORDER BY u.created_at;

-- Step 8: Show role distribution
SELECT 
    role,
    COUNT(*) as user_count
FROM user_roles ur
JOIN users u ON ur.user_id = u.id
WHERE u.deleted_at IS NULL
GROUP BY role
ORDER BY role;
