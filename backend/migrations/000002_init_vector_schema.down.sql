-- 回滚：删除向量库表
DROP INDEX IF EXISTS idx_document_embeddings_embedding;
DROP INDEX IF EXISTS idx_document_embeddings_document_id;
DROP TABLE IF EXISTS document_embeddings;
