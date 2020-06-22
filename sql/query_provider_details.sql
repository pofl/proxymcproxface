SELECT
    provider_url,
    MAX(ts) AS last_update
FROM fetch_runs
WHERE provider_url = $1
GROUP BY provider_url
