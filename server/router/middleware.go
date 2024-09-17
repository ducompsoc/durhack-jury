package router

import (
	"context"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/oauth2"
	"net/http"
	"server/auth"
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
			UserId string       `bson:"user_id"`
			Token  oauth2.Token `bson:"token_set"`
		}
		err := db.Collection("token_set").FindOne(
			context.Background(),
			gin.H{"user_id": userId},
		).Decode(&tokenSet)
		if err != nil {
			deleteSessionCookieAndNext()
			return
		}

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
			_, err = db.Collection("token_set").UpdateOne(
				context.Background(),
				gin.H{"user_id": userId},
				gin.H{"$set": gin.H{"token_set": newToken}},
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

		var userInfoClaims models.UserInfoClaims
		err = userInfo.Claims(&userInfoClaims)
		if err != nil {
			_ = ctx.AbortWithError(http.StatusInternalServerError, err)
			fmt.Println(err.Error())
			return
		}

		ctx.Set("user", userInfo)
		ctx.Set("user_token_set", newToken)
		ctx.Set("user_claims", userInfoClaims)
		ctx.Next()
	}
}

func AuthoriseJudge() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		maybeClaims, exists := ctx.Get("user_claims")
		if !exists {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		claims := maybeClaims.(models.UserInfoClaims)
		if !slices.Contains(claims.Groups, "/judges") {
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}
		ctx.Next()
	}
}

func AuthoriseAdmin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		maybeClaims, exists := ctx.Get("user_claims")
		if !exists {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		claims := maybeClaims.(models.UserInfoClaims)
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
