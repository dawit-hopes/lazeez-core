-- Restaurant feedback: order and stay ratings from guest review forms

CREATE TABLE IF NOT EXISTS order_ratings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    session_key VARCHAR(255) NOT NULL,
    rating INTEGER NOT NULL,
    comment TEXT,
    tags TEXT[] DEFAULT '{}',
    phone_number VARCHAR(20),
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT fk_order_ratings_order FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
    CONSTRAINT fk_order_ratings_branch FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE CASCADE,
    CONSTRAINT chk_order_ratings_rating CHECK (rating >= 1 AND rating <= 5)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_order_ratings_order_id
    ON order_ratings(order_id)
    WHERE is_deleted = FALSE;

CREATE INDEX IF NOT EXISTS idx_order_ratings_branch_created
    ON order_ratings(branch_id, created_at DESC)
    WHERE is_deleted = FALSE;

CREATE INDEX IF NOT EXISTS idx_order_ratings_rating
    ON order_ratings(rating)
    WHERE is_deleted = FALSE;

CREATE TABLE IF NOT EXISTS stay_ratings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_key VARCHAR(255) NOT NULL,
    branch_id UUID NOT NULL,
    table_name VARCHAR(255),
    rating INTEGER NOT NULL,
    comment TEXT,
    tags TEXT[] DEFAULT '{}',
    phone_number VARCHAR(20),
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT fk_stay_ratings_branch FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE CASCADE,
    CONSTRAINT chk_stay_ratings_rating CHECK (rating >= 1 AND rating <= 5)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_stay_ratings_session_key
    ON stay_ratings(session_key)
    WHERE is_deleted = FALSE;

CREATE INDEX IF NOT EXISTS idx_stay_ratings_branch_created
    ON stay_ratings(branch_id, created_at DESC)
    WHERE is_deleted = FALSE;

CREATE INDEX IF NOT EXISTS idx_stay_ratings_rating
    ON stay_ratings(rating)
    WHERE is_deleted = FALSE;

CREATE TRIGGER update_order_ratings_updated_at
    BEFORE UPDATE ON order_ratings
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_stay_ratings_updated_at
    BEFORE UPDATE ON stay_ratings
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
