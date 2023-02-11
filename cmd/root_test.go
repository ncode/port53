package cmd

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestRoot_initConfig(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "init config",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initConfig()
			assert.Equal(t, viper.GetString("bindAddr"), ":9023")
			assert.Equal(t, viper.GetString("serviceUrl"), "http://localhost:9023")
			assert.Equal(t, viper.GetString("logLevel"), "DEBUG")
			assert.Equal(t, viper.GetString("database"), "/tmp/trutinha.db")
		})
	}
}
