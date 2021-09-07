package dsn

import (
	"net"
	"strconv"
	"strings"

	"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql/internal/dsn/mysql"
	"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql/internal/dsn/postgres"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

// ParsePostgres parses the dataSourceName for a known driver. The database
// name and a slice of OpenTelemetry attributes conforming with semantic
// conventions are returned.
func Parse(driverName, dataSourceName string) (string, []attribute.KeyValue, error) {
	var (
		dbname string
		attr   []attribute.KeyValue
		err    error
	)

	switch driverName {
	case "mysql":
		dbname, attr, err = ParseMySQL(dataSourceName)
	case "postgres", "pgx":
		dbname, attr, err = ParsePostgres(dataSourceName)
	default:
		// Not supported.
		attr = append(attr, semconv.DBSystemOtherSQL)
	}

	return dbname, attr, err
}

// ParsePostgres parses the MySQL DSN string and returns the database name
// the connection is to along with a slice of OpenTelemetry attributes
// conforming with semantic conventions.
func ParseMySQL(dataSourceName string) (string, []attribute.KeyValue, error) {
	cfg, err := mysql.ParseDSN(dataSourceName)
	if err != nil {
		return "", nil, err
	}

	if cfg.Passwd != "" {
		// Redact credentials.
		cfg.Passwd = ""
	}

	name := cfg.DBName
	attr := []attribute.KeyValue{
		semconv.DBSystemMySQL,
		semconv.DBNameKey.String(name),
		semconv.DBConnectionStringKey.String(cfg.FormatDSN()),
	}

	if cfg.User != "" {
		attr = append(attr, semconv.DBUserKey.String(cfg.User))
	}

	if cfg.Net != "" {
		switch cfg.Net {
		case "pipe":
			attr = append(attr, semconv.NetTransportPipe)
		case "unix", "socket":
			attr = append(attr, semconv.NetTransportUnix)
		case "memory":
			attr = append(attr, semconv.NetTransportInProc)
		case "tcp":
			attr = append(attr, semconv.NetTransportTCP)
		}

		host, port, err := net.SplitHostPort(cfg.Addr)
		if err == nil {
			if p, err := strconv.Atoi(port); err == nil {
				attr = append(attr, semconv.NetPeerPortKey.Int(p))
			}
			if ip := net.ParseIP(host); ip == nil {
				attr = append(attr, semconv.NetPeerNameKey.String(host))
			} else {
				attr = append(attr, semconv.NetPeerIPKey.String(ip.String()))
			}
		}
	}

	return name, attr, nil
}

// ParsePostgres parses the Postgres DSN string and returns the database name
// the connection is to along with a slice of OpenTelemetry attributes
// conforming with semantic conventions.
func ParsePostgres(dataSourceName string) (string, []attribute.KeyValue, error) {
	settings, err := postgres.ParseDSN(dataSourceName)
	if err != nil {
		return "", nil, err
	}

	name := settings["database"]
	attr := []attribute.KeyValue{
		semconv.DBSystemPostgreSQL,
		semconv.DBNameKey.String(name),
	}

	if _, dsnContainsPass := settings["password"]; !dsnContainsPass {
		attr = append(attr, semconv.DBConnectionStringKey.String(dataSourceName))
		// TODO: redact the password if it exsits.
	}

	if u, ok := settings["user"]; ok {
		attr = append(attr, semconv.DBUserKey.String(u))
	}

	if hosts, ok := settings["hostaddr"]; ok && hosts != "" {
		attr = append(attr, semconv.NetTransportTCP)
		// Use the first host if multiple specified.
		host := strings.Split(hosts, ",")[0]
		if ip := net.ParseIP(host); ip != nil {
			attr = append(attr, semconv.NetPeerIPKey.String(ip.String()))
		}
	} else if hosts, ok = settings["host"]; ok && hosts != "" {
		// Use the first host if multiple specified.
		host := strings.Split(hosts, ",")[0]
		if strings.HasPrefix(host, "/") {
			attr = append(attr, semconv.NetTransportUnix)
			attr = append(attr, semconv.NetPeerNameKey.String(host))
		} else {
			attr = append(attr, semconv.NetTransportTCP)
			if ip := net.ParseIP(host); ip != nil {
				attr = append(attr, semconv.NetPeerIPKey.String(ip.String()))
			} else {
				attr = append(attr, semconv.NetPeerNameKey.String(host))
			}
		}
	}

	if ports, ok := settings["port"]; ok && ports != "" {
		// Use the first port if multiple specified.
		port := strings.Split(ports, ",")[0]
		if p, err := strconv.Atoi(port); err == nil {
			attr = append(attr, semconv.NetPeerPortKey.Int(p))
		}
	}

	return name, attr, nil
}
