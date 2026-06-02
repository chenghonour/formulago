-- reverse: modify "sys_roles" table
ALTER TABLE `sys_roles` MODIFY COLUMN `default_router` varchar(255) NOT NULL DEFAULT "dashboard" COMMENT "default menu : dashboard | 默认登录页面";
-- reverse: modify "sys_menus" table
ALTER TABLE `sys_menus` MODIFY COLUMN `parent_id` bigint unsigned NULL;
-- reverse: modify "sys_dictionary_details" table
ALTER TABLE `sys_dictionary_details` MODIFY COLUMN `dictionary_id` bigint unsigned NULL;
