-- Create "sys_config" table
CREATE TABLE "sys_config" (
  "id" bigint NOT NULL,
  "config_key" character varying(128) NOT NULL,
  "config_value" text NOT NULL DEFAULT '',
  "config_desc" text NULL,
  "is_builtin" boolean NOT NULL DEFAULT false,
  "sort_id" integer NOT NULL DEFAULT 0,
  "deleted_at" timestamp NULL,
  "created_by" bigint NULL,
  "created_at" timestamp NOT NULL,
  "updated_by" bigint NULL,
  "updated_at" timestamp NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uk_sys_config_config_key" UNIQUE ("config_key")
);
-- Create index "idx_sys_config_deleted_at" to table: "sys_config"
CREATE INDEX "idx_sys_config_deleted_at" ON "sys_config" ("deleted_at");
-- Set comment to table: "sys_config"
COMMENT ON TABLE "sys_config" IS '系统配置表，用于集中存放可在后台维护的运行时配置。';
-- Set comment to column: "id" on table: "sys_config"
COMMENT ON COLUMN "sys_config"."id" IS '雪花算法生成的主键 ID。';
-- Set comment to column: "config_key" on table: "sys_config"
COMMENT ON COLUMN "sys_config"."config_key" IS '配置键，程序通过该键读取配置。';
-- Set comment to column: "config_value" on table: "sys_config"
COMMENT ON COLUMN "sys_config"."config_value" IS '配置值。';
-- Set comment to column: "config_desc" on table: "sys_config"
COMMENT ON COLUMN "sys_config"."config_desc" IS '配置项描述。';
-- Set comment to column: "is_builtin" on table: "sys_config"
COMMENT ON COLUMN "sys_config"."is_builtin" IS '是否为系统内置配置，内置配置通常不允许删除。';
-- Set comment to column: "sort_id" on table: "sys_config"
COMMENT ON COLUMN "sys_config"."sort_id" IS '排序值。';
-- Set comment to column: "deleted_at" on table: "sys_config"
COMMENT ON COLUMN "sys_config"."deleted_at" IS '逻辑删除时间，NULL 表示未删除。';
-- Set comment to column: "created_by" on table: "sys_config"
COMMENT ON COLUMN "sys_config"."created_by" IS '创建人用户 ID。';
-- Set comment to column: "created_at" on table: "sys_config"
COMMENT ON COLUMN "sys_config"."created_at" IS '创建时间。';
-- Set comment to column: "updated_by" on table: "sys_config"
COMMENT ON COLUMN "sys_config"."updated_by" IS '最后更新人用户 ID。';
-- Set comment to column: "updated_at" on table: "sys_config"
COMMENT ON COLUMN "sys_config"."updated_at" IS '最后更新时间。';
-- Create "sys_dict_collection" table
CREATE TABLE "sys_dict_collection" (
  "id" bigint NOT NULL,
  "code" character varying(64) NOT NULL,
  "name" character varying(128) NOT NULL,
  "is_enable" boolean NOT NULL DEFAULT true,
  "description" text NULL,
  "item_length" integer NULL,
  "is_builtin" boolean NOT NULL DEFAULT false,
  "sort_id" integer NOT NULL DEFAULT 0,
  "deleted_at" timestamp NULL,
  "created_by" bigint NULL,
  "created_at" timestamp NOT NULL,
  "updated_by" bigint NULL,
  "updated_at" timestamp NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uk_sys_dict_collection_code" UNIQUE ("code")
);
-- Create index "idx_sys_dict_collection_deleted_at" to table: "sys_dict_collection"
CREATE INDEX "idx_sys_dict_collection_deleted_at" ON "sys_dict_collection" ("deleted_at");
-- Set comment to table: "sys_dict_collection"
COMMENT ON TABLE "sys_dict_collection" IS '系统字典集表，用于定义一类字典的元信息。';
-- Set comment to column: "id" on table: "sys_dict_collection"
COMMENT ON COLUMN "sys_dict_collection"."id" IS '雪花算法生成的主键 ID。';
-- Set comment to column: "code" on table: "sys_dict_collection"
COMMENT ON COLUMN "sys_dict_collection"."code" IS '字典集唯一编码，程序通过该编码读取字典项。';
-- Set comment to column: "name" on table: "sys_dict_collection"
COMMENT ON COLUMN "sys_dict_collection"."name" IS '字典集名称。';
-- Set comment to column: "is_enable" on table: "sys_dict_collection"
COMMENT ON COLUMN "sys_dict_collection"."is_enable" IS '字典集是否启用。';
-- Set comment to column: "description" on table: "sys_dict_collection"
COMMENT ON COLUMN "sys_dict_collection"."description" IS '字典集说明。';
-- Set comment to column: "item_length" on table: "sys_dict_collection"
COMMENT ON COLUMN "sys_dict_collection"."item_length" IS '字典值推荐长度上限。';
-- Set comment to column: "is_builtin" on table: "sys_dict_collection"
COMMENT ON COLUMN "sys_dict_collection"."is_builtin" IS '是否为系统内置字典集。';
-- Set comment to column: "sort_id" on table: "sys_dict_collection"
COMMENT ON COLUMN "sys_dict_collection"."sort_id" IS '排序值。';
-- Set comment to column: "deleted_at" on table: "sys_dict_collection"
COMMENT ON COLUMN "sys_dict_collection"."deleted_at" IS '逻辑删除时间，NULL 表示未删除。';
-- Set comment to column: "created_by" on table: "sys_dict_collection"
COMMENT ON COLUMN "sys_dict_collection"."created_by" IS '创建人用户 ID。';
-- Set comment to column: "created_at" on table: "sys_dict_collection"
COMMENT ON COLUMN "sys_dict_collection"."created_at" IS '创建时间。';
-- Set comment to column: "updated_by" on table: "sys_dict_collection"
COMMENT ON COLUMN "sys_dict_collection"."updated_by" IS '最后更新人用户 ID。';
-- Set comment to column: "updated_at" on table: "sys_dict_collection"
COMMENT ON COLUMN "sys_dict_collection"."updated_at" IS '最后更新时间。';
-- Create "sys_dict_item" table
CREATE TABLE "sys_dict_item" (
  "id" bigint NOT NULL,
  "collection_id" bigint NOT NULL,
  "collection_code" character varying(64) NOT NULL,
  "code" character varying(64) NOT NULL,
  "parent_code" character varying(64) NOT NULL DEFAULT '0',
  "label" character varying(128) NOT NULL,
  "is_default_value" boolean NOT NULL DEFAULT false,
  "is_enable" boolean NOT NULL DEFAULT true,
  "description" text NULL,
  "sort_id" integer NOT NULL DEFAULT 0,
  "deleted_at" timestamp NULL,
  "created_by" bigint NULL,
  "created_at" timestamp NOT NULL,
  "updated_by" bigint NULL,
  "updated_at" timestamp NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uk_sys_dict_item_collection_code_code" UNIQUE ("collection_code", "code"),
  CONSTRAINT "fk_sys_dict_item_collection_id" FOREIGN KEY ("collection_id") REFERENCES "sys_dict_collection" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_sys_dict_item_collection_code" to table: "sys_dict_item"
CREATE INDEX "idx_sys_dict_item_collection_code" ON "sys_dict_item" ("collection_code");
-- Create index "idx_sys_dict_item_collection_id" to table: "sys_dict_item"
CREATE INDEX "idx_sys_dict_item_collection_id" ON "sys_dict_item" ("collection_id");
-- Create index "idx_sys_dict_item_deleted_at" to table: "sys_dict_item"
CREATE INDEX "idx_sys_dict_item_deleted_at" ON "sys_dict_item" ("deleted_at");
-- Set comment to table: "sys_dict_item"
COMMENT ON TABLE "sys_dict_item" IS '系统字典项表，用于保存某个字典集下的具体值。';
-- Set comment to column: "id" on table: "sys_dict_item"
COMMENT ON COLUMN "sys_dict_item"."id" IS '雪花算法生成的主键 ID。';
-- Set comment to column: "collection_id" on table: "sys_dict_item"
COMMENT ON COLUMN "sys_dict_item"."collection_id" IS '字典集 ID，关联 sys_dict_collection.id。';
-- Set comment to column: "collection_code" on table: "sys_dict_item"
COMMENT ON COLUMN "sys_dict_item"."collection_code" IS '字典集编码，便于按编码直接查询字典项。';
-- Set comment to column: "code" on table: "sys_dict_item"
COMMENT ON COLUMN "sys_dict_item"."code" IS '字典项编码，程序实际使用的值。';
-- Set comment to column: "parent_code" on table: "sys_dict_item"
COMMENT ON COLUMN "sys_dict_item"."parent_code" IS '父级字典项编码，0 表示顶级节点。';
-- Set comment to column: "label" on table: "sys_dict_item"
COMMENT ON COLUMN "sys_dict_item"."label" IS '字典项展示文本。';
-- Set comment to column: "is_default_value" on table: "sys_dict_item"
COMMENT ON COLUMN "sys_dict_item"."is_default_value" IS '是否为默认值。';
-- Set comment to column: "is_enable" on table: "sys_dict_item"
COMMENT ON COLUMN "sys_dict_item"."is_enable" IS '字典项是否启用。';
-- Set comment to column: "description" on table: "sys_dict_item"
COMMENT ON COLUMN "sys_dict_item"."description" IS '字典项描述。';
-- Set comment to column: "sort_id" on table: "sys_dict_item"
COMMENT ON COLUMN "sys_dict_item"."sort_id" IS '排序值。';
-- Set comment to column: "deleted_at" on table: "sys_dict_item"
COMMENT ON COLUMN "sys_dict_item"."deleted_at" IS '逻辑删除时间，NULL 表示未删除。';
-- Set comment to column: "created_by" on table: "sys_dict_item"
COMMENT ON COLUMN "sys_dict_item"."created_by" IS '创建人用户 ID。';
-- Set comment to column: "created_at" on table: "sys_dict_item"
COMMENT ON COLUMN "sys_dict_item"."created_at" IS '创建时间。';
-- Set comment to column: "updated_by" on table: "sys_dict_item"
COMMENT ON COLUMN "sys_dict_item"."updated_by" IS '最后更新人用户 ID。';
-- Set comment to column: "updated_at" on table: "sys_dict_item"
COMMENT ON COLUMN "sys_dict_item"."updated_at" IS '最后更新时间。';
-- Create "sys_menu" table
CREATE TABLE "sys_menu" (
  "id" bigint NOT NULL,
  "parent_id" bigint NOT NULL DEFAULT 0,
  "app_id" character varying(64) NULL,
  "title" character varying(128) NOT NULL,
  "menu_type" smallint NOT NULL,
  "route_path" text NULL,
  "component_path" text NULL,
  "permission_code" character varying(128) NULL,
  "icon" character varying(128) NULL,
  "is_visible" boolean NOT NULL DEFAULT true,
  "is_cached" boolean NOT NULL DEFAULT true,
  "sort_id" integer NOT NULL DEFAULT 0,
  "deleted_at" timestamp NULL,
  "created_by" bigint NULL,
  "created_at" timestamp NOT NULL,
  "updated_by" bigint NULL,
  "updated_at" timestamp NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uk_sys_menu_permission_code" UNIQUE ("permission_code")
);
-- Create index "idx_sys_menu_deleted_at" to table: "sys_menu"
CREATE INDEX "idx_sys_menu_deleted_at" ON "sys_menu" ("deleted_at");
-- Create index "idx_sys_menu_parent_id" to table: "sys_menu"
CREATE INDEX "idx_sys_menu_parent_id" ON "sys_menu" ("parent_id");
-- Set comment to table: "sys_menu"
COMMENT ON TABLE "sys_menu" IS '系统菜单与按钮权限表，统一承载菜单、目录、按钮三类资源。';
-- Set comment to column: "id" on table: "sys_menu"
COMMENT ON COLUMN "sys_menu"."id" IS '雪花算法生成的主键 ID。';
-- Set comment to column: "parent_id" on table: "sys_menu"
COMMENT ON COLUMN "sys_menu"."parent_id" IS '父级菜单 ID，0 表示顶级目录。';
-- Set comment to column: "app_id" on table: "sys_menu"
COMMENT ON COLUMN "sys_menu"."app_id" IS '所属应用编码，用于多应用场景区分菜单树。';
-- Set comment to column: "title" on table: "sys_menu"
COMMENT ON COLUMN "sys_menu"."title" IS '菜单或按钮展示标题。';
-- Set comment to column: "menu_type" on table: "sys_menu"
COMMENT ON COLUMN "sys_menu"."menu_type" IS '资源类型：0-菜单，1-按钮，2-目录。';
-- Set comment to column: "route_path" on table: "sys_menu"
COMMENT ON COLUMN "sys_menu"."route_path" IS '前端路由路径。';
-- Set comment to column: "component_path" on table: "sys_menu"
COMMENT ON COLUMN "sys_menu"."component_path" IS '前端组件路径。';
-- Set comment to column: "permission_code" on table: "sys_menu"
COMMENT ON COLUMN "sys_menu"."permission_code" IS '权限资源编码，例如 system:user:read。';
-- Set comment to column: "icon" on table: "sys_menu"
COMMENT ON COLUMN "sys_menu"."icon" IS '菜单图标标识。';
-- Set comment to column: "is_visible" on table: "sys_menu"
COMMENT ON COLUMN "sys_menu"."is_visible" IS '是否在前端菜单树中可见。';
-- Set comment to column: "is_cached" on table: "sys_menu"
COMMENT ON COLUMN "sys_menu"."is_cached" IS '前端页面是否缓存。';
-- Set comment to column: "sort_id" on table: "sys_menu"
COMMENT ON COLUMN "sys_menu"."sort_id" IS '排序值，值越小越靠前。';
-- Set comment to column: "deleted_at" on table: "sys_menu"
COMMENT ON COLUMN "sys_menu"."deleted_at" IS '逻辑删除时间，NULL 表示未删除。';
-- Set comment to column: "created_by" on table: "sys_menu"
COMMENT ON COLUMN "sys_menu"."created_by" IS '创建人用户 ID。';
-- Set comment to column: "created_at" on table: "sys_menu"
COMMENT ON COLUMN "sys_menu"."created_at" IS '创建时间。';
-- Set comment to column: "updated_by" on table: "sys_menu"
COMMENT ON COLUMN "sys_menu"."updated_by" IS '最后更新人用户 ID。';
-- Set comment to column: "updated_at" on table: "sys_menu"
COMMENT ON COLUMN "sys_menu"."updated_at" IS '最后更新时间。';
-- Create "sys_role" table
CREATE TABLE "sys_role" (
  "id" bigint NOT NULL,
  "code" character varying(64) NOT NULL,
  "name" character varying(128) NOT NULL,
  "remark" text NULL,
  "is_builtin" boolean NOT NULL DEFAULT false,
  "is_enable" boolean NOT NULL DEFAULT true,
  "sort_id" integer NOT NULL DEFAULT 0,
  "deleted_at" timestamp NULL,
  "created_by" bigint NULL,
  "created_at" timestamp NOT NULL,
  "updated_by" bigint NULL,
  "updated_at" timestamp NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uk_sys_role_code" UNIQUE ("code")
);
-- Create index "idx_sys_role_deleted_at" to table: "sys_role"
CREATE INDEX "idx_sys_role_deleted_at" ON "sys_role" ("deleted_at");
-- Create index "idx_sys_role_is_enable" to table: "sys_role"
CREATE INDEX "idx_sys_role_is_enable" ON "sys_role" ("is_enable");
-- Set comment to table: "sys_role"
COMMENT ON TABLE "sys_role" IS '系统角色表，用于承载后台角色定义。';
-- Set comment to column: "id" on table: "sys_role"
COMMENT ON COLUMN "sys_role"."id" IS '雪花算法生成的主键 ID。';
-- Set comment to column: "code" on table: "sys_role"
COMMENT ON COLUMN "sys_role"."code" IS '角色编码，程序内稳定引用，例如 admin、viewer。';
-- Set comment to column: "name" on table: "sys_role"
COMMENT ON COLUMN "sys_role"."name" IS '角色名称，用于管理界面展示。';
-- Set comment to column: "remark" on table: "sys_role"
COMMENT ON COLUMN "sys_role"."remark" IS '角色备注说明。';
-- Set comment to column: "is_builtin" on table: "sys_role"
COMMENT ON COLUMN "sys_role"."is_builtin" IS '是否为系统内置角色，内置角色通常不允许删除。';
-- Set comment to column: "is_enable" on table: "sys_role"
COMMENT ON COLUMN "sys_role"."is_enable" IS '角色是否启用，禁用后不参与授权。';
-- Set comment to column: "sort_id" on table: "sys_role"
COMMENT ON COLUMN "sys_role"."sort_id" IS '角色排序值，值越小越靠前。';
-- Set comment to column: "deleted_at" on table: "sys_role"
COMMENT ON COLUMN "sys_role"."deleted_at" IS '逻辑删除时间，NULL 表示未删除。';
-- Set comment to column: "created_by" on table: "sys_role"
COMMENT ON COLUMN "sys_role"."created_by" IS '创建人用户 ID。';
-- Set comment to column: "created_at" on table: "sys_role"
COMMENT ON COLUMN "sys_role"."created_at" IS '创建时间。';
-- Set comment to column: "updated_by" on table: "sys_role"
COMMENT ON COLUMN "sys_role"."updated_by" IS '最后更新人用户 ID。';
-- Set comment to column: "updated_at" on table: "sys_role"
COMMENT ON COLUMN "sys_role"."updated_at" IS '最后更新时间。';
-- Create "sys_role_menu" table
CREATE TABLE "sys_role_menu" (
  "id" bigint NOT NULL,
  "role_id" bigint NOT NULL,
  "menu_id" bigint NOT NULL,
  "deleted_at" timestamp NULL,
  "created_by" bigint NULL,
  "created_at" timestamp NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uk_sys_role_menu_role_id_menu_id" UNIQUE ("role_id", "menu_id"),
  CONSTRAINT "fk_sys_role_menu_menu_id" FOREIGN KEY ("menu_id") REFERENCES "sys_menu" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_sys_role_menu_role_id" FOREIGN KEY ("role_id") REFERENCES "sys_role" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_sys_role_menu_menu_id" to table: "sys_role_menu"
CREATE INDEX "idx_sys_role_menu_menu_id" ON "sys_role_menu" ("menu_id");
-- Create index "idx_sys_role_menu_role_id" to table: "sys_role_menu"
CREATE INDEX "idx_sys_role_menu_role_id" ON "sys_role_menu" ("role_id");
-- Set comment to table: "sys_role_menu"
COMMENT ON TABLE "sys_role_menu" IS '系统角色与菜单/按钮资源关系表，用于角色授权。';
-- Set comment to column: "id" on table: "sys_role_menu"
COMMENT ON COLUMN "sys_role_menu"."id" IS '雪花算法生成的主键 ID。';
-- Set comment to column: "role_id" on table: "sys_role_menu"
COMMENT ON COLUMN "sys_role_menu"."role_id" IS '系统角色 ID，关联 sys_role.id。';
-- Set comment to column: "menu_id" on table: "sys_role_menu"
COMMENT ON COLUMN "sys_role_menu"."menu_id" IS '系统菜单资源 ID，关联 sys_menu.id。';
-- Set comment to column: "deleted_at" on table: "sys_role_menu"
COMMENT ON COLUMN "sys_role_menu"."deleted_at" IS '逻辑删除时间，NULL 表示未删除。';
-- Set comment to column: "created_by" on table: "sys_role_menu"
COMMENT ON COLUMN "sys_role_menu"."created_by" IS '创建人用户 ID。';
-- Set comment to column: "created_at" on table: "sys_role_menu"
COMMENT ON COLUMN "sys_role_menu"."created_at" IS '创建时间。';
-- Create "sys_user" table
CREATE TABLE "sys_user" (
  "id" bigint NOT NULL,
  "username" character varying(64) NOT NULL,
  "password_hash" text NOT NULL,
  "nickname" character varying(128) NULL,
  "phone" character varying(32) NULL,
  "email" character varying(255) NULL,
  "avatar_url" text NULL,
  "is_super_admin" boolean NOT NULL DEFAULT false,
  "is_enable" boolean NOT NULL DEFAULT true,
  "sort_id" integer NOT NULL DEFAULT 0,
  "deleted_at" timestamp NULL,
  "created_by" bigint NULL,
  "created_at" timestamp NOT NULL,
  "updated_by" bigint NULL,
  "updated_at" timestamp NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uk_sys_user_email" UNIQUE ("email"),
  CONSTRAINT "uk_sys_user_phone" UNIQUE ("phone"),
  CONSTRAINT "uk_sys_user_username" UNIQUE ("username")
);
-- Create index "idx_sys_user_deleted_at" to table: "sys_user"
CREATE INDEX "idx_sys_user_deleted_at" ON "sys_user" ("deleted_at");
-- Create index "idx_sys_user_is_enable" to table: "sys_user"
CREATE INDEX "idx_sys_user_is_enable" ON "sys_user" ("is_enable");
-- Set comment to table: "sys_user"
COMMENT ON TABLE "sys_user" IS '系统后台用户表，用于后台登录、审计和权限归属。';
-- Set comment to column: "id" on table: "sys_user"
COMMENT ON COLUMN "sys_user"."id" IS '雪花算法生成的主键 ID。';
-- Set comment to column: "username" on table: "sys_user"
COMMENT ON COLUMN "sys_user"."username" IS '登录用户名，全局唯一。';
-- Set comment to column: "password_hash" on table: "sys_user"
COMMENT ON COLUMN "sys_user"."password_hash" IS '经过安全哈希后的密码密文。';
-- Set comment to column: "nickname" on table: "sys_user"
COMMENT ON COLUMN "sys_user"."nickname" IS '用户昵称或显示名。';
-- Set comment to column: "phone" on table: "sys_user"
COMMENT ON COLUMN "sys_user"."phone" IS '手机号，可用于登录或通知。';
-- Set comment to column: "email" on table: "sys_user"
COMMENT ON COLUMN "sys_user"."email" IS '邮箱地址，可用于找回密码和通知。';
-- Set comment to column: "avatar_url" on table: "sys_user"
COMMENT ON COLUMN "sys_user"."avatar_url" IS '头像资源地址。';
-- Set comment to column: "is_super_admin" on table: "sys_user"
COMMENT ON COLUMN "sys_user"."is_super_admin" IS '是否为超级管理员，超级管理员通常拥有全量权限。';
-- Set comment to column: "is_enable" on table: "sys_user"
COMMENT ON COLUMN "sys_user"."is_enable" IS '账号是否启用。';
-- Set comment to column: "sort_id" on table: "sys_user"
COMMENT ON COLUMN "sys_user"."sort_id" IS '排序值，便于后台列表定制顺序。';
-- Set comment to column: "deleted_at" on table: "sys_user"
COMMENT ON COLUMN "sys_user"."deleted_at" IS '逻辑删除时间，NULL 表示未删除。';
-- Set comment to column: "created_by" on table: "sys_user"
COMMENT ON COLUMN "sys_user"."created_by" IS '创建人用户 ID。';
-- Set comment to column: "created_at" on table: "sys_user"
COMMENT ON COLUMN "sys_user"."created_at" IS '创建时间。';
-- Set comment to column: "updated_by" on table: "sys_user"
COMMENT ON COLUMN "sys_user"."updated_by" IS '最后更新人用户 ID。';
-- Set comment to column: "updated_at" on table: "sys_user"
COMMENT ON COLUMN "sys_user"."updated_at" IS '最后更新时间。';
-- Create "sys_user_role" table
CREATE TABLE "sys_user_role" (
  "id" bigint NOT NULL,
  "user_id" bigint NOT NULL,
  "role_id" bigint NOT NULL,
  "deleted_at" timestamp NULL,
  "created_by" bigint NULL,
  "created_at" timestamp NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uk_sys_user_role_user_id_role_id" UNIQUE ("user_id", "role_id"),
  CONSTRAINT "fk_sys_user_role_role_id" FOREIGN KEY ("role_id") REFERENCES "sys_role" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_sys_user_role_user_id" FOREIGN KEY ("user_id") REFERENCES "sys_user" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_sys_user_role_role_id" to table: "sys_user_role"
CREATE INDEX "idx_sys_user_role_role_id" ON "sys_user_role" ("role_id");
-- Create index "idx_sys_user_role_user_id" to table: "sys_user_role"
CREATE INDEX "idx_sys_user_role_user_id" ON "sys_user_role" ("user_id");
-- Set comment to table: "sys_user_role"
COMMENT ON TABLE "sys_user_role" IS '系统用户与角色关系表，用于描述一个用户拥有多个角色。';
-- Set comment to column: "id" on table: "sys_user_role"
COMMENT ON COLUMN "sys_user_role"."id" IS '雪花算法生成的主键 ID。';
-- Set comment to column: "user_id" on table: "sys_user_role"
COMMENT ON COLUMN "sys_user_role"."user_id" IS '系统用户 ID，关联 sys_user.id。';
-- Set comment to column: "role_id" on table: "sys_user_role"
COMMENT ON COLUMN "sys_user_role"."role_id" IS '系统角色 ID，关联 sys_role.id。';
-- Set comment to column: "deleted_at" on table: "sys_user_role"
COMMENT ON COLUMN "sys_user_role"."deleted_at" IS '逻辑删除时间，NULL 表示未删除。';
-- Set comment to column: "created_by" on table: "sys_user_role"
COMMENT ON COLUMN "sys_user_role"."created_by" IS '创建人用户 ID。';
-- Set comment to column: "created_at" on table: "sys_user_role"
COMMENT ON COLUMN "sys_user_role"."created_at" IS '创建时间。';
