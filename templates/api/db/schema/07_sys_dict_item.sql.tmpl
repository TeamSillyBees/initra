-- 系统字典项表。
-- 设计参考基准文件中的 sys_dict_item，并通过 collection_id + collection_code 双字段兼顾关联完整性和编码查询便利性。

CREATE TABLE "sys_dict_item" (
    "id" BIGINT NOT NULL,
    "collection_id" BIGINT NOT NULL,
    "collection_code" VARCHAR(64) NOT NULL,
    "code" VARCHAR(64) NOT NULL,
    "parent_code" VARCHAR(64) NOT NULL DEFAULT '0',
    "label" VARCHAR(128) NOT NULL,
    "is_default_value" BOOLEAN NOT NULL DEFAULT false,
    "is_enable" BOOLEAN NOT NULL DEFAULT true,
    "description" TEXT NULL,
    "sort_id" INTEGER NOT NULL DEFAULT 0,
    "deleted_at" TIMESTAMP NULL,
    "created_by" BIGINT NULL,
    "created_at" TIMESTAMP NOT NULL,
    "updated_by" BIGINT NULL,
    "updated_at" TIMESTAMP NOT NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT "uk_sys_dict_item_collection_code_code" UNIQUE ("collection_code", "code"),
    CONSTRAINT "fk_sys_dict_item_collection_id" FOREIGN KEY ("collection_id") REFERENCES "sys_dict_collection" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);

COMMENT ON TABLE "sys_dict_item" IS '系统字典项表，用于保存某个字典集下的具体值。';
COMMENT ON COLUMN "sys_dict_item"."id" IS '雪花算法生成的主键 ID。';
COMMENT ON COLUMN "sys_dict_item"."collection_id" IS '字典集 ID，关联 sys_dict_collection.id。';
COMMENT ON COLUMN "sys_dict_item"."collection_code" IS '字典集编码，便于按编码直接查询字典项。';
COMMENT ON COLUMN "sys_dict_item"."code" IS '字典项编码，程序实际使用的值。';
COMMENT ON COLUMN "sys_dict_item"."parent_code" IS '父级字典项编码，0 表示顶级节点。';
COMMENT ON COLUMN "sys_dict_item"."label" IS '字典项展示文本。';
COMMENT ON COLUMN "sys_dict_item"."is_default_value" IS '是否为默认值。';
COMMENT ON COLUMN "sys_dict_item"."is_enable" IS '字典项是否启用。';
COMMENT ON COLUMN "sys_dict_item"."description" IS '字典项描述。';
COMMENT ON COLUMN "sys_dict_item"."sort_id" IS '排序值。';
COMMENT ON COLUMN "sys_dict_item"."deleted_at" IS '逻辑删除时间，NULL 表示未删除。';
COMMENT ON COLUMN "sys_dict_item"."created_by" IS '创建人用户 ID。';
COMMENT ON COLUMN "sys_dict_item"."created_at" IS '创建时间。';
COMMENT ON COLUMN "sys_dict_item"."updated_by" IS '最后更新人用户 ID。';
COMMENT ON COLUMN "sys_dict_item"."updated_at" IS '最后更新时间。';

CREATE INDEX "idx_sys_dict_item_collection_id" ON "sys_dict_item" ("collection_id");
CREATE INDEX "idx_sys_dict_item_collection_code" ON "sys_dict_item" ("collection_code");
CREATE INDEX "idx_sys_dict_item_deleted_at" ON "sys_dict_item" ("deleted_at");
