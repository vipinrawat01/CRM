// Package env loads a .env file into the process environment.
//
// Deliberately dependency-free (no godotenv): this is the only piece of
// "config" the app needs, and a 20-line parser keeps go.mod free of a
// third-party module for something stdlib nearly does already. Real
// environment variables always win over .env values, so `ADDR=:9000 go run .`
// still works as expected on top of a committed .env.
package env

import (
	"bufio"
	"os"
	"strings"
)

// Load reads key=value pairs from path and sets them via os.Setenv, skipping
// blank lines, comments, and any key already present in the environment. It
// is not an error for path to not exist (e.g. .env wasn't created yet).
func Load(path string) error {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
		if key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		os.Setenv(key, value)
	}
	return scanner.Err()
}
