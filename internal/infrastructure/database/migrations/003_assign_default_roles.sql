-- Assign default roles to existing users
-- This migration assigns default 'user' role to all existing users who don't have any roles

INSERT INTO user_roles (user_id, role)
SELECT u.id, 'user'
FROM users u
WHERE u.id NOT IN (
    SELECT DISTINCT user_id 
    FROM user_roles
)
ON CONFLICT (user_id, role) DO NOTHING;
