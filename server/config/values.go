package config

var (
	Port   = GetOptEnv("PORT", "8000")
	Origin = GetOptEnv("ORIGIN", "http://localhost")
)
