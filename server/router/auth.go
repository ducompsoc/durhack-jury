package router

import (
	"context"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/oauth2"
	"net/http"
	"net/url"
	"server/auth"
	"server/config"
	"server/models"
	"slices"
)

func getOrGenerateCodeVerifier(ctx *gin.Context) (string, error) {
	session := sessions.Default(ctx)
	codeVerifier := session.Get("keycloak_code_verifier")
	switch codeVerifier.(type) {
	case string:
		return codeVerifier.(string), nil
	}
	codeVerifier = oauth2.GenerateVerifier()
	session.Set("keycloak_code_verifier", codeVerifier)
	err := session.Save()
	if err != nil {
		return "", err
	}
	return codeVerifier.(string), nil
}

func BeginKeycloakOAuth2Flow() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		codeVerifier, err := getOrGenerateCodeVerifier(ctx)
		if err != nil {
			_ = ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		authURL := auth.KeycloakOAuth2Config.AuthCodeURL(
			"",
			oauth2.S256ChallengeOption(codeVerifier),
		)
		ctx.Redirect(http.StatusFound, authURL)
	}
}

func KeycloakOAuth2FlowCallback() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		codeVerifier := session.Get("keycloak_code_verifier")
		switch codeVerifier.(type) {
		case string:
			break
		default:
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{})
			return
		}

		session.Delete("keycloak_code_verifier")
		err := session.Save()
		if err != nil {
			_ = ctx.AbortWithError(http.StatusInternalServerError, err)
			fmt.Println(err.Error())
			return
		}

		oauth2Token, err := auth.KeycloakOAuth2Config.Exchange(
			context.Background(),
			ctx.Query("code"),
			oauth2.VerifierOption(codeVerifier.(string)),
		)
		if err != nil {
			_ = ctx.AbortWithError(http.StatusInternalServerError, err)
			fmt.Println(err.Error())
			return
		}

		userInfo, err := auth.KeycloakOIDCProvider.UserInfo(
			context.Background(),
			oauth2.StaticTokenSource(oauth2Token),
			//keycloakOAuth2Config.TokenSource(context.Background(), oauth2Token),
		)
		if err != nil {
			_ = ctx.AbortWithError(http.StatusInternalServerError, err)
			fmt.Println(err.Error())
			return
		}

		session.Set("user_id", userInfo.Subject)
		err = session.Save()
		if err != nil {
			_ = ctx.AbortWithError(http.StatusInternalServerError, err)
			fmt.Println(err.Error())
			return
		}

		db := ctx.MustGet("db").(*mongo.Database)
		_, err = db.Collection("token_set").UpdateOne(
			context.Background(),
			gin.H{"user_id": userInfo.Subject},
			gin.H{"$set": gin.H{
				"user_id":   userInfo.Subject,
				"token_set": oauth2Token,
			},
			},
			mongoOptions.Update().SetUpsert(true),
		)

		if err != nil { // lucatodo: proper error handling of database errors (fmt.Println -> logging in middleware)
			_ = ctx.AbortWithError(http.StatusInternalServerError, err)
			fmt.Println(err.Error())
			return
		}

		var claims models.UserInfoClaims
		err = userInfo.Claims(&claims)
		if err != nil {
			_ = ctx.AbortWithError(http.StatusInternalServerError, err)
			fmt.Println(err.Error())
			return
		}

		ctx.Set("user", userInfo)
		ctx.Set("user_token_set", oauth2Token)
		ctx.Set("user_claims", claims)
		ctx.Next()
	}
}

func HandleLoginSuccess() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		claims := ctx.MustGet("user_claims").(models.UserInfoClaims)
		// Handle admins
		if slices.Contains(claims.Groups, "/admins") {
			urlPath, err := url.JoinPath(config.Origin, "/admin")
			if err != nil {
				_ = ctx.AbortWithError(http.StatusInternalServerError, err)
				fmt.Println(err.Error())
				return
			}
			ctx.Redirect(http.StatusFound, urlPath)
			return
		}

		// Handle judges
		if slices.Contains(claims.Groups, "/judges") {
			urlPath, err := url.JoinPath(config.Origin, "/judge")
			if err != nil {
				_ = ctx.AbortWithError(http.StatusInternalServerError, err)
				fmt.Println(err.Error())
				return
			}
			ctx.Redirect(http.StatusFound, urlPath)
			return
		}

		// Handle everyone else
		urlPath, err := url.JoinPath(config.Origin, "/error?status=403&message=Forbidden")
		if err != nil {
			_ = ctx.AbortWithError(http.StatusInternalServerError, err)
			fmt.Println(err.Error())
			return
		}
		ctx.Redirect(http.StatusFound, urlPath)
		return
	}
}
