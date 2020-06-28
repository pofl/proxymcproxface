with check_stats as (
    select
        proxy,
        max(this_proxy_check_start) as last_success
    from checks
    where success = true
    group by proxy
),
fetch_stats as (
    select
        proxy,
        max(ts) as last_seen,
        min(ts) as first_seen
    from fetch_runs
    group by proxy
)
select
    c.proxy as proxy,
    c.last_success as last_success,
    f.last_seen as last_seen,
    f.first_seen as first_seen
from check_stats as c
join fetch_stats as f
on c.proxy = f.proxy
