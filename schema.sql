CREATE TABLE proxies ()
CREATE TABLE providers ()
CREATE TABLE test_urls ()
CREATE TABLE checks ()
CREATE VIEW proxy_details (
    -- IP address
    -- port number
    -- date of the last successful basic functionality test
    -- date when the address was last found in any proxy list
    -- date when the address was first found in any proxy list
)
CREATE VIEW provider_details (
    -- base address (URL) of the proxy list
    -- details on extracting the HTTP(S) proxy addresses from the list
    -- date of the last successful update of the proxy list
    -- date of the last attempt to update with the number of records found
    -- indication of an error that occurred during the update latest attempt
)
CREATE VIEW test_url_details (
    -- test URL
    -- details of functionality test validation
    -- the date of the last successful functionality test
)
