-- Modify "sys_user" table
ALTER TABLE "sys_user" ADD COLUMN "session_version" bigint NOT NULL DEFAULT 1;
COMMENT ON COLUMN "sys_user"."session_version" IS '用户级会话版本，递增后使全部旧 access/refresh token 失效。';
