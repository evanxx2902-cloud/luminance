-- 扩展 users 表，新增会员和认证相关字段
ALTER TABLE users
  ADD COLUMN IF NOT EXISTS is_member        BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS member_level     SMALLINT NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS member_expire_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS free_trial_count INTEGER NOT NULL DEFAULT 1,
  ADD COLUMN IF NOT EXISTS avatar           TEXT;

-- email 改为可为空（支持手机号/第三方登录等注册方式）
ALTER TABLE users ALTER COLUMN email DROP NOT NULL;

-- 添加 salt 字段（存储 hex 编码的 32 字节随机盐，共 64 字符）
ALTER TABLE users ADD COLUMN IF NOT EXISTS salt VARCHAR(64) NOT NULL DEFAULT '';

-- 登录尝试记录表（用于防暴力破解、安全审计）
CREATE TABLE IF NOT EXISTS login_attempts (
  id         BIGSERIAL PRIMARY KEY,
  username   VARCHAR(64),
  ip_address INET NOT NULL,
  success    BOOLEAN NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_login_attempts_username_ip
  ON login_attempts (username, ip_address, created_at DESC);
