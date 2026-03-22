-- 向量库表结构 (在 port 5433 的向量库中执行)
-- 启用 pgvector 扩展
CREATE EXTENSION IF NOT EXISTS vector;

-- document_embeddings: 文档向量嵌入表
CREATE TABLE IF NOT EXISTS document_embeddings (
    id              SERIAL PRIMARY KEY,
    document_id     INTEGER NOT NULL,
    chunk_index     INTEGER NOT NULL DEFAULT 0,
    chunk_text      TEXT,
    embedding       vector(1536),  -- 默认 1536 维，适配 OpenAI text-embedding-3-small
    metadata        JSONB,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(document_id, chunk_index)
);

CREATE INDEX idx_document_embeddings_document_id ON document_embeddings(document_id);

-- 向量相似度搜索索引 (使用 IVFFlat，适用于较大的数据集)
CREATE INDEX IF NOT EXISTS idx_document_embeddings_embedding
ON document_embeddings
USING ivfflat (embedding vector_cosine_ops)
WITH (lists = 100);
