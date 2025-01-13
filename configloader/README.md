# 🔧 configloader

A flexible, type-safe configuration loader for Go that populates structs
from environment variables with zero dependencies.

## Features

- **Type-safe**: Automatic type conversion for primitive types and slices
- **Nested Structs**: Support for complex nested configurations
- **Custom Type Handlers**: Extensible with custom type conversions
- **Tag Support**: Flexible environment variable naming using struct tags
- **Zero Dependencies**: Only uses Go standard library
- **Prefix Support**: Automatic environment variable prefixing
- **Built-in Support** for common types like:
  - `time.Duration`
  - `net.IP`
  - `*url.URL`
  - `*regexp.Regexp`
  - `time.Time`
  - Base64 encoded bytes
  - JSON
  - CIDR


## Usage Example

```go
type DatabaseConfig struct {
    Host     string
    Port     int
    Username string `env:"DB_USER"`
    Password string
}

type Config struct {
    Environment string
    Debug       bool
    Database    DatabaseConfig
    Timeout     time.Duration
    AllowedIPs  []net.IP
}

```

The typical load config for sensitive or typed values:

```go
func loadConfig() (*Config, error) {
    cfg := &Config{}

    dbPassword := os.Getenv("DATABASE_PASSWORD")
    if dbPassword == "" {
        return nil, errors.New("DATABASE_PASSWORD is required")
    }
    cfg.Database.Password = dbPassword

    // Load Timeout
    timeoutStr := os.Getenv("TIMEOUT")
    if timeoutStr != "" {
        timeout, err := time.ParseDuration(timeoutStr)
        if err != nil {
            return nil, fmt.Errorf("invalid TIMEOUT value: %s", timeoutStr)
        }
        cfg.Timeout = timeout
    } else {
        // Set default timeout
        cfg.Timeout = 30 * time.Second
    }

    // Load Allowed IPs
    allowedIPsStr := os.Getenv("ALLOWED_IPS")
    if allowedIPsStr != "" {
        ipStrings := strings.Split(allowedIPsStr, ",")
        cfg.AllowedIPs = make([]net.IP, 0, len(ipStrings))
        
        for _, ipStr := range ipStrings {
            ip := net.ParseIP(strings.TrimSpace(ipStr))
            if ip == nil {
                return nil, fmt.Errorf("invalid IP address: %s", ipStr)
            }
            cfg.AllowedIPs = append(cfg.AllowedIPs, ip)
        }
    }

    // Validate interdependent configurations
    if cfg.Environment == "production" {
        if cfg.Debug {
            return nil, errors.New("debug mode cannot be enabled in production")
        }
        if len(cfg.AllowedIPs) == 0 {
            return nil, errors.New("ALLOWED_IPS must be set in production")
        }
    }

    return cfg, nil
}
```

```go
# The configloader way

func main() {
    var cfg Config
    err := configloader.Load(&cfg)
    if err != nil {
        log.Fatal(err)
    }
}
```


Example .env:

```text
ENVIRONMENT=production
DEBUG=true
DATABASE_HOST=localhost
DATABASE_PORT=5432
DB_USER=admin
DATABASE_PASSWORD=secret
TIMEOUT=5s
ALLOWED_IPS=192.168.1.1,192.168.1.2
```

## 🔄 Advanced Usage

### Custom Type Handlers

```go
type Config struct {
    ServerURL url.URL
    IPAddr    net.IP
}

err := configloader.Load(&cfg,
    configloader.WithTypeHandler(func(s string) (net.IP, error) {
        return net.ParseIP(s), nil
    }),
)
```

### Custom Tags and Prefix

```go
type Config struct {
    APIKey string `json:"api_key"`
}

var cfg Config
err := configloader.Load(&cfg,
    configloader.WithNameTag("json"),
    configloader.WithPrefix("MYAPP"),
)
```

It will attempt to look for `MYAPP_API_KEY`

### Custom Environment Lookup

```go
err := configloader.Load(&cfg,
    configloader.WithEnv(func(key string) string {
        // Custom environment variable lookup logic
        return os.Getenv(strings.ToUpper(key))
    }),
)
```

## 🎯 Why configloader?

- **Simple API**: Just define your struct and load
- **Type Safety**: Compile-time type checking
- **Errors and Guarantees**: You can define what is optional or required.
- **Flexible**: Supports various naming conventions and customizations
- **Production Ready**: Well-tested and used in production
- **Maintainable**: Clear error messages and predictable behavior

## Installation

```bash
go get github.com/coopnorge/member-lib/configloader
```
