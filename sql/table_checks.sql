CREATE TABLE IF NOT EXISTS checks (
    proxy TEXT,
    test_url TEXT,
    ts TIMESTAMP NOT NULL,
    success BOOLEAN NOT NULL,
    status_code INTEGER,
    error_msg TEXT
)
