package config

var (
	Port      = GetOptEnv("PORT", "8000")
	Origin    = GetOptEnv("ORIGIN", "http://localhost:3000")
	ApiOrigin = GetOptEnv("API_ORIGIN", "http://localhost:8000")
)
