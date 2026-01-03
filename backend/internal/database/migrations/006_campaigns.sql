CREATE TABLE IF NOT EXISTS campaigns (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    cabinet_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    trigger_type VARCHAR(50) NOT NULL DEFAULT 'on_pending',
    is_active BOOLEAN DEFAULT false,
    quiet_hours_enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS campaign_steps (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    step_order INTEGER NOT NULL,
    delay_hours INTEGER DEFAULT 0,
    channel VARCHAR(50) NOT NULL,
    template_id VARCHAR(255),
    config JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS campaign_executions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    pending_line_id UUID NOT NULL, 
    current_step_order INTEGER DEFAULT 0,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    stop_reason VARCHAR(50),
    last_step_executed_at TIMESTAMP WITH TIME ZONE,
    next_step_scheduled_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_campaigns_cabinet_id ON campaigns(cabinet_id);
CREATE INDEX IF NOT EXISTS idx_campaign_executions_pending_line_id ON campaign_executions(pending_line_id);
CREATE INDEX IF NOT EXISTS idx_campaign_executions_campaign_id ON campaign_executions(campaign_id);
CREATE INDEX IF NOT EXISTS idx_campaign_executions_status ON campaign_executions(status);
