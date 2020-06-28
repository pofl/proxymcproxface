with count_by_run as (
    select
        provider_url,
		ts,
        max(ts) over (partition by provider_url) as latest_fetch,
        count(*) as found
    from fetch_runs
    group by provider_url, ts
)
select
    provider_url,
    latest_fetch,
    found
from count_by_run
where latest_fetch = ts and provider_url = $1
