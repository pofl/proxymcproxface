CREATE TABLE fetch_runs (
    provider_url TEXT,
    proxy TEXT,
    ts TIMESTAMP
)
CREATE TABLE checks (
    proxy TEXT,
    test_url TEXT,
    ts TIMESTAMP NOT NULL,
    success BOOLEAN NOT NULL,
    status_code NUMBER,
    error_msg TEXT
)
CREATE VIEW proxy_details AS (
    WITH check_stats AS (
        SELECT
            proxy,
            MAX(ts) AS last_success
        FROM checks
        WHERE success = true
        GROUP BY proxy
    ),
    fetch_stats AS (
        SELECT
            proxy,
            MAX(ts) AS last_seen,
            MIN(ts) AS first_seen
        FROM fetch_runs
        GROUP BY proxy
    )
    SELECT
        c.proxy AS proxy, -- IP address, port number
        c.last_success AS last_success, -- date of the last successful basic functionality test
        f.last_seen AS last_seen, -- date when the address was last found in any proxy list
        f.first_seen AS first_seen,-- date when the address was first found in any proxy list
    FROM check_stats AS c
    JOIN fetch_stats AS f
    ON c.proxy = f.proxy
)
CREATE VIEW provider_details AS (
    SELECT
        provider_url, -- base address (URL) of the proxy list
        MAX(ts) AS last_update, -- date of the last successful update of the proxy list
        COUNT(proxy) AS last_count -- date of the last attempt to update with the number of records found
    FROM fetch_runs
    GROUP BY provider_url
)
