package auth

import (
	"fmt"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"log"
	"server/config"
)

var (
	clientID             = config.GetEnv("KEYCLOAK_OAUTH2_CLIENT_ID")
	clientSecret         = config.GetEnv("KEYCLOAK_OAUTH2_CLIENT_SECRET")
	keycloakOIDCProvider *oidc.Provider
	keycloakOAuth2Config *oauth2.Config
	KeycloakOIDCProvider = getKeycloakOIDCProvider()
	KeycloakOAuth2Config = getKeycloakOAuth2Config()
)

func getKeycloakOIDCProvider() *oidc.Provider {
	if keycloakOIDCProvider != nil {
		return keycloakOIDCProvider
	}

	ctx := context.Background()

	keycloakOIDCProvider, err := oidc.NewProvider(ctx, "https://auth.durhack.com/realms/durhack")
	if err != nil {
		log.Fatal(err)
	}

	return keycloakOIDCProvider
}

func getKeycloakOAuth2Config() *oauth2.Config {
	if keycloakOAuth2Config != nil {
		return keycloakOAuth2Config
	}

	keycloakOAuth2Config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     getKeycloakOIDCProvider().Endpoint(),
		RedirectURL:  fmt.Sprintf("http://localhost:%s/api/auth/keycloak/callback", config.Port),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	return keycloakOAuth2Config
}
