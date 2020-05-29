# go-pypi

Go client for [Warehouse / PyPI's API endpoints](https://warehouse.readthedocs.io/api-reference/#).


## Example

Downloading a specific package version to the current working directory from pypi.org:

```go
index := pypi.NewPackageIndex("https://pypi.org")
filename, err := index.DownloadRelease(".", "hvac", "0.10.1")
```
