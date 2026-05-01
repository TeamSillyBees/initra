-- 默认管理员账号: admin / admin123
-- 该种子数据同时准备 admin、viewer 两个基础角色，便于 user 模块默认角色逻辑直接工作。

INSERT INTO sys_role (
    id,
    code,
    name,
    remark,
    is_builtin,
    is_enable,
    sort_id,
    created_at,
    updated_at,
    created_by,
    updated_by
) VALUES
(
    1000000000101,
    'admin',
    '系统管理员',
    '默认超级管理员角色',
    true,
    true,
    0,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP,
    0,
    0
),
(
    1000000000102,
    'viewer',
    '只读用户',
    '默认基础只读角色',
    true,
    true,
    10,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP,
    0,
    0
)
ON CONFLICT (code) DO NOTHING;

INSERT INTO sys_user (
    id,
    username,
    password_hash,
    nickname,
    phone,
    email,
    avatar_url,
    is_super_admin,
    is_enable,
    sort_id,
    created_at,
    updated_at,
    created_by,
    updated_by
) VALUES (
    1000000000001,
    'admin',
    '$2a$10$2JNGca7Fqsq/IAvSmG2QC.XDzt5VRO9ofixT0jmZsnAic6pzBnV7C',
    'System Admin',
    NULL,
    'admin@example.com',
    NULL,
    true,
    true,
    0,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP,
    0,
    0
)
ON CONFLICT (username) DO NOTHING;

INSERT INTO sys_user_role (
    id,
    user_id,
    role_id,
    created_by,
    created_at
)
SELECT
    1000000000201,
    su.id,
    sr.id,
    0,
    CURRENT_TIMESTAMP
FROM sys_user AS su
INNER JOIN sys_role AS sr ON sr.code = 'admin'
WHERE su.username = 'admin'
ON CONFLICT (user_id, role_id) DO NOTHING;
