CREATE TABLE customers (
    id SERIAL PRIMARY KEY,
    customer_id VARCHAR(50) UNIQUE NOT NULL,
    username VARCHAR(100),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'customer',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE customer_profiles (
    id SERIAL PRIMARY KEY,
    customer_id INT UNIQUE REFERENCES customers(id) ON DELETE CASCADE,
    full_name VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE accounts (
    id SERIAL PRIMARY KEY,
    customer_id INT REFERENCES customers(id) ON DELETE CASCADE,
    account_no VARCHAR(50) UNIQUE NOT NULL,
    type VARCHAR(50),
    balance NUMERIC(15,2) DEFAULT 0,
    currency VARCHAR(10) DEFAULT 'IDR',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    account_id INT REFERENCES accounts(id) ON DELETE CASCADE,
    trx_id VARCHAR(100) UNIQUE,
    type VARCHAR(50),
    amount NUMERIC(15,2),
    status VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE user_activities (
    id SERIAL PRIMARY KEY,
    customer_id INT REFERENCES customers(id) ON DELETE CASCADE,
    activity_type VARCHAR(100),
    feature VARCHAR(100),
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE analytics_events (
    id SERIAL PRIMARY KEY,
    customer_id INT REFERENCES customers(id) ON DELETE CASCADE,
    event_type VARCHAR(100),
    feature VARCHAR(100),
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE segments (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT
);

CREATE TABLE user_segments (
    id SERIAL PRIMARY KEY,
    customer_id INT UNIQUE NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    segment_id INT NOT NULL REFERENCES segments(id) ON DELETE CASCADE,
    confidence FLOAT NOT NULL CHECK (confidence >= 0 AND confidence <= 1),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE recommendations (
    id SERIAL PRIMARY KEY,
    segment_id INT REFERENCES segments(id) ON DELETE CASCADE,
    feature VARCHAR(100),
    reason TEXT,
    priority INT DEFAULT 1
);

CREATE INDEX idx_accounts_customer ON accounts(customer_id);
CREATE INDEX idx_transactions_account ON transactions(account_id);
CREATE INDEX idx_user_segments_customer ON user_segments(customer_id);
CREATE INDEX idx_analytics_customer ON analytics_events(customer_id);

-- DASHBOARD
-- Tabel untuk A/B Testing
CREATE TABLE ab_tests (
    id SERIAL PRIMARY KEY,
    test_name VARCHAR(100) NOT NULL,
    feature VARCHAR(100) NOT NULL,
    variant_a VARCHAR(255) NOT NULL,
    variant_b VARCHAR(255) NOT NULL,
    start_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    end_date TIMESTAMP,
    status VARCHAR(20) DEFAULT 'active', -- active, completed, paused
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Mapping customer ke A/B test variant
CREATE TABLE ab_test_assignments (
    id SERIAL PRIMARY KEY,
    customer_id INT REFERENCES customers(id) ON DELETE CASCADE,
    ab_test_id INT REFERENCES ab_tests(id) ON DELETE CASCADE,
    variant VARCHAR(10) NOT NULL, -- 'A' atau 'B'
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(customer_id, ab_test_id)
);

-- Tabel engagement metrics per segment
CREATE TABLE engagement_metrics (
    id SERIAL PRIMARY KEY,
    segment_id INT REFERENCES segments(id) ON DELETE CASCADE,
    metric_date DATE DEFAULT CURRENT_DATE,
    total_customers INT DEFAULT 0,
    active_customers INT DEFAULT 0,
    recommendation_impressions INT DEFAULT 0,
    recommendation_clicks INT DEFAULT 0,
    engagement_rate FLOAT,
    feature VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(segment_id, metric_date, feature)
);

-- Performance personalisasi & rekomendasi
CREATE TABLE personalization_performance (
    id SERIAL PRIMARY KEY,
    segment_id INT REFERENCES segments(id),
    date_period DATE DEFAULT CURRENT_DATE,
    total_recommendations INT,
    clicked_recommendations INT,
    avg_response_time FLOAT,
    customer_satisfaction FLOAT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_ab_tests_status ON ab_tests(status);
CREATE INDEX idx_engagement_segment_date ON engagement_metrics(segment_id, metric_date);
CREATE INDEX idx_personalization_date ON personalization_performance(date_period);
