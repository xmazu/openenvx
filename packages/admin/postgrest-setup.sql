-- PostgREST Database Setup Script
-- Run this in your PostgreSQL database (e.g., Neon) to configure roles for PostgREST

-- ============================================
-- STEP 1: Create authenticator role
-- This is the role PostgREST uses to connect to the database
-- It then switches to anon or admin_service based on JWT token
-- ============================================
CREATE ROLE authenticator WITH LOGIN PASSWORD 'change-this-password';

-- ============================================
-- STEP 2: Create roles for JWT authentication
-- ============================================

-- Role for unauthenticated requests
CREATE ROLE anon NOLOGIN;

-- Role for admin panel (authenticated requests)
CREATE ROLE admin_service WITH LOGIN;

-- ============================================
-- STEP 3: Grant role switching permissions
-- Authenticator must be able to switch to these roles
-- ============================================
GRANT anon TO authenticator;
GRANT admin_service TO authenticator;

-- ============================================
-- STEP 4: Grant schema usage
-- ============================================
GRANT USAGE ON SCHEMA public TO anon;
GRANT USAGE ON SCHEMA public TO admin_service;

-- ============================================
-- STEP 5: Grant table permissions
-- Adjust these permissions based on your security requirements
-- ============================================

-- For anon role (unauthenticated users)
GRANT SELECT ON ALL TABLES IN SCHEMA public TO anon;

-- For admin_service role (full access for admin panel)
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO admin_service;
GRANT ALL ON ALL SEQUENCES IN SCHEMA public TO admin_service;

-- ============================================
-- STEP 6: Default privileges for future tables
-- ============================================
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO anon;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO admin_service;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO admin_service;

-- ============================================
-- STEP 7: Sequence permissions
-- ============================================
GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO anon;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT USAGE ON SEQUENCES TO anon;

-- ============================================
-- USAGE INSTRUCTIONS
-- ============================================
--
-- 1. Update the authenticator password above
-- 2. Run this script in your database (e.g., Neon SQL Editor)
-- 3. Update your PostgREST DATABASE_URL to use authenticator role:
--    postgresql://authenticator:your-password@host/database
--
-- ============================================
-- PRODUCTION SECURITY NOTES
-- ============================================
--
-- 1. Change the default password for authenticator role
-- 2. Consider limiting anon permissions to only SELECT
-- 3. Enable Row Level Security (RLS) for data isolation:
--    ALTER TABLE your_table ENABLE ROW LEVEL SECURITY;
-- 4. Create RLS policies for fine-grained access control
-- 5. Use SSL mode require for production connections
