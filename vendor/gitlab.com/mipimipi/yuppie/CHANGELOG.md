## [v0.3.2](https://gitlab.com/mipimipi/yuppie/-/tags/v0.3.2) (2020-12-18)

### Added

* Device icons can be served

## [v0.3.1](https://gitlab.com/mipimipi/yuppie/-/tags/v0.3.1) (2020-12-16)

### Added

* If root device description contains an empty UUID as UDN, a new UUID is generated

### Changed

* Corrected error that led to dump when using multicast eventing

## [v0.3.0](https://gitlab.com/mipimipi/yuppie/-/tags/v0.3.0) (2020-12-13)

### Added

* HTTP requests are served also if the UPnP server is not connected. That's useful to display status information

## [v0.2.1](https://gitlab.com/mipimipi/yuppie/-/tags/v0.2.1) (2020-12-05)

### Changed

* Refined eventing

## [v0.2.0](https://gitlab.com/mipimipi/yuppie/-/tags/v0.2.0) (2020-11-30)

### Changed

* Refined error handling

### Removed

* Type `StateVars` removed. Use `map[string]StateVar` instead. 

## [v0.1.0](https://gitlab.com/mipimipi/yuppie/-/tags/v0.1.0) (2020-11-29)

* Basic functionality for UPnP servers based on UPnP Device Architecture version 2.0

