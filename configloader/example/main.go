package example

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/coopnorge/member-lib/configloader"
)

type DatabaseConfig struct {
	Host     string `default:"wow.myhost.com"`
	Port     int
	Username string `env:"DB_USER"`
	Password string
}
type Config struct {
	Environment string
	Debug       bool
	Database    DatabaseConfig
	Timeout     time.Duration
	SomeIp      *net.IP
	AllowedIPs  []*net.IP `env:"ALLOWED_IPS"`
}

// .... More services, More Config Structs, More Config Loading

// Problem Statement:
// - Configuration is essential to any application
// - Lack of proper type loading possibilities.
// - Lack of reliable, simple and easy to use libs.
// - Flexibility for loading env vars into code.
// - We want reliable errors to troubleshoot.
// - Remove redundancy (Use naming that is already written aka struct fields)

// The typical way:

func loadConfig() (*Config, error) {
	cfg := &Config{}

	// Load Timeout
	timeoutStr := os.Getenv("TIMEOUT")
	if timeoutStr != "" {
		timeout, err := time.ParseDuration(timeoutStr)
		if err != nil {
			// Tomorrow we add a new Timeout field
			// We have to change this.
			return nil, fmt.Errorf("invalid TIMEOUT value: %s", timeoutStr)
		}
		cfg.Timeout = timeout
	} else {
		// Set defaults all over the place, if values are optional.
		// This becomes tricky at large configurations.
		cfg.Timeout = 30 * time.Second
	}

	// ..... Infinite amount of loading
	// On all projects ....
	//
	// A lot of code to remember

	return cfg, nil
}

func main() {
	var cfg Config
	err := configloader.Load(
		&cfg,
		configloader.WithPrefix("ECOM"),
		configloader.WithNameTag("env"),
		// Need a specific type loaded? OR a customer type?
		configloader.WithTypeHandler(IneedAdifferentWayToLoadIps),
		configloader.WithTypeHandler(MyCustomTypeHandler),

		// and more:
		//
		// Want default values: configloader.WithDefaultTag("default")
		// SHOW DEFAULT VALUE
	)
	if err != nil {
		log.Fatal(err)
	}
}

type MyComplextType int32

func MyCustomTypeHandler(val string) (MyComplextType, error) {
	intVal, err := strconv.ParseInt(val, 10, 32)
	if err != nil {
		return MyComplextType(0), fmt.Errorf("failed to parse value %s: %w", val, err)
	}

	return MyComplextType(intVal), nil
}

func IneedAdifferentWayToLoadIps(val string) ([]*net.IP, error) {
	parsedIps := []*net.IP{}
	ips := strings.Split(val, ",")

	for _, ipStr := range ips {
		ip := net.ParseIP(strings.TrimSpace(ipStr))
		if ip == nil {
			return nil, errors.New("Oh NO we can't parse the env")
		}
		parsedIps = append(parsedIps, &ip)
	}

	return parsedIps, nil
}
