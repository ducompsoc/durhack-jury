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
	KeycloakAdminBaseUrl = GetOptEnv("KEYCLOAK_ADMIN_BASE_URL", "https://admin.auth.durhack.com")
	ClientID = GetEnv("KEYCLOAK_OAUTH2_CLIENT_ID")
	DatabaseName = GetEnv("DATABASE_NAME")
}
