-- Create outbox_events table for transactional event publishing (Outbox pattern)
-- 트랜잭션 이벤트 발행을 위한 outbox_events 테이블 생성 (Outbox 패턴)

CREATE TABLE IF NOT EXISTS outbox_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    topic VARCHAR(255) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    data JSONB NOT NULL,
    workspace_id UUID,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    retry_count INTEGER DEFAULT 0,
    last_error TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    published_at TIMESTAMP
);

-- Create indexes for better performance
-- 성능 향상을 위한 인덱스 생성
CREATE INDEX idx_outbox_events_status ON outbox_events(status);
CREATE INDEX idx_outbox_events_created_at ON outbox_events(created_at);
CREATE INDEX idx_outbox_events_topic ON outbox_events(topic);
CREATE INDEX idx_outbox_events_event_type ON outbox_events(event_type);
CREATE INDEX idx_outbox_events_workspace_id ON outbox_events(workspace_id);
CREATE INDEX idx_outbox_events_published_at ON outbox_events(published_at);

-- Add comment to table
-- 테이블에 주석 추가
COMMENT ON TABLE outbox_events IS 'Stores events to be published to NATS after transaction commit (Outbox pattern)';
COMMENT ON COLUMN outbox_events.status IS 'Event status: pending, processing, published, failed';
COMMENT ON COLUMN outbox_events.retry_count IS 'Number of times the event publishing was retried';

