package config

var (
	Port           = GetOptEnv("PORT", "3301")
	Origin         = GetOptEnv("ORIGIN", "http://jury.durhack-dev.com")
	ApiOrigin      = GetOptEnv("API_ORIGIN", "http://jury.durhack-dev.com")
	KeycloakIssuer = "https://auth.durhack.com/realms/durhack-dev"
	ClientID       = GetEnv("KEYCLOAK_OAUTH2_CLIENT_ID")
)
