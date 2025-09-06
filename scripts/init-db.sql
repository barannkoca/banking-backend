-- Database initialization script for Banking Backend
-- Bu script PostgreSQL container'ı başlatıldığında çalışır

-- Database oluştur (eğer yoksa)
-- CREATE DATABASE IF NOT EXISTS banking_db;

-- Banking Backend için gerekli extension'ları yükle
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Timezone ayarla
SET timezone = 'UTC';

-- Banking Backend kullanıcısına gerekli yetkileri ver
GRANT ALL PRIVILEGES ON DATABASE banking_db TO barankoca;
GRANT ALL PRIVILEGES ON SCHEMA public TO barankoca;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO barankoca;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO barankoca;

-- Default privileges ayarla
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO barankoca;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO barankoca;
