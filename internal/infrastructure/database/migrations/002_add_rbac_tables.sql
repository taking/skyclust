-- Add RBAC tables migration
-- This migration adds role-based access control tables

-- User roles table
CREATE TABLE IF NOT EXISTS user_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Role permissions table
CREATE TABLE IF NOT EXISTS role_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role VARCHAR(20) NOT NULL,
    permission VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Create indexes for better performance
CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_user_roles_role ON user_roles(role);
CREATE INDEX idx_user_roles_user_role ON user_roles(user_id, role);

CREATE INDEX idx_role_permissions_role ON role_permissions(role);
CREATE INDEX idx_role_permissions_permission ON role_permissions(permission);
CREATE INDEX idx_role_permissions_role_permission ON role_permissions(role, permission);

-- Insert default role permissions
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
