-- Sprint 4: Documents table for OCR results
-- Run: psql fiducia < backend/migrations/004_documents.sql

CREATE TABLE IF NOT EXISTS documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pending_line_id UUID REFERENCES pending_lines(id) ON DELETE SET NULL,
    client_id UUID REFERENCES clients(id) ON DELETE SET NULL,
    message_id UUID REFERENCES messages(id) ON DELETE SET NULL,
    
    -- File info
    file_path TEXT NOT NULL,
    file_name TEXT,
    file_type TEXT,  -- 'image/jpeg', 'application/pdf', etc.
    file_size INTEGER,
    twilio_media_url TEXT,  -- Original Twilio URL
    
    -- OCR Results
    ocr_text TEXT,
    ocr_data JSONB,  -- Structured data: {date, amount, vendor, invoice_number, etc.}
    ocr_status TEXT DEFAULT 'pending',  -- pending, processing, completed, failed
    ocr_error TEXT,
    
    -- Matching
    match_confidence DECIMAL(3,2) DEFAULT 0.00,  -- 0.00 to 1.00
    match_status TEXT DEFAULT 'pending',  -- pending, approved, rejected, auto_matched
    matched_by UUID,  -- User who approved/rejected
    matched_at TIMESTAMP,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_documents_pending_line ON documents(pending_line_id);
CREATE INDEX IF NOT EXISTS idx_documents_client ON documents(client_id);
CREATE INDEX IF NOT EXISTS idx_documents_message ON documents(message_id);
CREATE INDEX IF NOT EXISTS idx_documents_match_status ON documents(match_status);
CREATE INDEX IF NOT EXISTS idx_documents_created ON documents(created_at DESC);

-- Update trigger
CREATE OR REPLACE FUNCTION update_documents_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_documents_updated_at ON documents;
CREATE TRIGGER trg_documents_updated_at
    BEFORE UPDATE ON documents
    FOR EACH ROW
    EXECUTE FUNCTION update_documents_updated_at();
