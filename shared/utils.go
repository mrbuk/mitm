package shared

import "os"

func GetOsEnv(name string, defaultStr string) string {

	env := os.Getenv(name)

	if len(env) == 0 {
		return defaultStr
	}

	return env
}
