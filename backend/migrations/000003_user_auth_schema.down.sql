-- 回滚：删除 login_attempts 表及其索引
DROP INDEX IF EXISTS idx_login_attempts_username_ip;
DROP TABLE IF EXISTS login_attempts;

-- 回滚：移除 users 表新增字段
ALTER TABLE users
  DROP COLUMN IF EXISTS is_member,
  DROP COLUMN IF EXISTS member_level,
  DROP COLUMN IF EXISTS member_expire_at,
  DROP COLUMN IF EXISTS free_trial_count,
  DROP COLUMN IF EXISTS avatar,
  DROP COLUMN IF EXISTS salt;

-- email 重新设为 NOT NULL（可选，取决于是否需要恢复原有约束）
-- ALTER TABLE users ALTER COLUMN email SET NOT NULL;
