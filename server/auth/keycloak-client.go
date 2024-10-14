package auth

import (
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"log"
	"net/url"
	"server/config"
)

type DurHackKeycloakUserInfo struct {
	// KeycloakUserInfo structure available at:
	//https://github.com/ducompsoc/durhack/blob/130a71ab674288cbe1a6e0e2f3a518773658bc9f/server/src/lib/keycloak-client.ts#L47
	oidc.UserInfo
	Groups         []string `json:"groups"`
	PreferredNames *string  `json:"preferred_names"` // preferred_names can be null
	FirstNames     string   `json:"first_names"`
}

func (p *DurHackKeycloakUserInfo) GetNames() string {
	if p.PreferredNames != nil {
		return *p.PreferredNames
	}
	return p.FirstNames
}

type DurHackKeycloakProvider struct {
	*oidc.Provider
}

func (p *DurHackKeycloakProvider) UserInfo(ctx context.Context, tokenSource oauth2.TokenSource) (*DurHackKeycloakUserInfo, error) {
	userInfo, err := p.Provider.UserInfo(ctx, tokenSource)
	if err != nil {
		return nil, err
	}
	durhackUserInfo := &DurHackKeycloakUserInfo{UserInfo: *userInfo}
	err = durhackUserInfo.Claims(durhackUserInfo)
	if err != nil {
		return nil, err
	}
	return durhackUserInfo, nil
}

var (
	clientSecret         = config.GetEnv("KEYCLOAK_OAUTH2_CLIENT_SECRET")
	keycloakOIDCProvider *DurHackKeycloakProvider
	keycloakOAuth2Config *oauth2.Config
	KeycloakOIDCProvider = getKeycloakOIDCProvider()
	KeycloakOAuth2Config = getKeycloakOAuth2Config()
)

func getKeycloakOIDCProvider() *DurHackKeycloakProvider {
	if keycloakOIDCProvider != nil {
		return keycloakOIDCProvider
	}

	ctx := context.Background()

	provider, err := oidc.NewProvider(ctx, config.KeycloakIssuer)
	if err != nil {
		log.Fatal(err)
	}
	keycloakOIDCProvider = &DurHackKeycloakProvider{Provider: provider}

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
