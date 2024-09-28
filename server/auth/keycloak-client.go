package auth

import (
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"log"
	"net/url"
	"server/config"
)

var (
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

	keycloakOIDCProvider, err := oidc.NewProvider(ctx, config.KeycloakIssuer)
	if err != nil {
		log.Fatal(err)
	}

	return keycloakOIDCProvider
}

func getKeycloakOAuth2Config() *oauth2.Config {
	if keycloakOAuth2Config != nil {
		return keycloakOAuth2Config
	}

	parsedUrl, err := url.JoinPath(config.ApiOrigin, "/api/auth/keycloak/callback")
	if err != nil {
		log.Fatalf("Failed to create keycloak callback URL: %v", err)
	}
	keycloakOAuth2Config := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: clientSecret,
		Endpoint:     getKeycloakOIDCProvider().Endpoint(),
		RedirectURL:  parsedUrl,
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	return keycloakOAuth2Config
}
