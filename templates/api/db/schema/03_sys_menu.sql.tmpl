-- 系统菜单与权限资源表。
-- 设计参考基准文件中的 sys_menu，并将主键统一为 BIGINT，以适配脚手架统一雪花 ID 策略。

CREATE TABLE "sys_menu" (
    "id" BIGINT NOT NULL,
    "parent_id" BIGINT NULL,
    "app_id" VARCHAR(64) NULL,
    "title" VARCHAR(128) NOT NULL,
    "menu_type" SMALLINT NOT NULL,
    "route_path" TEXT NULL,
    "component_path" TEXT NULL,
    "permission_code" VARCHAR(128) NULL,
    "icon" VARCHAR(128) NULL,
    "is_visible" BOOLEAN NOT NULL DEFAULT true,
    "is_cached" BOOLEAN NOT NULL DEFAULT true,
    "sort_id" INTEGER NOT NULL DEFAULT 0,
    "deleted_at" TIMESTAMP NULL,
    "created_by" BIGINT NULL,
    "created_at" TIMESTAMP NOT NULL,
    "updated_by" BIGINT NULL,
    "updated_at" TIMESTAMP NOT NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT "uk_sys_menu_permission_code" UNIQUE ("permission_code")
);

COMMENT ON TABLE "sys_menu" IS '系统菜单与按钮权限表，统一承载菜单、目录、按钮三类资源。';
COMMENT ON COLUMN "sys_menu"."id" IS '雪花算法生成的主键 ID。';
COMMENT ON COLUMN "sys_menu"."parent_id" IS '父级菜单 ID，NULL 表示顶级目录。';
COMMENT ON COLUMN "sys_menu"."app_id" IS '所属应用编码，用于多应用场景区分菜单树。';
COMMENT ON COLUMN "sys_menu"."title" IS '菜单或按钮展示标题。';
COMMENT ON COLUMN "sys_menu"."menu_type" IS '资源类型：0-菜单，1-按钮，2-目录。';
COMMENT ON COLUMN "sys_menu"."route_path" IS '前端路由路径。';
COMMENT ON COLUMN "sys_menu"."component_path" IS '前端组件路径。';
COMMENT ON COLUMN "sys_menu"."permission_code" IS '权限资源编码，例如 system:user:read。';
COMMENT ON COLUMN "sys_menu"."icon" IS '菜单图标标识。';
COMMENT ON COLUMN "sys_menu"."is_visible" IS '是否在前端菜单树中可见。';
COMMENT ON COLUMN "sys_menu"."is_cached" IS '前端页面是否缓存。';
COMMENT ON COLUMN "sys_menu"."sort_id" IS '排序值，值越小越靠前。';
COMMENT ON COLUMN "sys_menu"."deleted_at" IS '逻辑删除时间，NULL 表示未删除。';
COMMENT ON COLUMN "sys_menu"."created_by" IS '创建人用户 ID。';
COMMENT ON COLUMN "sys_menu"."created_at" IS '创建时间。';
COMMENT ON COLUMN "sys_menu"."updated_by" IS '最后更新人用户 ID。';
COMMENT ON COLUMN "sys_menu"."updated_at" IS '最后更新时间。';

CREATE INDEX "idx_sys_menu_parent_id" ON "sys_menu" ("parent_id");
CREATE INDEX "idx_sys_menu_deleted_at" ON "sys_menu" ("deleted_at");
