package router

import (
	"context"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/oauth2"
	"net/http"
	"server/auth"
	"server/database"
	"server/models"
	"server/util"

	"slices"
)

func Authenticate() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		userId := session.Get("user_id")
		switch userId.(type) {
		case string:
			break
		default:
			ctx.Next()
			return
		}

		deleteSessionCookieAndNext := func() {
			session.Delete("user_id")
			err := session.Save()
			if err != nil {
				_ = ctx.AbortWithError(http.StatusInternalServerError, err)
				return
			}
			ctx.Next()
		}

		db := ctx.MustGet("db").(*mongo.Database)
		var tokenSet struct {
			UserId  string       `bson:"user_id"`
			Token   oauth2.Token `bson:"token_set"`
			IdToken string       `bson:"id_token"`
		}
		err := db.Collection("token_set").FindOne(
			context.Background(),
			gin.H{"user_id": userId},
		).Decode(&tokenSet)
		if err != nil {
			deleteSessionCookieAndNext()
			return
		}

		idToken := tokenSet.IdToken
		newToken, err := auth.KeycloakOAuth2Config.TokenSource(context.Background(), &tokenSet.Token).Token()
		if util.IsNetworkError(err) {
			_ = ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		if err != nil {
			deleteSessionCookieAndNext()
			return
		}

		if newToken.AccessToken != tokenSet.Token.AccessToken {
			idToken = newToken.Extra("id_token").(string)
			_, err = db.Collection("token_set").UpdateOne(
				context.Background(),
				gin.H{"user_id": userId},
				gin.H{"$set": gin.H{
					"token_set": newToken,
					"id_token":  idToken,
				}},
			)
			if err != nil {
				_ = ctx.AbortWithError(http.StatusInternalServerError, err)
				return
			}
		}

		userInfo, err := auth.KeycloakOIDCProvider.UserInfo(
			context.Background(),
			oauth2.StaticTokenSource(newToken),
		)
		if util.IsNetworkError(err) {
			_ = ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		if err != nil {
			deleteSessionCookieAndNext()
			return
		}

		ctx.Set("user", userInfo)
		ctx.Set("user_token_set", newToken)
		ctx.Set("user_id_token", idToken)
		ctx.Next()
	}
}

func AuthoriseJudge() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		maybeUserInfo, exists := ctx.Get("user")
		if !exists {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		claims := maybeUserInfo.(*auth.DurHackKeycloakUserInfo)
		if !slices.Contains(claims.Groups, "/judges") {
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}

		// Get the database from the context
		db := ctx.MustGet("db").(*mongo.Database)
		userInfo := ctx.MustGet("user").(*auth.DurHackKeycloakUserInfo)
		judge := models.NewJudge(userInfo.Subject)

		// Insert the judge into the database
		err := database.GetOrCreateJudge(db, judge)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.Set("judge", judge)
		ctx.Next()
	}
}

func AuthoriseAdmin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		maybeUserInfo, exists := ctx.Get("user")
		if !exists {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		claims := maybeUserInfo.(*auth.DurHackKeycloakUserInfo)
		if !slices.Contains(claims.Groups, "/admins") {
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}
		ctx.Next()
	}
}

// When auth invalid, send a 401 error
func no(msg string, ctx *gin.Context) {
	ctx.AbortWithStatusJSON(401, gin.H{"error": msg})
}
