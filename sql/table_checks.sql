CREATE TABLE IF NOT EXISTS checks (
    proxy TEXT,
    test_url TEXT,
    check_start TIMESTAMP NOT NULL,
    this_proxy_check_start TIMESTAMP NOT NULL,
    success BOOLEAN NOT NULL,
    status_code INTEGER,
    error_msg TEXT
)
