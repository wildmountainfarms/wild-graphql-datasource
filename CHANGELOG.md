# Changelog


## 1.0.1

Fixes
* Numeric values are now exported as `float64` values, which allows alerting queries to work correctly
* An undefined variables object used to cause an error to be logged. Null/undefined `variables` field defined via provisioned queries now cause no error message to be logged.


## 1.0.0

Initial release.
