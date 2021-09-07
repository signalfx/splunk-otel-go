// Copyright (c) 2019-2021 Jack Christensen
//
// MIT License
//
// Permission is hereby granted, free of charge, to any person obtaining
// a copy of this software and associated documentation files (the
// "Software"), to deal in the Software without restriction, including
// without limitation the rights to use, copy, modify, merge, publish,
// distribute, sublicense, and/or sell copies of the Software, and to
// permit persons to whom the Software is furnished to do so, subject to
// the following conditions:
//
// The above copyright notice and this permission notice shall be
// included in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
// OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
// WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
//
// Copied from:
// https://github.com/jackc/pgconn/blob/53f5fed36c570f0b5c98d6ec2415658c7b9bd11c/config.go

package postgres

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/user"
	"strings"
)

// ParseDSN parses the DSN string to a map of connection settings.
func ParseDSN(dsn string) (map[string]string, error) {
	if dsn == "" {
		return nil, nil
	}

	var (
		settings map[string]string
		err      error
	)

	// dsn may be a database URL or a DSN.
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		settings, err = withDefaults(parseURLSettings(dsn))
		if err != nil {
			return nil, err
		}
	} else {
		settings, err = withDefaults(parseDSNSettings(dsn))
		if err != nil {
			return nil, err
		}
	}
	return settings, nil
}

func withDefaults(settings map[string]string, err error) (map[string]string, error) {
	if err != nil {
		return settings, err
	}

	n := len(settings)
	for _, k := range []string{"host", "port", "user"} {
		if _, ok := settings[k]; !ok {
			n++
		}
	}

	merged := make(map[string]string, n)
	merged["host"] = defaultHost()
	merged["port"] = "5432"

	// Default to the OS user name. Purposely ignoring err getting user name from
	// OS. The client application will simply have to specify the user in that
	// case (which they typically will be doing anyway).
	user, err := user.Current()
	if err == nil {
		merged["user"] = user.Username
	}

	for k, v := range settings {
		merged[k] = v
	}
	return merged, nil
}

// defaultHost attempts to mimic libpq's default host. libpq uses the default unix socket location on *nix and localhost
// on Windows. The default socket location is compiled into libpq. Since pgx does not have access to that default it
// checks the existence of common locations.
func defaultHost() string {
	candidatePaths := []string{
		"/var/run/postgresql", // Debian
		"/private/tmp",        // OSX - homebrew
		"/tmp",                // standard PostgreSQL
	}

	for _, path := range candidatePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return "localhost"
}

func parseURLSettings(connString string) (map[string]string, error) {
	settings := make(map[string]string)

	url, err := url.Parse(connString)
	if err != nil {
		return nil, err
	}

	if url.User != nil {
		settings["user"] = url.User.Username()
		if password, present := url.User.Password(); present {
			settings["password"] = password
		}
	}

	// Handle multiple host:port's in url.Host by splitting them into host,host,host and port,port,port.
	var hosts []string
	var ports []string
	for _, host := range strings.Split(url.Host, ",") {
		if host == "" {
			continue
		}
		if isIPOnly(host) {
			hosts = append(hosts, strings.Trim(host, "[]"))
			continue
		}
		h, p, err := net.SplitHostPort(host)
		if err != nil {
			return nil, fmt.Errorf("failed to split host:port in '%s', err: %w", host, err)
		}
		if h != "" {
			hosts = append(hosts, h)
		}
		if p != "" {
			ports = append(ports, p)
		}
	}
	if len(hosts) > 0 {
		settings["host"] = strings.Join(hosts, ",")
	}
	if len(ports) > 0 {
		settings["port"] = strings.Join(ports, ",")
	}

	database := strings.TrimLeft(url.Path, "/")
	if database != "" {
		settings["database"] = database
	}

	nameMap := map[string]string{
		"dbname": "database",
	}

	for k, v := range url.Query() {
		if k2, present := nameMap[k]; present {
			k = k2
		}

		settings[k] = v[0]
	}

	return settings, nil
}

func isIPOnly(host string) bool {
	return net.ParseIP(strings.Trim(host, "[]")) != nil || !strings.Contains(host, ":")
}

var asciiSpace = [256]uint8{'\t': 1, '\n': 1, '\v': 1, '\f': 1, '\r': 1, ' ': 1}

func parseDSNSettings(s string) (map[string]string, error) {
	settings := make(map[string]string)

	nameMap := map[string]string{
		"dbname": "database",
	}

	for len(s) > 0 {
		var key, val string
		eqIdx := strings.IndexRune(s, '=')
		if eqIdx < 0 {
			return nil, errors.New("invalid dsn")
		}

		key = strings.Trim(s[:eqIdx], " \t\n\r\v\f")
		s = strings.TrimLeft(s[eqIdx+1:], " \t\n\r\v\f")
		if len(s) == 0 {
		} else if s[0] != '\'' {
			end := 0
			for ; end < len(s); end++ {
				if asciiSpace[s[end]] == 1 {
					break
				}
				if s[end] == '\\' {
					end++
					if end == len(s) {
						return nil, errors.New("invalid backslash")
					}
				}
			}
			val = strings.Replace(strings.Replace(s[:end], "\\\\", "\\", -1), "\\'", "'", -1)
			if end == len(s) {
				s = ""
			} else {
				s = s[end+1:]
			}
		} else { // quoted string
			s = s[1:]
			end := 0
			for ; end < len(s); end++ {
				if s[end] == '\'' {
					break
				}
				if s[end] == '\\' {
					end++
				}
			}
			if end == len(s) {
				return nil, errors.New("unterminated quoted string in connection info string")
			}
			val = strings.Replace(strings.Replace(s[:end], "\\\\", "\\", -1), "\\'", "'", -1)
			if end == len(s) {
				s = ""
			} else {
				s = s[end+1:]
			}
		}

		if k, ok := nameMap[key]; ok {
			key = k
		}

		if key == "" {
			return nil, errors.New("invalid dsn")
		}

		settings[key] = val
	}

	return settings, nil
}
