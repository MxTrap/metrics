package config

import (
	"fmt"
	"strconv"
	"strings"
)

type HTTPConfig struct {
	Host string
	Port int
}

func (c HTTPConfig) String() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func (c *HTTPConfig) Set(flagValue string) error {
	fields := strings.SplitN(flagValue, ":", 2)
	if len(fields) != 2 {
		return fmt.Errorf("invalid net address: %s", flagValue)
	}
	c.Host = fields[0]
	var err error
	c.Port, err = strconv.Atoi(fields[1])
	if err != nil {
		return fmt.Errorf("invalid net address: %s", flagValue)
	}
	return nil
}
