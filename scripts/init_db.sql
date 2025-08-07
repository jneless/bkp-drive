CREATE DATABASE IF NOT EXISTS bkp_drive CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE bkp_drive;

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id VARCHAR(12) UNIQUE NOT NULL COMMENT '用户唯一标识符，bkp-开头',
    username VARCHAR(30) UNIQUE NOT NULL COMMENT '用户名',
    password VARCHAR(255) NOT NULL COMMENT '密码hash',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_username (username)
) ENGINE=InnoDB COMMENT='用户表';

-- 插入示例数据（可选，用于测试）
-- INSERT INTO users (user_id, username, password) VALUES 
-- ('bkp-testuser', 'testuser', '$2a$10$example_hash_here');