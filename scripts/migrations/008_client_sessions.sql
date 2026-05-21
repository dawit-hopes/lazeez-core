CREATE TABLE IF NOT EXISTS client_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_key VARCHAR(255) NOT NULL,
    table_reference VARCHAR(255) NOT NULL,
    table_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT fk_client_sessions_table FOREIGN KEY (table_id) REFERENCES tables(id) ON DELETE CASCADE,
    CONSTRAINT fk_client_sessions_branch FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_client_sessions_session_key
    ON client_sessions(session_key)
    WHERE is_deleted = FALSE;

CREATE INDEX IF NOT EXISTS idx_client_sessions_table_reference
    ON client_sessions(table_reference)
    WHERE is_deleted = FALSE;

CREATE INDEX IF NOT EXISTS idx_client_sessions_expires_at
    ON client_sessions(expires_at)
    WHERE is_deleted = FALSE;

COMMENT ON TABLE client_sessions IS 'Ephemeral guest sessions created when a table QR is scanned';
COMMENT ON COLUMN client_sessions.session_key IS 'Random client identifier used for order APIs (distinct from table QR reference)';
COMMENT ON COLUMN client_sessions.table_reference IS 'Permanent table QR reference used to load the menu';
COMMENT ON COLUMN client_sessions.expires_at IS 'Session expiry; new scan after expiry creates a new session';

CREATE TRIGGER update_client_sessions_updated_at
    BEFORE UPDATE ON client_sessions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

COMMENT ON COLUMN orders.session_key IS 'Guest session key from client_sessions; used for client get/list';
