-- 系统配置表。
-- 设计参考基准文件中的 sys_config，并增加唯一 key 约束，保证配置项可稳定读取。

CREATE TABLE "sys_config" (
    "id" BIGINT NOT NULL,
    "config_key" VARCHAR(128) NOT NULL,
    "config_value" TEXT NOT NULL DEFAULT '',
    "config_desc" TEXT NULL,
    "is_builtin" BOOLEAN NOT NULL DEFAULT false,
    "sort_id" INTEGER NOT NULL DEFAULT 0,
    "deleted_at" TIMESTAMP NULL,
    "created_by" BIGINT NULL,
    "created_at" TIMESTAMP NOT NULL,
    "updated_by" BIGINT NULL,
    "updated_at" TIMESTAMP NOT NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT "uk_sys_config_config_key" UNIQUE ("config_key")
);

COMMENT ON TABLE "sys_config" IS '系统配置表，用于集中存放可在后台维护的运行时配置。';
COMMENT ON COLUMN "sys_config"."id" IS '雪花算法生成的主键 ID。';
COMMENT ON COLUMN "sys_config"."config_key" IS '配置键，程序通过该键读取配置。';
COMMENT ON COLUMN "sys_config"."config_value" IS '配置值。';
COMMENT ON COLUMN "sys_config"."config_desc" IS '配置项描述。';
COMMENT ON COLUMN "sys_config"."is_builtin" IS '是否为系统内置配置，内置配置通常不允许删除。';
COMMENT ON COLUMN "sys_config"."sort_id" IS '排序值。';
COMMENT ON COLUMN "sys_config"."deleted_at" IS '逻辑删除时间，NULL 表示未删除。';
COMMENT ON COLUMN "sys_config"."created_by" IS '创建人用户 ID。';
COMMENT ON COLUMN "sys_config"."created_at" IS '创建时间。';
COMMENT ON COLUMN "sys_config"."updated_by" IS '最后更新人用户 ID。';
COMMENT ON COLUMN "sys_config"."updated_at" IS '最后更新时间。';

CREATE INDEX "idx_sys_config_deleted_at" ON "sys_config" ("deleted_at");
