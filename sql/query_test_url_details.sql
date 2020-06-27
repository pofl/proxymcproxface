-- get the any of the most recently working proxies and indicate whether they worked during the last run
with pre as (
    select
        test_url,
        proxy,
        success,
        this_proxy_check_start,
	    check_start,
        max(check_start) over (partition by test_url) as most_recent_run
    from checks
)
select
    test_url,
    proxy,
    this_proxy_check_start as last_success,
    check_start = most_recent_run AS is_most_recent_run
from pre
where success is true
    and test_url = $1
order by this_proxy_check_start desc
limit 1
