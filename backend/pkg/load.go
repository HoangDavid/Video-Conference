package pkg

import (
	"fmt"

	"github.com/spf13/viper"
)

func LoadStunServers() ([]string, error) {
	v := viper.New()
	v.SetConfigName("sfu")
	v.AddConfigPath("configs")
	v.SetConfigType("yaml")

	v.AutomaticEnv()

	// yaml read error handler
	err := v.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("sfu.yaml Read Error: %w", err)
	}

	// Stun server exists
	servers := v.GetStringSlice("sfu.stun_servers")
	if len(servers) == 0 {
		return nil, fmt.Errorf("no stun server was found")
	}

	return servers, nil
}
