package util

import "os"

var Env *env = &env{}

type env struct {
}

func (*env) GetEnvDefault(key, defaultValue string) string {
	if env, isExist := os.LookupEnv(key); isExist {
		return env
	}
	return defaultValue
}
