package auth

import (
	"context"
	"github.com/Nerzal/gocloak/v13"
	"log"
	"server/config"
	"time"
)

var (
	keycloakAdminClient                     *gocloak.GoCloak
	KeycloakAdminClient                     = getKeycloakAdminClient()
	keycloakAdminClientAccessToken          *string
	keycloakAdminClientAccessTokenExpiresAt *int64 // Seconds since epoch
)

func getKeycloakAdminClient() *gocloak.GoCloak {
	if keycloakAdminClient != nil {
		return keycloakAdminClient
	}

	ctx := context.Background()

	keycloakAdminClient = gocloak.NewClient(config.KeycloakAdminBaseUrl)
	_, err := keycloakAdminClient.LoginClient(ctx, config.ClientID, clientSecret, config.KeycloakRealm)
	if err != nil {
		log.Fatal(err)
	}

	return keycloakAdminClient
}

func GetKeycloakAdminClientAccessToken(ctx context.Context) (*string, error) {
	if keycloakAdminClientAccessToken != nil && *keycloakAdminClientAccessTokenExpiresAt <= time.Now().Unix() {
		return keycloakAdminClientAccessToken, nil
	}

	jwt, err := keycloakAdminClient.LoginClient(ctx, config.ClientID, clientSecret, config.KeycloakRealm)
	if err != nil {
		return nil, err
	}
	// 10 seconds just to accommodate request time
	*keycloakAdminClientAccessTokenExpiresAt = time.Now().Unix() + int64(jwt.ExpiresIn) - 10
	*keycloakAdminClientAccessToken = jwt.AccessToken
	return keycloakAdminClientAccessToken, nil
}
