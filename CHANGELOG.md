# v 0.5.4 (2019-03-14)

* Fix #34,#35

# v 0.5.3 (2019-02-20)

* Added Default Timezone for get data from the Nmon Files in the config file  ( when no way to get from remote files with variable TZ or /etc/timezone file) , if eny error in this settings the  default zone "Europe/Paris" will be used, if no configuration for both local timezone will be used

# v 0.5.2 (2019-02-19)

* Added new NmonFilters (USER FILTERS) option into the device section as an comma separated REGEX array, allowing users skip some data from the nmon files. ( implemements #33) , the default will be allways "^TOP" tha ussually sends to muchs data to our databases

# v 0.5.1 (2019-01-09)

* Fix #31


# v 0.5.0 (2018-11-14)

* Updated All Dependencies to last version.
* Migrated Dependencies tool godep to dep
* Added support for set remote TimeZone no remote nmon file load.


# v 0.4.2 (2018-10-08)

* Updated SFTP base library
* Updated all Crypto base library
* upgrade go compiler to 1.11

### fixes

* #25,#27,#28

# v 0.1.X  (unreleased)
### New features.
*  Added HMC Device connection and Measurement Gathering data.
*  Added  metric/measurment spent time to gather data statistics 

### fixes

### breaking changes
