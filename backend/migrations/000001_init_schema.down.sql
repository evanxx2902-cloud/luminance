-- 回滚：删除业务库表
DROP INDEX IF EXISTS idx_documents_status;
DROP INDEX IF EXISTS idx_documents_user_id;
DROP TABLE IF EXISTS documents;
DROP TABLE IF EXISTS users;
