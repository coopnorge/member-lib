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

	"github.com/coopnorge/member-lib/configloader" // Our lib. Standalone, no other attachments.
)

// All applications depend on Configs such as:

type TypicalDBConfig struct {
	Host     string
	Port     int
	Username string
	Password string
}

type TypicalConfig struct { // very mild and small example.
	Environment string
	Debug       bool
	Database    DatabaseConfig
	Timeout     time.Duration
	SomeIp      *net.IP
	AllowedIPs  []*net.IP
}

// .... More services, More Config Structs, More Config Loading

// As configuration grows, and project _quicky_ iterates, we did not feel confident in our deployments.
// And one day... The worst happened. So:
// We set out to find a lib which we did not find.

// Problem Statement and requriements:
// - Configuration is essential to any application
// - Proper type loading, not just strings...
// - Easy to use.
// - Flexibility for loading env vars into code.
// - We want reliable errors to troubleshoot.
// - Remove redundancy (Use naming that is already written aka struct fields)
// - Provide emotional safety

// The typical way:

func loadConfig() (*TypicalConfig, error) {
	cfg := &TypicalConfig{}

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

// The configloader way:
// - Recursively load all fields
// - Add a Prefix to all vars
// - look for the tag "env" and look for that name instead.
// - Custom Handlers for my types.
//
// -> return an error with all vars that are not found so we can troubleshoot.

type DatabaseConfig struct {
	Host     string `default:"wow.myhost.com"` // Why not declare defaults right here? :)
	Port     int
	Username string `env:"USER"` // -> DATABASE_USERNAME becomes DATABASE_USER
	Password string
}

type Config struct { // very mild and small example.
	Environment      string
	Debug            bool
	Database         DatabaseConfig
	Timeout          time.Duration // -> Timeout becomes ECOM_TIMEOUT
	SomeIp           *net.IP
	AllowedIPs       []*net.IP      `env:"ALLOWED_IPS"`
	VeryComplexField MyComplextType `env:"COMP_TYPE"` // Need a specific type loaded? OR a custom type?
}

func main() {
	var cfg Config
	err := configloader.Load(
		&cfg,
		configloader.WithPrefix("ECOM"),
		configloader.WithNameTag("env"),
		configloader.WithTypeHandler(IneedAdifferentWayToLoadIps),
		configloader.WithTypeHandler(MyCustomTypeHandler),
	)
	if err != nil {
		log.Fatal(err)
	}
}

type MyComplextType int32

// Very complex type
func MyCustomTypeHandler(val string) (MyComplextType, error) {
	intVal, err := strconv.ParseInt(val, 10, 32)
	if err != nil {
		return MyComplextType(0), fmt.Errorf("failed to parse value %s: %w", val, err)
	}

	return MyComplextType(intVal), nil
}

// We don't like how we load ips.
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
