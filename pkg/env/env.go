package env

import "github.com/joho/godotenv"

var Env map[string]string

func GetEnv(key, def string) string {
	val, ok := Env[key]
	if !ok {
		return def
	}

	return val
}

func SetupEnvFile() {
	envFile := ".env"
	var err error
	Env, err = godotenv.Read(envFile)
	if err != nil {
		panic(err)
	}
}
