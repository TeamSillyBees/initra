-- 初始化脚手架 V1 所需的系统基础表结构。
-- 该迁移与 db/schema 目录中的 Atlas desired schema 保持一致，不再保留历史 users 示例表。

CREATE TABLE IF NOT EXISTS sys_role (
    id BIGINT PRIMARY KEY,
    code VARCHAR(64) NOT NULL UNIQUE,
    name VARCHAR(128) NOT NULL,
    remark TEXT NULL,
    is_builtin BOOLEAN NOT NULL DEFAULT false,
    is_enable BOOLEAN NOT NULL DEFAULT true,
    sort_id INTEGER NOT NULL DEFAULT 0,
    deleted_at TIMESTAMP NULL,
    created_by BIGINT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_by BIGINT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_sys_role_is_enable ON sys_role (is_enable);
CREATE INDEX IF NOT EXISTS idx_sys_role_deleted_at ON sys_role (deleted_at);

CREATE TABLE IF NOT EXISTS sys_user (
    id BIGINT PRIMARY KEY,
    username VARCHAR(64) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    nickname VARCHAR(128) NULL,
    phone VARCHAR(32) NULL UNIQUE,
    email VARCHAR(255) NULL UNIQUE,
    avatar_url TEXT NULL,
    is_super_admin BOOLEAN NOT NULL DEFAULT false,
    is_enable BOOLEAN NOT NULL DEFAULT true,
    sort_id INTEGER NOT NULL DEFAULT 0,
    deleted_at TIMESTAMP NULL,
    created_by BIGINT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_by BIGINT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_sys_user_is_enable ON sys_user (is_enable);
CREATE INDEX IF NOT EXISTS idx_sys_user_deleted_at ON sys_user (deleted_at);

CREATE TABLE IF NOT EXISTS sys_user_role (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    role_id BIGINT NOT NULL,
    deleted_at TIMESTAMP NULL,
    created_by BIGINT NULL,
    created_at TIMESTAMP NOT NULL,
    CONSTRAINT uk_sys_user_role_user_id_role_id UNIQUE (user_id, role_id),
    CONSTRAINT fk_sys_user_role_user_id FOREIGN KEY (user_id) REFERENCES sys_user (id) ON UPDATE NO ACTION ON DELETE NO ACTION,
    CONSTRAINT fk_sys_user_role_role_id FOREIGN KEY (role_id) REFERENCES sys_role (id) ON UPDATE NO ACTION ON DELETE NO ACTION
);

CREATE INDEX IF NOT EXISTS idx_sys_user_role_user_id ON sys_user_role (user_id);
CREATE INDEX IF NOT EXISTS idx_sys_user_role_role_id ON sys_user_role (role_id);

CREATE TABLE IF NOT EXISTS sys_menu (
    id BIGINT PRIMARY KEY,
    parent_id BIGINT NOT NULL DEFAULT 0,
    app_id VARCHAR(64) NULL,
    title VARCHAR(128) NOT NULL,
    menu_type SMALLINT NOT NULL,
    route_path TEXT NULL,
    component_path TEXT NULL,
    permission_code VARCHAR(128) NULL UNIQUE,
    icon VARCHAR(128) NULL,
    is_visible BOOLEAN NOT NULL DEFAULT true,
    is_cached BOOLEAN NOT NULL DEFAULT true,
    sort_id INTEGER NOT NULL DEFAULT 0,
    deleted_at TIMESTAMP NULL,
    created_by BIGINT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_by BIGINT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_sys_menu_parent_id ON sys_menu (parent_id);
CREATE INDEX IF NOT EXISTS idx_sys_menu_deleted_at ON sys_menu (deleted_at);

CREATE TABLE IF NOT EXISTS sys_role_menu (
    id BIGINT PRIMARY KEY,
    role_id BIGINT NOT NULL,
    menu_id BIGINT NOT NULL,
    deleted_at TIMESTAMP NULL,
    created_by BIGINT NULL,
    created_at TIMESTAMP NOT NULL,
    CONSTRAINT uk_sys_role_menu_role_id_menu_id UNIQUE (role_id, menu_id),
    CONSTRAINT fk_sys_role_menu_role_id FOREIGN KEY (role_id) REFERENCES sys_role (id) ON UPDATE NO ACTION ON DELETE NO ACTION,
    CONSTRAINT fk_sys_role_menu_menu_id FOREIGN KEY (menu_id) REFERENCES sys_menu (id) ON UPDATE NO ACTION ON DELETE NO ACTION
);

CREATE INDEX IF NOT EXISTS idx_sys_role_menu_role_id ON sys_role_menu (role_id);
CREATE INDEX IF NOT EXISTS idx_sys_role_menu_menu_id ON sys_role_menu (menu_id);

CREATE TABLE IF NOT EXISTS sys_config (
    id BIGINT PRIMARY KEY,
    config_key VARCHAR(128) NOT NULL UNIQUE,
    config_value TEXT NOT NULL DEFAULT '',
    config_desc TEXT NULL,
    is_builtin BOOLEAN NOT NULL DEFAULT false,
    sort_id INTEGER NOT NULL DEFAULT 0,
    deleted_at TIMESTAMP NULL,
    created_by BIGINT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_by BIGINT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_sys_config_deleted_at ON sys_config (deleted_at);

CREATE TABLE IF NOT EXISTS sys_dict_collection (
    id BIGINT PRIMARY KEY,
    code VARCHAR(64) NOT NULL UNIQUE,
    name VARCHAR(128) NOT NULL,
    is_enable BOOLEAN NOT NULL DEFAULT true,
    description TEXT NULL,
    item_length INTEGER NULL,
    is_builtin BOOLEAN NOT NULL DEFAULT false,
    sort_id INTEGER NOT NULL DEFAULT 0,
    deleted_at TIMESTAMP NULL,
    created_by BIGINT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_by BIGINT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_sys_dict_collection_deleted_at ON sys_dict_collection (deleted_at);

CREATE TABLE IF NOT EXISTS sys_dict_item (
    id BIGINT PRIMARY KEY,
    collection_id BIGINT NOT NULL,
    collection_code VARCHAR(64) NOT NULL,
    code VARCHAR(64) NOT NULL,
    parent_code VARCHAR(64) NOT NULL DEFAULT '0',
    label VARCHAR(128) NOT NULL,
    is_default_value BOOLEAN NOT NULL DEFAULT false,
    is_enable BOOLEAN NOT NULL DEFAULT true,
    description TEXT NULL,
    sort_id INTEGER NOT NULL DEFAULT 0,
    deleted_at TIMESTAMP NULL,
    created_by BIGINT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_by BIGINT NULL,
    updated_at TIMESTAMP NOT NULL,
    CONSTRAINT uk_sys_dict_item_collection_code_code UNIQUE (collection_code, code),
    CONSTRAINT fk_sys_dict_item_collection_id FOREIGN KEY (collection_id) REFERENCES sys_dict_collection (id) ON UPDATE NO ACTION ON DELETE NO ACTION
);

CREATE INDEX IF NOT EXISTS idx_sys_dict_item_collection_id ON sys_dict_item (collection_id);
CREATE INDEX IF NOT EXISTS idx_sys_dict_item_collection_code ON sys_dict_item (collection_code);
CREATE INDEX IF NOT EXISTS idx_sys_dict_item_deleted_at ON sys_dict_item (deleted_at);
