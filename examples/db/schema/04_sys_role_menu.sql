-- 系统角色与菜单关系表。
-- 设计参考基准文件中的 sys_role_menu，并保留独立主键与唯一约束，便于审计与去重。

CREATE TABLE "sys_role_menu" (
    "id" BIGINT NOT NULL,
    "role_id" BIGINT NOT NULL,
    "menu_id" BIGINT NOT NULL,
    "deleted_at" TIMESTAMP NULL,
    "created_by" BIGINT NULL,
    "created_at" TIMESTAMP NOT NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT "uk_sys_role_menu_role_id_menu_id" UNIQUE ("role_id", "menu_id"),
    CONSTRAINT "fk_sys_role_menu_role_id" FOREIGN KEY ("role_id") REFERENCES "sys_role" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
    CONSTRAINT "fk_sys_role_menu_menu_id" FOREIGN KEY ("menu_id") REFERENCES "sys_menu" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);

COMMENT ON TABLE "sys_role_menu" IS '系统角色与菜单/按钮资源关系表，用于角色授权。';
COMMENT ON COLUMN "sys_role_menu"."id" IS '雪花算法生成的主键 ID。';
COMMENT ON COLUMN "sys_role_menu"."role_id" IS '系统角色 ID，关联 sys_role.id。';
COMMENT ON COLUMN "sys_role_menu"."menu_id" IS '系统菜单资源 ID，关联 sys_menu.id。';
COMMENT ON COLUMN "sys_role_menu"."deleted_at" IS '逻辑删除时间，NULL 表示未删除。';
COMMENT ON COLUMN "sys_role_menu"."created_by" IS '创建人用户 ID。';
COMMENT ON COLUMN "sys_role_menu"."created_at" IS '创建时间。';

CREATE INDEX "idx_sys_role_menu_role_id" ON "sys_role_menu" ("role_id");
CREATE INDEX "idx_sys_role_menu_menu_id" ON "sys_role_menu" ("menu_id");
