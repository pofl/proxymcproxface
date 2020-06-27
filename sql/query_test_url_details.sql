-- get the any of the most recently working proxies and indicate whether they worked during the last run
SELECT
    test_url,
    proxy,
    this_proxy_check_start,
    check_start = (SELECT MAX(check_start) FROM checks) AS is_most_recent_run
FROM checks
WHERE test_url = $1
    AND success IS TRUE
ORDER BY this_proxy_check_start DESC
LIMIT 1
