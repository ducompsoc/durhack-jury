package config

var (
	Port           = GetOptEnv("PORT", "3301")
	Origin         = GetOptEnv("ORIGIN", "http://jury.durhack-dev.com")
	ApiOrigin      = GetOptEnv("API_ORIGIN", "http://jury.durhack-dev.com")
	KeycloakIssuer = "https://auth.durhack.com/realms/durhack-dev"
	// lucatodo: accomodate additional realm structure (i.e. admin)
	// https://github.com/ducompsoc/durhack/blob/130a71ab674288cbe1a6e0e2f3a518773658bc9f/server/src/config/default.ts#L22C3-L30C5
	ClientID = GetEnv("KEYCLOAK_OAUTH2_CLIENT_ID")
)
