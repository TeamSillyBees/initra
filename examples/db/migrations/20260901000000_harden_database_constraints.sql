-- Modify "sys_dict_collection" table
ALTER TABLE "sys_dict_collection" ADD CONSTRAINT "ck_sys_dict_collection_item_length" CHECK ("item_length" IS NULL OR "item_length" > 0);
-- Modify "sys_menu" table
ALTER TABLE "sys_menu" ADD CONSTRAINT "ck_sys_menu_parent_not_self" CHECK ("parent_id" IS NULL OR "parent_id" <> "id"), ADD CONSTRAINT "ck_sys_menu_type" CHECK ("menu_type" IN (0, 1, 2));
-- 删除历史物理外键，关联一致性由应用事务校验与行锁保证
ALTER TABLE "sys_dict_item" DROP CONSTRAINT IF EXISTS "fk_sys_dict_item_collection";
ALTER TABLE "sys_role_menu" DROP CONSTRAINT IF EXISTS "fk_sys_role_menu_menu", DROP CONSTRAINT IF EXISTS "fk_sys_role_menu_role";
ALTER TABLE "sys_user_role" DROP CONSTRAINT IF EXISTS "fk_sys_user_role_role", DROP CONSTRAINT IF EXISTS "fk_sys_user_role_user";
-- 删除已被复合唯一索引最左列覆盖的重复索引
DROP INDEX IF EXISTS "sysrolemenu_role_id";
DROP INDEX IF EXISTS "sysuserrole_user_id";
