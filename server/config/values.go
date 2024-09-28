package config

var (
	Port           = GetOptEnv("PORT", "3301")
	Origin         = GetOptEnv("ORIGIN", "http://localhost:3300")
	ApiOrigin      = GetOptEnv("API_ORIGIN", "http://localhost:3301")
	KeycloakIssuer = "https://auth.durhack.com/realms/durhack-dev"
	ClientID       = GetEnv("KEYCLOAK_OAUTH2_CLIENT_ID")
)
