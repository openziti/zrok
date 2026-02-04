package limits

import (
	"testing"

	"github.com/openziti/zrok/v2/controller/metrics"
	"github.com/stretchr/testify/assert"
)

func TestNewInfluxReaderSimple(t *testing.T) {
	cfg := &metrics.InfluxConfig{
		Url:    "http://localhost:8086",
		Token:  "test-token",
		Org:    "test-org",
		Bucket: "test-bucket",
	}

	reader := newInfluxReader(cfg)

	assert.NotNil(t, reader)
	assert.Equal(t, cfg, reader.cfg)
	assert.NotNil(t, reader.idb)
	assert.NotNil(t, reader.queryApi)
}