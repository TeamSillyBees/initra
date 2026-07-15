-- Modify "sys_dict_item" table
ALTER TABLE "sys_dict_item" ADD CONSTRAINT "fk_sys_dict_item_collection" FOREIGN KEY ("collection_id") REFERENCES "sys_dict_collection" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT;
-- Modify "sys_role_menu" table
ALTER TABLE "sys_role_menu" ADD CONSTRAINT "fk_sys_role_menu_menu" FOREIGN KEY ("menu_id") REFERENCES "sys_menu" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT, ADD CONSTRAINT "fk_sys_role_menu_role" FOREIGN KEY ("role_id") REFERENCES "sys_role" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT;
-- Modify "sys_user_role" table
ALTER TABLE "sys_user_role" ADD CONSTRAINT "fk_sys_user_role_role" FOREIGN KEY ("role_id") REFERENCES "sys_role" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT, ADD CONSTRAINT "fk_sys_user_role_user" FOREIGN KEY ("user_id") REFERENCES "sys_user" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT;
