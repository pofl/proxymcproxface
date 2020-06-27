-- CREATE TABLE IF NOT EXISTS checks (
--     proxy TEXT,
--     test_url TEXT,
--     check_start TIMESTAMP NOT NULL,
--     this_proxy_check_start TIMESTAMP NOT NULL,
--     success BOOLEAN NOT NULL,
--     status_code INTEGER,
--     error_msg TEXT
-- )

-- type TestURLCheckResult struct {
-- 	TestURL         string
-- 	Proxy           string
-- 	TS              time.Time
-- 	IsMostRecentRun bool
-- }

-- get the any of the most recently working proxies and indicate whether they worked during the last run
WITH last_run_ts AS (
    SELECT MAX(check_start) AS max FROM checks
)
SELECT
    test_url,
    proxy,
    this_proxy_check_start,
    check_start = (SELECT * FROM last_run_ts) AS is_most_recent_run
FROM checks
WHERE test_url = $1
    AND success IS TRUE
ORDER BY this_proxy_check_start DESC
LIMIT 1
