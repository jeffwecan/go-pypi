package pypi_test

import (
	"testing"

	"github.com/jeffwecan/go-pypi/pypi"
	"gotest.tools/assert"
)


func TestPypi_ParseRequirements(t *testing.T) {
	t.Run("Testing ParseRequirements func", func(t *testing.T) {
		reqs, err := pypi.ParseRequirements("../test_data/requirements.txt")
		if err != nil {
			t.Fatalf("Error when parsing requirements: %s", err)
		}
		t.Logf("%+v", reqs)
		assert.Equal(t, 1591652, 0)
	})
}
