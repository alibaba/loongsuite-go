// Copyright (c) 2025 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sqlx

import (
	"fmt"
	"net"
	nurl "net/url"
	"regexp"
	"strings"
)

// DatabaseConfig contains parsed database connection information
type DatabaseConfig struct {
	Endpoint string // Connection endpoint (host:port or socket path)
	DBName   string // Database name
	Host     string // Database host
	Port     string // Database port
	User     string // Username
	Password string // Password
}

// parseDSN parses database connection string and returns configuration information
func parseDSN(driverName, dsn string) (*DatabaseConfig, error) {
	switch strings.ToLower(driverName) {
	case "mysql":
		return parseMySQL(dsn)
	case "postgres", "postgresql":
		return parsePostgres(dsn)
	case "sqlite3", "sqlite":
		return parseSQLite(dsn)
	default:
		return nil, fmt.Errorf("unsupported driver: %s", driverName)
	}
}

// parsePostgres parses PostgreSQL connection string
func parsePostgres(dsn string) (*DatabaseConfig, error) {
	// Add default protocol prefix if missing
	if !strings.Contains(dsn, "://") {
		dsn = "postgres://" + dsn
	}

	u, err := nurl.Parse(dsn)
	if err != nil {
		return nil, fmt.Errorf("invalid PostgreSQL DSN: %w", err)
	}

	if u.Scheme != "postgres" && u.Scheme != "postgresql" {
		return nil, fmt.Errorf("invalid connection protocol: %s", u.Scheme)
	}

	config := &DatabaseConfig{
		Endpoint: u.Host,
		DBName:   strings.TrimPrefix(u.Path, "/"),
	}

	// Parse host and port
	if u.Host != "" {
		if host, port, err := net.SplitHostPort(u.Host); err == nil {
			config.Host = host
			config.Port = port
		} else {
			config.Host = u.Host
			config.Port = "5432" // PostgreSQL default port
		}
	}

	// Parse username and password
	if u.User != nil {
		config.User = u.User.Username()
		if password, ok := u.User.Password(); ok {
			config.Password = password
		}
	}

	// Get database name from query parameters (for some formats)
	if config.DBName == "" {
		config.DBName = u.Query().Get("dbname")
	}

	return config, nil
}

// parseMySQL parses MySQL connection string
func parseMySQL(dsn string) (*DatabaseConfig, error) {
	config := &DatabaseConfig{}

	// Handle TCP connection format: user:password@tcp(host:port)/dbname
	if strings.Contains(dsn, "@tcp(") {
		re := regexp.MustCompile(`([^:]+):([^@]*)@tcp\(([^:]+):([^)]+)\)/([^?]*)`)
		matches := re.FindStringSubmatch(dsn)
		if len(matches) == 6 {
			config.User = matches[1]
			config.Password = matches[2]
			config.Host = matches[3]
			config.Port = matches[4]
			config.DBName = matches[5]
			config.Endpoint = config.Host + ":" + config.Port
			return config, nil
		}
	}

	// Handle Unix socket format: user:password@unix(/path/to/socket)/dbname
	if strings.Contains(dsn, "@unix(") {
		re := regexp.MustCompile(`([^:]+):([^@]*)@unix\(([^)]+)\)/([^?]*)`)
		matches := re.FindStringSubmatch(dsn)
		if len(matches) == 5 {
			config.User = matches[1]
			config.Password = matches[2]
			config.Endpoint = matches[3] // socket path
			config.DBName = matches[4]
			return config, nil
		}
	}

	// Handle simple format: user:password@host:port/dbname
	re := regexp.MustCompile(`([^:]+):([^@]*)@([^:]+):([^/]+)/([^?]*)`)
	matches := re.FindStringSubmatch(dsn)
	if len(matches) == 6 {
		config.User = matches[1]
		config.Password = matches[2]
		config.Host = matches[3]
		config.Port = matches[4]
		config.DBName = matches[5]
		config.Endpoint = config.Host + ":" + config.Port
		return config, nil
	}

	// Handle minimal format: user@host/dbname
	re = regexp.MustCompile(`([^@]*)@([^/]+)/([^?]*)`)
	matches = re.FindStringSubmatch(dsn)
	if len(matches) == 4 {
		config.User = matches[1]
		config.Host = matches[2]
		config.DBName = matches[3]
		config.Port = "3306" // MySQL default port
		config.Endpoint = config.Host + ":" + config.Port
		return config, nil
	}

	return nil, fmt.Errorf("invalid MySQL DSN: %s", dsn)
}

// parseSQLite parses SQLite connection string
func parseSQLite(dsn string) (*DatabaseConfig, error) {
	// SQLite connection string is usually a file path or :memory:
	config := &DatabaseConfig{
		DBName: dsn,
	}

	if dsn == ":memory:" {
		config.Endpoint = "in-memory"
	} else {
		config.Endpoint = "file:" + dsn
	}

	return config, nil
}
