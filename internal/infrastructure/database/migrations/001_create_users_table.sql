-- Create extension for UUID generation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(100) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    role VARCHAR(20) NOT NULL CHECK (role IN ('admin', 'mechanic', 'cashier')),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE NULL
);

-- User sessions table
CREATE TABLE IF NOT EXISTS user_sessions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    session_token VARCHAR(255) UNIQUE NOT NULL,
    login_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    logout_at TIMESTAMP WITH TIME ZONE NULL,
    ip_address INET,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE NULL
);

-- Indexes for better performance
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);

CREATE INDEX IF NOT EXISTS idx_user_sessions_token ON user_sessions(session_token) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_user_sessions_active ON user_sessions(is_active) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_user_sessions_deleted_at ON user_sessions(deleted_at);

-- Trigger to auto-update updated_at column
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_sessions_updated_at BEFORE UPDATE ON user_sessions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert default admin user (password: admin123)
INSERT INTO users (username, email, password_hash, full_name, phone, role) 
VALUES (
    'admin', 
    'admin@showroom.com', 
    '$2a$10$Xe4Wq8LIH.6CK7FdVXNZSuA8C.8vZ1xvJkP9zHxOxQyH5w.Kq2qcq', -- admin123
    'System Administrator', 
    '+62812345678', 
    'admin'
) ON CONFLICT (username) DO NOTHING;

-- Insert test mechanic user (password: mechanic123)
INSERT INTO users (username, email, password_hash, full_name, phone, role) 
VALUES (
    'mechanic1', 
    'mechanic@showroom.com', 
    '$2a$10$rH.TGzVs1gF8K3J.dVkCtO6L8Q9N2xM4o7P6q8R9sA1B2c3D4e5F6', -- mechanic123
    'Budi Santoso', 
    '+62812345679', 
    'mechanic'
) ON CONFLICT (username) DO NOTHING;

-- Insert test cashier user (password: cashier123)
INSERT INTO users (username, email, password_hash, full_name, phone, role) 
VALUES (
    'cashier1', 
    'cashier@showroom.com', 
    '$2a$10$sI.UHaWt2hG9L4K.eWlDuP7M9R0O3yN5p8Q7r9S0tB2c4D5e6F7g8', -- cashier123
    'Siti Nurhaliza', 
    '+62812345680', 
    'cashier'
) ON CONFLICT (username) DO NOTHING;
