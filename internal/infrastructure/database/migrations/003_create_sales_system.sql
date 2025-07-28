-- Drop tables yang tidak sesuai ERD original (jika ada)
DROP TABLE IF EXISTS sales_commissions CASCADE;
DROP TABLE IF EXISTS payment_records CASCADE;

-- SALES_TRANSACTIONS sesuai ERD Original
CREATE TABLE IF NOT EXISTS sales_transactions (
    id SERIAL PRIMARY KEY,
    transaction_number VARCHAR(20) UNIQUE NOT NULL,
    invoice_number VARCHAR(20) UNIQUE NOT NULL,
    vehicle_id INTEGER NOT NULL REFERENCES vehicles(id) ON DELETE RESTRICT,
    customer_id INTEGER NOT NULL REFERENCES customers(id) ON DELETE RESTRICT,
    vehicle_price DECIMAL(15,2) NOT NULL,
    tax_amount DECIMAL(15,2) DEFAULT 0,
    discount_amount DECIMAL(15,2) DEFAULT 0,
    total_amount DECIMAL(15,2) NOT NULL,
    payment_method VARCHAR(20) NOT NULL CHECK (payment_method IN ('cash', 'transfer', 'check', 'credit', 'mixed')),
    payment_reference VARCHAR(100),
    transaction_date TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    cashier_id INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    status VARCHAR(20) NOT NULL DEFAULT 'completed' CHECK (status IN ('completed', 'cancelled')),
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT valid_amounts CHECK (vehicle_price > 0 AND total_amount > 0)
);

-- PURCHASE_TRANSACTIONS sesuai ERD Original (beli dari customer)
CREATE TABLE IF NOT EXISTS purchase_transactions (
    id SERIAL PRIMARY KEY,
    transaction_number VARCHAR(20) UNIQUE NOT NULL,
    invoice_number VARCHAR(20) UNIQUE NOT NULL,
    vehicle_id INTEGER NOT NULL REFERENCES vehicles(id) ON DELETE RESTRICT,
    customer_id INTEGER NOT NULL REFERENCES customers(id) ON DELETE RESTRICT,
    vehicle_price DECIMAL(15,2) NOT NULL,
    tax_amount DECIMAL(15,2) DEFAULT 0,
    total_amount DECIMAL(15,2) NOT NULL,
    payment_method VARCHAR(20) NOT NULL CHECK (payment_method IN ('cash', 'transfer', 'check', 'mixed')),
    payment_reference VARCHAR(100),
    transaction_date TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    cashier_id INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    status VARCHAR(20) NOT NULL DEFAULT 'completed' CHECK (status IN ('completed', 'cancelled')),
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT valid_purchase_amounts CHECK (vehicle_price > 0 AND total_amount > 0)
);

-- Indexes untuk performance
CREATE INDEX idx_sales_transactions_vehicle ON sales_transactions(vehicle_id);
CREATE INDEX idx_sales_transactions_customer ON sales_transactions(customer_id);
CREATE INDEX idx_sales_transactions_cashier ON sales_transactions(cashier_id);
CREATE INDEX idx_sales_transactions_date ON sales_transactions(transaction_date);
CREATE INDEX idx_sales_transactions_number ON sales_transactions(transaction_number);
CREATE INDEX idx_sales_transactions_invoice ON sales_transactions(invoice_number);
CREATE INDEX idx_sales_transactions_status ON sales_transactions(status);

CREATE INDEX idx_purchase_transactions_vehicle ON purchase_transactions(vehicle_id);
CREATE INDEX idx_purchase_transactions_customer ON purchase_transactions(customer_id);
CREATE INDEX idx_purchase_transactions_cashier ON purchase_transactions(cashier_id);
CREATE INDEX idx_purchase_transactions_date ON purchase_transactions(transaction_date);
CREATE INDEX idx_purchase_transactions_number ON purchase_transactions(transaction_number);
CREATE INDEX idx_purchase_transactions_invoice ON purchase_transactions(invoice_number);

-- Function untuk generate sales transaction number
CREATE OR REPLACE FUNCTION generate_sales_transaction_number()
RETURNS TRIGGER AS $$
DECLARE
    new_number VARCHAR(20);
    year_month VARCHAR(6);
    sequence_num INTEGER;
BEGIN
    -- Format: STX-YYYYMM-XXXX (Sales Transaction)
    year_month := TO_CHAR(CURRENT_DATE, 'YYYYMM');
    
    SELECT COALESCE(MAX(CAST(SUBSTRING(transaction_number FROM 12) AS INTEGER)), 0) + 1
    INTO sequence_num
    FROM sales_transactions
    WHERE transaction_number LIKE 'STX-' || year_month || '-%';
    
    new_number := 'STX-' || year_month || '-' || LPAD(sequence_num::TEXT, 4, '0');
    NEW.transaction_number := new_number;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_generate_sales_transaction_number
    BEFORE INSERT ON sales_transactions
    FOR EACH ROW
    WHEN (NEW.transaction_number IS NULL OR NEW.transaction_number = '')
    EXECUTE FUNCTION generate_sales_transaction_number();

-- Function untuk generate sales invoice number
CREATE OR REPLACE FUNCTION generate_sales_invoice_number()
RETURNS TRIGGER AS $$
DECLARE
    new_number VARCHAR(20);
    year_month VARCHAR(6);
    sequence_num INTEGER;
BEGIN
    -- Format: SINV-YYYYMM-XXXX (Sales Invoice)
    year_month := TO_CHAR(CURRENT_DATE, 'YYYYMM');
    
    SELECT COALESCE(MAX(CAST(SUBSTRING(invoice_number FROM 13) AS INTEGER)), 0) + 1
    INTO sequence_num
    FROM sales_transactions
    WHERE invoice_number LIKE 'SINV-' || year_month || '-%';
    
    new_number := 'SINV-' || year_month || '-' || LPAD(sequence_num::TEXT, 4, '0');
    NEW.invoice_number := new_number;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_generate_sales_invoice_number
    BEFORE INSERT ON sales_transactions
    FOR EACH ROW
    WHEN (NEW.invoice_number IS NULL OR NEW.invoice_number = '')
    EXECUTE FUNCTION generate_sales_invoice_number();

-- Function untuk generate purchase transaction number
CREATE OR REPLACE FUNCTION generate_purchase_transaction_number()
RETURNS TRIGGER AS $$
DECLARE
    new_number VARCHAR(20);
    year_month VARCHAR(6);
    sequence_num INTEGER;
BEGIN
    -- Format: PTX-YYYYMM-XXXX (Purchase Transaction)
    year_month := TO_CHAR(CURRENT_DATE, 'YYYYMM');
    
    SELECT COALESCE(MAX(CAST(SUBSTRING(transaction_number FROM 12) AS INTEGER)), 0) + 1
    INTO sequence_num
    FROM purchase_transactions
    WHERE transaction_number LIKE 'PTX-' || year_month || '-%';
    
    new_number := 'PTX-' || year_month || '-' || LPAD(sequence_num::TEXT, 4, '0');
    NEW.transaction_number := new_number;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_generate_purchase_transaction_number
    BEFORE INSERT ON purchase_transactions
    FOR EACH ROW
    WHEN (NEW.transaction_number IS NULL OR NEW.transaction_number = '')
    EXECUTE FUNCTION generate_purchase_transaction_number();

-- Function untuk generate purchase invoice number
CREATE OR REPLACE FUNCTION generate_purchase_invoice_number()
RETURNS TRIGGER AS $$
DECLARE
    new_number VARCHAR(20);
    year_month VARCHAR(6);
    sequence_num INTEGER;
BEGIN
    -- Format: PINV-YYYYMM-XXXX (Purchase Invoice)
    year_month := TO_CHAR(CURRENT_DATE, 'YYYYMM');
    
    SELECT COALESCE(MAX(CAST(SUBSTRING(invoice_number FROM 13) AS INTEGER)), 0) + 1
    INTO sequence_num
    FROM purchase_transactions
    WHERE invoice_number LIKE 'PINV-' || year_month || '-%';
    
    new_number := 'PINV-' || year_month || '-' || LPAD(sequence_num::TEXT, 4, '0');
    NEW.invoice_number := new_number;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_generate_purchase_invoice_number
    BEFORE INSERT ON purchase_transactions
    FOR EACH ROW
    WHEN (NEW.invoice_number IS NULL OR NEW.invoice_number = '')
    EXECUTE FUNCTION generate_purchase_invoice_number();

-- Function untuk update vehicle status pada sales
CREATE OR REPLACE FUNCTION update_vehicle_on_sales()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status = 'completed' THEN
        UPDATE vehicles 
        SET status = 'sold', 
            sold_to_customer_id = NEW.customer_id, 
            sold_by_cashier = NEW.cashier_id, 
            sold_at = NEW.transaction_date,
            final_selling_price = NEW.vehicle_price
        WHERE id = NEW.vehicle_id;
    ELSIF NEW.status = 'cancelled' AND OLD.status = 'completed' THEN
        UPDATE vehicles 
        SET status = 'ready_to_sell', 
            sold_to_customer_id = NULL, 
            sold_by_cashier = NULL, 
            sold_at = NULL,
            final_selling_price = NULL
        WHERE id = NEW.vehicle_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_vehicle_on_sales
    AFTER INSERT OR UPDATE ON sales_transactions
    FOR EACH ROW
    EXECUTE FUNCTION update_vehicle_on_sales();

-- Function untuk update vehicle status pada purchase
CREATE OR REPLACE FUNCTION update_vehicle_on_purchase()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status = 'completed' THEN
        UPDATE vehicles 
        SET status = 'purchased', 
            purchased_from_customer_id = NEW.customer_id, 
            purchased_by_cashier = NEW.cashier_id, 
            purchased_at = NEW.transaction_date,
            purchase_price = NEW.vehicle_price
        WHERE id = NEW.vehicle_id;
    ELSIF NEW.status = 'cancelled' AND OLD.status = 'completed' THEN
        UPDATE vehicles 
        SET status = 'purchased', -- Keep as purchased, just remove transaction link
            purchased_from_customer_id = NULL, 
            purchased_by_cashier = NULL, 
            purchased_at = NULL
        WHERE id = NEW.vehicle_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_vehicle_on_purchase
    AFTER INSERT OR UPDATE ON purchase_transactions
    FOR EACH ROW
    EXECUTE FUNCTION update_vehicle_on_purchase();
