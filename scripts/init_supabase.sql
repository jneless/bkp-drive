-- Supabase PostgreSQL 初始化脚本
-- bkp-drive 用户表结构

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(12) UNIQUE NOT NULL,  -- 用户唯一标识符，bkp-开头
    username VARCHAR(30) UNIQUE NOT NULL,  -- 用户名
    password VARCHAR(255) NOT NULL,  -- 密码hash (bcrypt)
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 创建索引以提升查询性能
CREATE INDEX IF NOT EXISTS idx_users_user_id ON users(user_id);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);

-- 添加表和字段注释
COMMENT ON TABLE users IS '用户表';
COMMENT ON COLUMN users.id IS '自增主键';
COMMENT ON COLUMN users.user_id IS '用户唯一标识符，bkp-开头';
COMMENT ON COLUMN users.username IS '用户名';
COMMENT ON COLUMN users.password IS 'bcrypt加密的密码hash';
COMMENT ON COLUMN users.created_at IS '创建时间';
COMMENT ON COLUMN users.updated_at IS '更新时间';

-- 创建更新时间自动更新函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 创建触发器：每次UPDATE时自动更新 updated_at
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 可选：插入测试数据（密码为 "test123" 的bcrypt hash）
-- INSERT INTO users (user_id, username, password) VALUES
-- ('bkp-testuser', 'testuser', '$2a$10$example_hash_here');

-- 提示：在Supabase SQL Editor中执行此脚本
-- 访问：https://app.supabase.com/project/_/sql
