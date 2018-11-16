# bgp-monitor

bgp-monitor is a prototype system designed to monitor specific AS's and their associated routes.

## Implementation

- Uses BGP update data from RIPE
- Supports multiple RIPE update data sources e.g. London, New York etc (https://www.ripe.net/analyse/internet-measurements/routing-information-service-ris/ris-raw-data)
- Uses historical BGP data to provide more specific alerting and anomoly detection
- Can be configured to highlight AS's from countries that "like" to hijack BGP traffic
- Checks internal country routes for paths external to that country
- Checks prefixes for direct hijacks e.g. AS1234567 is the end AS for 111.222.111.222

## Processing

- Downloads current AS data
- Download historic data (configurable months via config) - this only happens once
- Parse data, persists to postgres database, and hold in memory
- Checks for BGP update data every two minutes
- Parses new update data
- Performs detection on new data
- Alerts where applicable with High, Medium and Low priorities
- Updates historical data with new data
- On shutdown the historical data is persisted to postgres

## FAQ

1. - Would it alert on the recent Google "hijack"
   - Yes