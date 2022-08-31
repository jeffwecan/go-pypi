package pypi_test

import (
	"testing"

	"github.com/jeffwecan/go-pypi/pypi"
	"gotest.tools/assert"
)

func setup() *pypi.PackageIndex {
	// return pypi.NewPackageIndex("https://test.pypi.org/", "test-files.pythonhosted.org")
	return pypi.NewPackageIndex("http://127.1.2.0:5000/")
}

func TestPypi_GetLatest(t *testing.T) {
	t.Run("Testing latest func", func(t *testing.T) {
		// index := setup()
		index := pypi.NewPackageIndex("https://pypi.org")
		pkg, err := index.GetLatest("pandas")
		if err != nil {
			t.Fatalf("Error when looking up latest: %s", err)
		}
		t.Logf("%+v", pkg)
		assert.Equal(t, 1591652, pkg.LastSerial)
	})
}
func TestPypi_DownloadLatest(t *testing.T) {
	t.Run("Testing download latest func", func(t *testing.T) {
		// index := setup()
		index := pypi.NewPackageIndex("https://pypi.org")
		filename, err := index.DownloadLatest(".", "consullock")
		// pkg, err := index.GetLatest("hvac")
		if err != nil {
			t.Fatalf("Error when downloading release: %s", err)
		}
		// t.Logf("%+v", pkg)
		assert.Equal(t, "hi", filename)
	})
}

func TestPypi_GetRelease(t *testing.T) {
	t.Run("Testing release func", func(t *testing.T) {
		index := setup()
		pkg, err := index.GetRelease("hvac", "0.10.1")
		if err != nil {
			t.Fatalf("Error when looking up release: %s", err)
		}
		t.Logf("%+v", pkg)
		assert.Equal(t, 1591652, pkg.LastSerial)
	})
}

func TestPypi_DownloadRelease(t *testing.T) {
	t.Run("Testing download release func", func(t *testing.T) {
		// index := setup()
		index := pypi.NewPackageIndex("https://pypi.org")
		filename, err := index.DownloadRelease(".", "consullock", "4.2.0")
		// pkg, err = index.GetRelease("hvac", "0.10.1")
		if err != nil {
			t.Fatalf("Error when downloading release: %s", err)
		}
		// t.Logf("%+v", pkg)
		assert.Equal(t, "hi", filename)
	})
}

func TestPypi_DownloadFromRequirementsFile(t *testing.T) {
	t.Run("Testing DownloadFromRequirementsFile func", func(t *testing.T) {
		index := pypi.NewPackageIndex("https://pypi.org")
		reqs, err := index.DownloadFromRequirementsFile("../test_data/test_downloads/", "../test_data/requirements.txt")
		if err != nil {
			t.Fatalf("error when downloading from requirements: %s", err)
		}
		t.Logf("reqs: %+v", reqs)
		assert.Equal(t, 5, len(reqs))
	})
}
