package router

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"slices"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/oauth2"
	"server/auth"
	"server/config"
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

		idToken := oauth2Token.Extra("id_token").(string)
		db := ctx.MustGet("db").(*mongo.Database)
		_, err = db.Collection("token_set").UpdateOne(
			context.Background(),
			gin.H{"user_id": userInfo.Subject},
			gin.H{"$set": gin.H{
				"user_id":   userInfo.Subject,
				"token_set": oauth2Token,
				"id_token":  idToken,
			},
			},
			mongoOptions.Update().SetUpsert(true),
		)

		if err != nil {
			_ = ctx.AbortWithError(http.StatusInternalServerError, err)
			fmt.Println(err.Error())
			return
		}

		ctx.Set("user", userInfo)
		ctx.Set("user_token_set", oauth2Token)
		ctx.Set("user_id_token", idToken)
		ctx.Next()
	}
}

func HandleLoginSuccess() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userInfo := ctx.MustGet("user").(*auth.DurHackKeycloakUserInfo)
		// Handle admins
		if slices.Contains(userInfo.Groups, "/admins") {
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
		if slices.Contains(userInfo.Groups, "/judges") {
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
		// todo: add a page for this that handles non-/judge or /admin group users
		// todo: also handle if user times out session cookie and needs to re-login
		//  so requests aren't authenticated and so HTML is sent back when JSON is expected
		urlPath, err := url.JoinPath(config.Origin, "/error")
		if err != nil {
			_ = ctx.AbortWithError(http.StatusInternalServerError, err)
			fmt.Println(err.Error())
			return
		}
		url, err := url.Parse(urlPath)
		if err != nil {
			_ = ctx.AbortWithError(http.StatusInternalServerError, err)
			fmt.Println(err.Error())
			return
		}
		urlQuery := url.Query()
		urlQuery.Set("status", "403")
		urlQuery.Set("message", "Forbidden")
		urlQuery.Set("retry_href", url.JoinPath(config.ApiOrigin, "/api/auth/keycloak/login").String())
		url.RawQuery = urlQuery.Encode()

		ctx.Redirect(http.StatusFound, url.String())
		return
	}
}

func Logout() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		session.Delete("user_id")
		err := session.Save()
		if err != nil {
			_ = ctx.AbortWithError(http.StatusInternalServerError, err)
			fmt.Println(err.Error())
			return
		}

		idToken := ctx.MustGet("user_id_token").(string)

		// See https://github.com/coreos/go-oidc/pull/226#issuecomment-1130411016
		if err != nil {
			_ = ctx.AbortWithError(http.StatusInternalServerError, err)
			fmt.Println(err.Error())
			return
		}

		var claims struct {
			EndSessionURL string `json:"end_session_endpoint"`
		}
		err = auth.KeycloakOIDCProvider.Claims(&claims)
		if err != nil {
			_ = ctx.AbortWithError(http.StatusInternalServerError, err)
			fmt.Println(err.Error())
			return
		}
		// ... claims.EndSessionURL is now end session URL to use
		parsedURL, err := url.Parse(claims.EndSessionURL)
		if err != nil {
			_ = ctx.AbortWithError(http.StatusInternalServerError, err)
			fmt.Println(err.Error())
			return
		}

		queryValues := parsedURL.Query()
		queryValues.Add("id_token_hint", idToken)
		queryValues.Add("post_logout_redirect_uri", config.Origin)
		parsedURL.RawQuery = queryValues.Encode()

		ctx.Redirect(http.StatusFound, parsedURL.String())
	}
}
