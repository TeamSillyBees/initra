-- 系统用户表。
-- 设计参考基准文件中的 sys_user，并将逻辑删除统一为 deleted_at，以贴合脚手架审计字段约定。

CREATE TABLE "sys_user" (
    "id" BIGINT NOT NULL,
    "username" VARCHAR(64) NOT NULL,
    "password_hash" TEXT NOT NULL,
    "nickname" VARCHAR(128) NULL,
    "phone" VARCHAR(32) NULL,
    "email" VARCHAR(255) NULL,
    "avatar_url" TEXT NULL,
    "is_super_admin" BOOLEAN NOT NULL DEFAULT false,
    "is_enable" BOOLEAN NOT NULL DEFAULT true,
    "sort_id" INTEGER NOT NULL DEFAULT 0,
    "deleted_at" TIMESTAMP NULL,
    "created_by" BIGINT NULL,
    "created_at" TIMESTAMP NOT NULL,
    "updated_by" BIGINT NULL,
    "updated_at" TIMESTAMP NOT NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT "uk_sys_user_username" UNIQUE ("username"),
    CONSTRAINT "uk_sys_user_phone" UNIQUE ("phone"),
    CONSTRAINT "uk_sys_user_email" UNIQUE ("email")
);

COMMENT ON TABLE "sys_user" IS '系统后台用户表，用于后台登录、审计和权限归属。';
COMMENT ON COLUMN "sys_user"."id" IS '雪花算法生成的主键 ID。';
COMMENT ON COLUMN "sys_user"."username" IS '登录用户名，全局唯一。';
COMMENT ON COLUMN "sys_user"."password_hash" IS '经过安全哈希后的密码密文。';
COMMENT ON COLUMN "sys_user"."nickname" IS '用户昵称或显示名。';
COMMENT ON COLUMN "sys_user"."phone" IS '手机号，可用于登录或通知。';
COMMENT ON COLUMN "sys_user"."email" IS '邮箱地址，可用于找回密码和通知。';
COMMENT ON COLUMN "sys_user"."avatar_url" IS '头像资源地址。';
COMMENT ON COLUMN "sys_user"."is_super_admin" IS '是否为超级管理员，超级管理员通常拥有全量权限。';
COMMENT ON COLUMN "sys_user"."is_enable" IS '账号是否启用。';
COMMENT ON COLUMN "sys_user"."sort_id" IS '排序值，便于后台列表定制顺序。';
COMMENT ON COLUMN "sys_user"."deleted_at" IS '逻辑删除时间，NULL 表示未删除。';
COMMENT ON COLUMN "sys_user"."created_by" IS '创建人用户 ID。';
COMMENT ON COLUMN "sys_user"."created_at" IS '创建时间。';
COMMENT ON COLUMN "sys_user"."updated_by" IS '最后更新人用户 ID。';
COMMENT ON COLUMN "sys_user"."updated_at" IS '最后更新时间。';

CREATE INDEX "idx_sys_user_is_enable" ON "sys_user" ("is_enable");
CREATE INDEX "idx_sys_user_deleted_at" ON "sys_user" ("deleted_at");
