-- 系统角色表。
-- 设计参考基准文件中的 sys_role，并补充 code 唯一约束，便于程序内稳定引用。

CREATE TABLE "sys_role" (
    "id" BIGINT NOT NULL,
    "code" VARCHAR(64) NOT NULL,
    "name" VARCHAR(128) NOT NULL,
    "remark" TEXT NULL,
    "is_builtin" BOOLEAN NOT NULL DEFAULT false,
    "is_enable" BOOLEAN NOT NULL DEFAULT true,
    "sort_id" INTEGER NOT NULL DEFAULT 0,
    "deleted_at" TIMESTAMP NULL,
    "created_by" BIGINT NULL,
    "created_at" TIMESTAMP NOT NULL,
    "updated_by" BIGINT NULL,
    "updated_at" TIMESTAMP NOT NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT "uk_sys_role_code" UNIQUE ("code")
);

COMMENT ON TABLE "sys_role" IS '系统角色表，用于承载后台角色定义。';
COMMENT ON COLUMN "sys_role"."id" IS '雪花算法生成的主键 ID。';
COMMENT ON COLUMN "sys_role"."code" IS '角色编码，程序内稳定引用，例如 admin、viewer。';
COMMENT ON COLUMN "sys_role"."name" IS '角色名称，用于管理界面展示。';
COMMENT ON COLUMN "sys_role"."remark" IS '角色备注说明。';
COMMENT ON COLUMN "sys_role"."is_builtin" IS '是否为系统内置角色，内置角色通常不允许删除。';
COMMENT ON COLUMN "sys_role"."is_enable" IS '角色是否启用，禁用后不参与授权。';
COMMENT ON COLUMN "sys_role"."sort_id" IS '角色排序值，值越小越靠前。';
COMMENT ON COLUMN "sys_role"."deleted_at" IS '逻辑删除时间，NULL 表示未删除。';
COMMENT ON COLUMN "sys_role"."created_by" IS '创建人用户 ID。';
COMMENT ON COLUMN "sys_role"."created_at" IS '创建时间。';
COMMENT ON COLUMN "sys_role"."updated_by" IS '最后更新人用户 ID。';
COMMENT ON COLUMN "sys_role"."updated_at" IS '最后更新时间。';

CREATE INDEX "idx_sys_role_is_enable" ON "sys_role" ("is_enable");
CREATE INDEX "idx_sys_role_deleted_at" ON "sys_role" ("deleted_at");
