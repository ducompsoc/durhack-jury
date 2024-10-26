package util

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

// GetFullHostname returns the full hostname of the request, including scheme
func GetFullHostname(ctx *gin.Context) string {
	scheme := "http"
	if ctx.Request.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + ctx.Request.Host
}

// Now returns the current time as a primitive.DateTime
func Now() primitive.DateTime {
	return primitive.NewDateTimeFromTime(time.Now())
}
