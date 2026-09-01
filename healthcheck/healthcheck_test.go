package healthcheck

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthCheckReturnsNoError(t *testing.T) {
	// arrange
	var req *http.Request

	// act
	err := HealthCheck(req)

	// assert
	assert.NoError(t, err)
}
