# Task Description

## Backend

- [X] The backend aggregates and processes the data of the various proxy lists.
- [!] It prevents the updating of the proxy lists in a too small time interval. It waits at least 10 minutes between two requests to the same proxy list provider
- [X] there must not be more than one request per second to multiple URLs of one proxy list (for example for multi page lists).
- [X] It ensures a consistent storage of the data in a database, whereas duplicates of HTTP(S) proxy addresses are avoided.
- [X] It is able to test the functionality of HTTP(S) proxies in principle (basic functionality test).
- [X] It is able to test the functionality of HTTP(S) proxies in regard to a certain test URL.
- [ ] HTTP(S) proxies are removed if they are not currently on a proxy list and the basic functionality test is currently negative and has not been positive within the past week.

## Database

- [X] The database stores the aggregated and processed data of the proxy lists.
- [!] It can also be used to store the configuration.
- [X] The frontend must never communicate directly with the database.

## REST based interface

- [X] Furthermore, it is possible to use the REST based interface to 
    - [X] initiate an update of the proxy lists 
    - [X] initiate the functionality tests of the collected HTTP(S) proxy addresses by the backend. Therefore, it is not necessary to implement a proactive service which automatically performs the update.

## Frontend

- [X] The following detail information of all collected HTTP(S) proxies can be displayed:
    - [X] IP address
    - [X] port number
    - [X] date of the last successful basic functionality test
    - [X] date when the address was last found in any proxy list
    - [X] date when the address was first found in any proxy list
- [X] The following information can be displayed for each proxy list provider:
    - [X] base address (URL) of the proxy list
    - [!] details on extracting the HTTP(S) proxy addresses from the list
    - [X] date of the last successful update of the proxy list
    - [!] date of the last attempt to update
    - [!] number of records found on last update
    - [X] indication of an error that occurred during the last update 
- [!] The following information of all available test URLs can be displayed:
    - [!] test URL
    - [!] details of functionality test validation
    - [!] the date of the last successful functionality test
- [!] At least three different working test URLs together with corresponding test validation information are available for demonstration purposes.
- [!] At least five different proxy list providers are available for demonstration purposes. At least 100 HTTP(S) proxy addresses have to be found by the system with each of this five providers. Duplicates between different lists count for each list.
