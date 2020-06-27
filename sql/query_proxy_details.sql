WITH check_stats AS (
    SELECT
        proxy,
        MAX(this_proxy_check_start) AS last_success
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
    f.first_seen AS first_seen -- date when the address was first found in any proxy list
FROM check_stats AS c
JOIN fetch_stats AS f
ON c.proxy = f.proxy
