-- 系统字典集表。
-- 设计参考基准文件中的 sys_dict_collection，用于定义一组字典项的集合元数据。

CREATE TABLE "sys_dict_collection" (
    "id" BIGINT NOT NULL,
    "code" VARCHAR(64) NOT NULL,
    "name" VARCHAR(128) NOT NULL,
    "is_enable" BOOLEAN NOT NULL DEFAULT true,
    "description" TEXT NULL,
    "item_length" INTEGER NULL,
    "is_builtin" BOOLEAN NOT NULL DEFAULT false,
    "sort_id" INTEGER NOT NULL DEFAULT 0,
    "deleted_at" TIMESTAMP NULL,
    "created_by" BIGINT NULL,
    "created_at" TIMESTAMP NOT NULL,
    "updated_by" BIGINT NULL,
    "updated_at" TIMESTAMP NOT NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT "uk_sys_dict_collection_code" UNIQUE ("code")
);

COMMENT ON TABLE "sys_dict_collection" IS '系统字典集表，用于定义一类字典的元信息。';
COMMENT ON COLUMN "sys_dict_collection"."id" IS '雪花算法生成的主键 ID。';
COMMENT ON COLUMN "sys_dict_collection"."code" IS '字典集唯一编码，程序通过该编码读取字典项。';
COMMENT ON COLUMN "sys_dict_collection"."name" IS '字典集名称。';
COMMENT ON COLUMN "sys_dict_collection"."is_enable" IS '字典集是否启用。';
COMMENT ON COLUMN "sys_dict_collection"."description" IS '字典集说明。';
COMMENT ON COLUMN "sys_dict_collection"."item_length" IS '字典值推荐长度上限。';
COMMENT ON COLUMN "sys_dict_collection"."is_builtin" IS '是否为系统内置字典集。';
COMMENT ON COLUMN "sys_dict_collection"."sort_id" IS '排序值。';
COMMENT ON COLUMN "sys_dict_collection"."deleted_at" IS '逻辑删除时间，NULL 表示未删除。';
COMMENT ON COLUMN "sys_dict_collection"."created_by" IS '创建人用户 ID。';
COMMENT ON COLUMN "sys_dict_collection"."created_at" IS '创建时间。';
COMMENT ON COLUMN "sys_dict_collection"."updated_by" IS '最后更新人用户 ID。';
COMMENT ON COLUMN "sys_dict_collection"."updated_at" IS '最后更新时间。';

CREATE INDEX "idx_sys_dict_collection_deleted_at" ON "sys_dict_collection" ("deleted_at");
