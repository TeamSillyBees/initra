-- 系统用户与角色关系表。
-- 相比基准文件中的 sys_user.role_id 单角色设计，这里改为多对多关系，更符合 RBAC 脚手架长期演进需求。

CREATE TABLE "sys_user_role" (
    "id" BIGINT NOT NULL,
    "user_id" BIGINT NOT NULL,
    "role_id" BIGINT NOT NULL,
    "deleted_at" TIMESTAMP NULL,
    "created_by" BIGINT NULL,
    "created_at" TIMESTAMP NOT NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT "uk_sys_user_role_user_id_role_id" UNIQUE ("user_id", "role_id"),
    CONSTRAINT "fk_sys_user_role_user_id" FOREIGN KEY ("user_id") REFERENCES "sys_user" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
    CONSTRAINT "fk_sys_user_role_role_id" FOREIGN KEY ("role_id") REFERENCES "sys_role" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);

COMMENT ON TABLE "sys_user_role" IS '系统用户与角色关系表，用于描述一个用户拥有多个角色。';
COMMENT ON COLUMN "sys_user_role"."id" IS '雪花算法生成的主键 ID。';
COMMENT ON COLUMN "sys_user_role"."user_id" IS '系统用户 ID，关联 sys_user.id。';
COMMENT ON COLUMN "sys_user_role"."role_id" IS '系统角色 ID，关联 sys_role.id。';
COMMENT ON COLUMN "sys_user_role"."deleted_at" IS '逻辑删除时间，NULL 表示未删除。';
COMMENT ON COLUMN "sys_user_role"."created_by" IS '创建人用户 ID。';
COMMENT ON COLUMN "sys_user_role"."created_at" IS '创建时间。';

CREATE INDEX "idx_sys_user_role_user_id" ON "sys_user_role" ("user_id");
CREATE INDEX "idx_sys_user_role_role_id" ON "sys_user_role" ("role_id");
