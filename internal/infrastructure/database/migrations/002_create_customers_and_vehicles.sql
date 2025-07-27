-- Vehicle Categories Table
CREATE TABLE IF NOT EXISTS vehicle_categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE NULL
);

-- Customers Table
CREATE TABLE IF NOT EXISTS customers (
    id SERIAL PRIMARY KEY,
    customer_code VARCHAR(20) UNIQUE NOT NULL,
    full_name VARCHAR(100) NOT NULL,
    email VARCHAR(100),
    phone VARCHAR(20) NOT NULL,
    address TEXT,
    city VARCHAR(50),
    id_number VARCHAR(30),
    customer_type VARCHAR(20) DEFAULT 'individual' CHECK (customer_type IN ('individual', 'company')),
    company_name VARCHAR(100),
    tax_number VARCHAR(30),
    notes TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE NULL
);

-- Vehicles Table
CREATE TABLE IF NOT EXISTS vehicles (
    id SERIAL PRIMARY KEY,
    vehicle_code VARCHAR(20) UNIQUE NOT NULL,
    category_id INTEGER REFERENCES vehicle_categories(id),
    brand VARCHAR(50) NOT NULL,
    model VARCHAR(100) NOT NULL,
    year INTEGER NOT NULL,
    color VARCHAR(30) NOT NULL,
    engine_type VARCHAR(50),
    transmission VARCHAR(20) CHECK (transmission IN ('manual', 'automatic', 'cvt')),
    fuel_type VARCHAR(20) CHECK (fuel_type IN ('gasoline', 'diesel', 'electric', 'hybrid')),
    chassis_number VARCHAR(50) UNIQUE,
    engine_number VARCHAR(50),
    license_plate VARCHAR(15),
    purchase_price DECIMAL(15,2) DEFAULT 0,
    selling_price DECIMAL(15,2) DEFAULT 0,
    condition_status VARCHAR(20) DEFAULT 'good' CHECK (condition_status IN ('excellent', 'good', 'fair', 'poor')),
    availability_status VARCHAR(20) DEFAULT 'available' CHECK (availability_status IN ('available', 'sold', 'reserved', 'maintenance')),
    location VARCHAR(100),
    mileage INTEGER DEFAULT 0,
    description TEXT,
    features TEXT[],
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE NULL
);

-- Vehicle Photos Table
CREATE TABLE IF NOT EXISTS vehicle_photos (
    id SERIAL PRIMARY KEY,
    vehicle_id INTEGER NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    photo_url VARCHAR(500) NOT NULL,
    photo_type VARCHAR(20) DEFAULT 'exterior' CHECK (photo_type IN ('exterior', 'interior', 'engine', 'document')),
    is_primary BOOLEAN DEFAULT false,
    sort_order INTEGER DEFAULT 0,
    uploaded_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE NULL
);

-- Indexes for better performance
CREATE INDEX IF NOT EXISTS idx_customers_customer_code ON customers(customer_code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_customers_full_name ON customers(full_name) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_customers_phone ON customers(phone) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_customers_email ON customers(email) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_customers_type ON customers(customer_type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_customers_deleted_at ON customers(deleted_at);

CREATE INDEX IF NOT EXISTS idx_vehicles_vehicle_code ON vehicles(vehicle_code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_vehicles_brand_model ON vehicles(brand, model) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_vehicles_year ON vehicles(year) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_vehicles_price_range ON vehicles(selling_price) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_vehicles_status ON vehicles(availability_status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_vehicles_category ON vehicles(category_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_vehicles_deleted_at ON vehicles(deleted_at);

CREATE INDEX IF NOT EXISTS idx_vehicle_photos_vehicle_id ON vehicle_photos(vehicle_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_vehicle_photos_is_primary ON vehicle_photos(is_primary) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_vehicle_photos_deleted_at ON vehicle_photos(deleted_at);

CREATE INDEX IF NOT EXISTS idx_vehicle_categories_name ON vehicle_categories(name) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_vehicle_categories_deleted_at ON vehicle_categories(deleted_at);

-- Triggers for auto-update timestamps
CREATE TRIGGER update_customers_updated_at BEFORE UPDATE ON customers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_vehicles_updated_at BEFORE UPDATE ON vehicles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_vehicle_photos_updated_at BEFORE UPDATE ON vehicle_photos
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_vehicle_categories_updated_at BEFORE UPDATE ON vehicle_categories
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Function to generate customer code
CREATE OR REPLACE FUNCTION generate_customer_code() RETURNS TEXT AS $$
DECLARE
    next_id INTEGER;
    code TEXT;
BEGIN
    SELECT COALESCE(MAX(id), 0) + 1 INTO next_id FROM customers;
    code := 'CUST' || LPAD(next_id::TEXT, 6, '0');
    RETURN code;
END;
$$ LANGUAGE plpgsql;

-- Function to generate vehicle code
CREATE OR REPLACE FUNCTION generate_vehicle_code() RETURNS TEXT AS $$
DECLARE
    next_id INTEGER;
    code TEXT;
BEGIN
    SELECT COALESCE(MAX(id), 0) + 1 INTO next_id FROM vehicles;
    code := 'VHC' || LPAD(next_id::TEXT, 6, '0');
    RETURN code;
END;
$$ LANGUAGE plpgsql;

-- Insert default vehicle categories
INSERT INTO vehicle_categories (name, description) VALUES
('Sedan', 'Four-door passenger cars with separate trunk'),
('SUV', 'Sport Utility Vehicles with higher ground clearance'),
('Hatchback', 'Compact cars with rear access door'),
('MPV', 'Multi-Purpose Vehicles for families'),
('Pickup', 'Light trucks with open cargo bed'),
('Coupe', 'Two-door sports cars'),
('Convertible', 'Cars with retractable roof'),
('Wagon', 'Extended sedans with cargo space')
ON CONFLICT (name) DO NOTHING;

-- Insert sample customers
INSERT INTO customers (customer_code, full_name, email, phone, address, city, id_number, customer_type) VALUES
(generate_customer_code(), 'John Doe', 'john.doe@email.com', '+62812345678', 'Jl. Merdeka No. 123', 'Jakarta', '3171234567890123', 'individual'),
(generate_customer_code(), 'Jane Smith', 'jane.smith@email.com', '+62812345679', 'Jl. Sudirman No. 456', 'Jakarta', '3171234567890124', 'individual'),
(generate_customer_code(), 'PT. Auto Indonesia', 'contact@autoindo.com', '+62212345678', 'Jl. Thamrin No. 789', 'Jakarta', '1234567890123', 'company')
ON CONFLICT (customer_code) DO NOTHING;

-- Insert sample vehicles
INSERT INTO vehicles (
    vehicle_code, category_id, brand, model, year, color, engine_type, 
    transmission, fuel_type, chassis_number, engine_number, license_plate,
    purchase_price, selling_price, condition_status, availability_status,
    location, mileage, description
) VALUES
(generate_vehicle_code(), 1, 'Toyota', 'Camry', 2022, 'Silver', '2.5L I4', 'automatic', 'gasoline', 'TC22001234567890', 'TC220012345', 'B1234ABC', 450000000, 520000000, 'excellent', 'available', 'Showroom A-1', 15000, 'Like new condition, full service history'),
(generate_vehicle_code(), 2, 'Honda', 'CR-V', 2021, 'White', '1.5L Turbo', 'cvt', 'gasoline', 'HC21001234567890', 'HC210012345', 'B5678DEF', 380000000, 430000000, 'good', 'available', 'Showroom A-2', 25000, 'Well maintained, single owner'),
(generate_vehicle_code(), 3, 'Suzuki', 'Swift', 2023, 'Red', '1.2L I4', 'manual', 'gasoline', 'SS23001234567890', 'SS230012345', 'B9012GHI', 180000000, 210000000, 'excellent', 'available', 'Showroom B-1', 5000, 'Brand new, never been in accident')
ON CONFLICT (vehicle_code) DO NOTHING;
