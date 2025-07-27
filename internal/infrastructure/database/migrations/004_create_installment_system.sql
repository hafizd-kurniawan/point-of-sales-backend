-- Add installment-related fields to sales_transactions table
ALTER TABLE sales_transactions 
ADD COLUMN down_payment DECIMAL(15,2) DEFAULT 0,
ADD COLUMN remaining_amount DECIMAL(15,2) DEFAULT 0,
ADD COLUMN payment_status VARCHAR(20) DEFAULT 'completed' CHECK (payment_status IN ('completed', 'pending', 'overdue')),
ADD COLUMN installment_plan VARCHAR(20),
ADD COLUMN installment_months INTEGER DEFAULT 0,
ADD COLUMN monthly_payment DECIMAL(15,2) DEFAULT 0,
ADD COLUMN interest_rate DECIMAL(5,2) DEFAULT 0,
ADD COLUMN bank_name VARCHAR(100),
ADD COLUMN loan_reference VARCHAR(100),
ADD COLUMN down_payment_date TIMESTAMP WITH TIME ZONE;

-- Update payment_method constraint to include 'mixed'
ALTER TABLE sales_transactions 
DROP CONSTRAINT sales_transactions_payment_method_check;

ALTER TABLE sales_transactions 
ADD CONSTRAINT sales_transactions_payment_method_check 
CHECK (payment_method IN ('cash', 'transfer', 'check', 'credit', 'mixed'));

-- Create payment_installments table
CREATE TABLE IF NOT EXISTS payment_installments (
    id SERIAL PRIMARY KEY,
    sales_transaction_id INTEGER NOT NULL REFERENCES sales_transactions(id) ON DELETE CASCADE,
    installment_number INTEGER NOT NULL,
    due_date DATE NOT NULL,
    amount DECIMAL(15,2) NOT NULL,
    paid_amount DECIMAL(15,2) DEFAULT 0,
    paid_date TIMESTAMP WITH TIME ZONE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'paid', 'overdue', 'partial')),
    late_fee DECIMAL(15,2) DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT valid_installment_amounts CHECK (amount > 0 AND paid_amount >= 0 AND paid_amount <= amount),
    CONSTRAINT unique_installment_per_transaction UNIQUE (sales_transaction_id, installment_number)
);

-- Create payment_methods configuration table
CREATE TABLE IF NOT EXISTS payment_methods (
    id SERIAL PRIMARY KEY,
    method_name VARCHAR(20) UNIQUE NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    requires_down_payment BOOLEAN DEFAULT FALSE,
    min_down_payment_percentage DECIMAL(5,2) DEFAULT 0,
    max_installment_months INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Insert default payment method configurations
INSERT INTO payment_methods (method_name, display_name, requires_down_payment, min_down_payment_percentage, max_installment_months, is_active) VALUES
('cash', 'Cash Payment', FALSE, 0, 0, TRUE),
('transfer', 'Bank Transfer', FALSE, 0, 0, TRUE),
('check', 'Check Payment', FALSE, 0, 0, TRUE),
('credit', 'Credit/Financing', TRUE, 20.00, 60, TRUE),
('mixed', 'Mixed Payment', TRUE, 10.00, 48, TRUE);

-- Indexes for performance
CREATE INDEX idx_payment_installments_transaction ON payment_installments(sales_transaction_id);
CREATE INDEX idx_payment_installments_due_date ON payment_installments(due_date);
CREATE INDEX idx_payment_installments_status ON payment_installments(status);
CREATE INDEX idx_sales_transactions_payment_status ON sales_transactions(payment_status);
CREATE INDEX idx_sales_transactions_down_payment_date ON sales_transactions(down_payment_date);

-- Function to update payment status based on installments
CREATE OR REPLACE FUNCTION update_payment_status()
RETURNS TRIGGER AS $$
DECLARE
    total_installments INTEGER;
    paid_installments INTEGER;
    overdue_installments INTEGER;
    transaction_status VARCHAR(20);
BEGIN
    -- Count total installments for this transaction
    SELECT COUNT(*) INTO total_installments
    FROM payment_installments 
    WHERE sales_transaction_id = COALESCE(NEW.sales_transaction_id, OLD.sales_transaction_id);
    
    -- Count paid installments
    SELECT COUNT(*) INTO paid_installments
    FROM payment_installments 
    WHERE sales_transaction_id = COALESCE(NEW.sales_transaction_id, OLD.sales_transaction_id)
    AND status = 'paid';
    
    -- Count overdue installments  
    SELECT COUNT(*) INTO overdue_installments
    FROM payment_installments 
    WHERE sales_transaction_id = COALESCE(NEW.sales_transaction_id, OLD.sales_transaction_id)
    AND status = 'overdue';
    
    -- Determine transaction payment status
    IF total_installments = 0 THEN
        transaction_status := 'completed';
    ELSIF paid_installments = total_installments THEN
        transaction_status := 'completed';
    ELSIF overdue_installments > 0 THEN
        transaction_status := 'overdue';
    ELSE
        transaction_status := 'pending';
    END IF;
    
    -- Update sales transaction payment status
    UPDATE sales_transactions 
    SET payment_status = transaction_status
    WHERE id = COALESCE(NEW.sales_transaction_id, OLD.sales_transaction_id);
    
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- Trigger to update payment status when installments change
CREATE TRIGGER trigger_update_payment_status
    AFTER INSERT OR UPDATE OR DELETE ON payment_installments
    FOR EACH ROW
    EXECUTE FUNCTION update_payment_status();

-- Function to mark overdue installments
CREATE OR REPLACE FUNCTION mark_overdue_installments()
RETURNS void AS $$
BEGIN
    UPDATE payment_installments 
    SET status = 'overdue'
    WHERE due_date < CURRENT_DATE 
    AND status = 'pending';
END;
$$ LANGUAGE plpgsql;