SELECT
    provider_url, -- base address (URL) of the proxy list
    MAX(ts) AS last_update -- date of the last attempt to update
FROM fetch_runs
GROUP BY provider_url
