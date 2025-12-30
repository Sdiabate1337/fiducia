-- Fiducia Initial Schema
-- Multi-tenant architecture for accounting firms

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================
-- CORE MULTI-TENANT TABLES
-- ============================================

-- Cabinets (accounting firms)
CREATE TABLE cabinets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    siret VARCHAR(14),
    email VARCHAR(255),
    phone VARCHAR(20),
    address TEXT,
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Collaborators (cabinet employees)
CREATE TABLE collaborators (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    cabinet_id UUID NOT NULL REFERENCES cabinets(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    role VARCHAR(50) DEFAULT 'collaborator', -- admin, manager, collaborator
    voice_id VARCHAR(100), -- ElevenLabs voice ID
    voice_sample_url TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(cabinet_id, email)
);

-- Clients (cabinet's clients)
CREATE TABLE clients (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    cabinet_id UUID NOT NULL REFERENCES cabinets(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    siren VARCHAR(9),
    siret VARCHAR(14),
    phone VARCHAR(20), -- Format E.164 for WhatsApp
    email VARCHAR(255),
    contact_name VARCHAR(255),
    address TEXT,
    notes TEXT,
    whatsapp_opted_in BOOLEAN DEFAULT false,
    whatsapp_opted_in_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================
-- MODULE A: PENDING LINES (COMPTE 471)
-- ============================================

CREATE TYPE pending_line_status AS ENUM (
    'pending',      -- Awaiting contact
    'contacted',    -- Message sent, awaiting response
    'received',     -- Document received, awaiting validation
    'validated',    -- Approved and matched
    'rejected',     -- Rejected by collaborator
    'expired'       -- No response after deadline
);

CREATE TABLE pending_lines (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    cabinet_id UUID NOT NULL REFERENCES cabinets(id) ON DELETE CASCADE,
    client_id UUID REFERENCES clients(id) ON DELETE SET NULL,
    
    -- Transaction details
    amount DECIMAL(15,2) NOT NULL,
    transaction_date DATE NOT NULL,
    bank_label VARCHAR(500),
    account_number VARCHAR(50),
    
    -- Source tracking
    import_batch_id UUID,
    source_file VARCHAR(255),
    source_row_number INTEGER,
    
    -- Status
    status pending_line_status DEFAULT 'pending',
    last_contacted_at TIMESTAMPTZ,
    contact_count INTEGER DEFAULT 0,
    
    -- Assignment
    assigned_to UUID REFERENCES collaborators(id),
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Import batches tracking
CREATE TABLE import_batches (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    cabinet_id UUID NOT NULL REFERENCES cabinets(id) ON DELETE CASCADE,
    imported_by UUID REFERENCES collaborators(id),
    filename VARCHAR(255),
    file_type VARCHAR(50), -- csv, xlsx, etc.
    total_rows INTEGER,
    imported_rows INTEGER,
    failed_rows INTEGER,
    errors JSONB,
    status VARCHAR(50) DEFAULT 'processing', -- processing, completed, failed
    created_at TIMESTAMPTZ DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

-- ============================================
-- MODULE C: WHATSAPP MESSAGES
-- ============================================

CREATE TYPE message_direction AS ENUM ('outbound', 'inbound');
CREATE TYPE message_type AS ENUM ('text', 'voice', 'interactive', 'media', 'template');
CREATE TYPE message_status AS ENUM (
    'queued',
    'sending',
    'sent',
    'delivered',
    'read',
    'failed',
    'received'
);

CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    pending_line_id UUID REFERENCES pending_lines(id) ON DELETE CASCADE,
    client_id UUID REFERENCES clients(id) ON DELETE SET NULL,
    
    direction message_direction NOT NULL,
    message_type message_type NOT NULL,
    
    -- Content
    content TEXT,
    media_url TEXT,
    template_name VARCHAR(100),
    template_params JSONB,
    
    -- WhatsApp tracking
    wa_message_id VARCHAR(100),
    wa_conversation_id VARCHAR(100),
    
    status message_status DEFAULT 'queued',
    error_message TEXT,
    
    -- Timing
    scheduled_at TIMESTAMPTZ,
    sent_at TIMESTAMPTZ,
    delivered_at TIMESTAMPTZ,
    read_at TIMESTAMPTZ,
    
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================
-- MODULE D: RECEIVED DOCUMENTS
-- ============================================

CREATE TYPE document_type AS ENUM (
    'receipt',
    'invoice',
    'bank_statement',
    'contract',
    'other'
);

CREATE TABLE received_documents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    pending_line_id UUID REFERENCES pending_lines(id) ON DELETE CASCADE,
    message_id UUID REFERENCES messages(id) ON DELETE SET NULL,
    client_id UUID REFERENCES clients(id) ON DELETE SET NULL,
    
    -- File info
    file_url TEXT NOT NULL,
    file_type VARCHAR(50),
    file_size INTEGER,
    original_filename VARCHAR(255),
    
    -- OCR results
    document_type document_type,
    ocr_result JSONB,
    ocr_confidence DECIMAL(3,2),
    extracted_amount DECIMAL(15,2),
    extracted_date DATE,
    extracted_merchant VARCHAR(255),
    
    -- Processing status
    ocr_status VARCHAR(50) DEFAULT 'pending', -- pending, processing, completed, failed
    ocr_error TEXT,
    
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================
-- MODULE E: MATCHING PROPOSALS
-- ============================================

CREATE TYPE proposal_status AS ENUM (
    'pending',
    'approved',
    'rejected'
);

CREATE TABLE matching_proposals (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    pending_line_id UUID NOT NULL REFERENCES pending_lines(id) ON DELETE CASCADE,
    document_id UUID NOT NULL REFERENCES received_documents(id) ON DELETE CASCADE,
    
    -- Matching info
    proposed_by VARCHAR(20) NOT NULL, -- auto, manual
    match_confidence DECIMAL(3,2),
    match_reasons JSONB, -- {"amount_match": true, "date_match": true, ...}
    
    -- Validation
    status proposal_status DEFAULT 'pending',
    validated_by UUID REFERENCES collaborators(id),
    validated_at TIMESTAMPTZ,
    rejection_reason TEXT,
    
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================
-- EXPORTS (Writeback to ERP)
-- ============================================

CREATE TABLE exports (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    cabinet_id UUID NOT NULL REFERENCES cabinets(id) ON DELETE CASCADE,
    exported_by UUID REFERENCES collaborators(id),
    
    format VARCHAR(50) NOT NULL, -- tra, xml, csv
    file_url TEXT,
    
    -- Included items
    matching_proposal_ids UUID[],
    total_entries INTEGER,
    total_amount DECIMAL(15,2),
    
    status VARCHAR(50) DEFAULT 'generating', -- generating, ready, downloaded
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================
-- INDEXES FOR PERFORMANCE
-- ============================================

-- Cabinet-scoped queries
CREATE INDEX idx_collaborators_cabinet ON collaborators(cabinet_id);
CREATE INDEX idx_clients_cabinet ON clients(cabinet_id);
CREATE INDEX idx_pending_lines_cabinet ON pending_lines(cabinet_id);
CREATE INDEX idx_import_batches_cabinet ON import_batches(cabinet_id);

-- Status filtering
CREATE INDEX idx_pending_lines_status ON pending_lines(status);
CREATE INDEX idx_pending_lines_cabinet_status ON pending_lines(cabinet_id, status);
CREATE INDEX idx_messages_status ON messages(status);
CREATE INDEX idx_matching_proposals_status ON matching_proposals(status);

-- Message tracking
CREATE INDEX idx_messages_pending_line ON messages(pending_line_id);
CREATE INDEX idx_messages_wa_id ON messages(wa_message_id);

-- Document processing
CREATE INDEX idx_received_documents_pending_line ON received_documents(pending_line_id);
CREATE INDEX idx_received_documents_ocr_status ON received_documents(ocr_status);

-- Date-based queries
CREATE INDEX idx_pending_lines_date ON pending_lines(transaction_date);
CREATE INDEX idx_pending_lines_created ON pending_lines(created_at);

-- ============================================
-- UPDATED_AT TRIGGER
-- ============================================

CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER cabinets_updated_at BEFORE UPDATE ON cabinets
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER collaborators_updated_at BEFORE UPDATE ON collaborators
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER clients_updated_at BEFORE UPDATE ON clients
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER pending_lines_updated_at BEFORE UPDATE ON pending_lines
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
