package config

import (
	"fmt"

	"github.com/joho/godotenv"
)

var ( // this is run first before init
	Port                 string
	Origin               string
	ApiOrigin            string
	KeycloakRealm        string
	KeycloakBaseUrl      string
	KeycloakAdminBaseUrl string
	ClientID             string
	DatabaseName         string
)

func init() {
	// Load .env file into system environment variables which are then picked up below
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Did not load .env file (%s). This is expected when running in a Docker container\n", err.Error())
	}

	Port = GetOptEnv("PORT", "3301")
	Origin = GetOptEnv("ORIGIN", "http://jury.durhack-dev.com")
	ApiOrigin = GetOptEnv("API_ORIGIN", "http://jury.durhack-dev.com")
	KeycloakRealm = GetOptEnv("KEYCLOAK_REALM", "durhack-dev")
	KeycloakBaseUrl = GetOptEnv("KEYCLOAK_BASE_URL", "https://auth.durhack.com")
	// durhacktodo: accomodate additional realm structure (i.e. admin) to get user names without hard saving in Flag db etc.
	// https://github.com/ducompsoc/durhack/blob/130a71ab674288cbe1a6e0e2f3a518773658bc9f/server/src/config/default.ts#L22C3-L30C5
	KeycloakAdminBaseUrl = GetOptEnv("KEYCLOAK_ADMIN_BASE_URL", "https://admin.auth.durhack.com")
	ClientID = GetEnv("KEYCLOAK_OAUTH2_CLIENT_ID")
	DatabaseName = GetEnv("DATABASE_NAME")
}
