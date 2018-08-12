package middlewares

import (
	"ephemeral/modules/twitter"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"gopkg.in/appleboy/gin-jwt.v2"
)

var key string

func init() {
	if key = os.Getenv("JWT_KEY"); key == "" {
		panic("No JWT key have been set")
	}
}

// AuthMiddleWare used to authenticate user
var AuthMiddleWare = &jwt.GinJWTMiddleware{
	Realm:         "API Zone",
	Key:           []byte(key),
	Timeout:       time.Hour * 24 * 7,
	MaxRefresh:    time.Hour * 24 * 7,
	Authenticator: auth,
}

func auth(userID, password string, c *gin.Context) (string, bool) {

	accessToken, accessSecret, err := twitter.Config.AccessToken(userID, "", password)
	if err != nil {
		return "", false
	}
	_, user, err := twitter.NewClient(accessToken, accessSecret)
	if err != nil || user == nil {
		return "", false
	}

	return user.ScreenName, true

}
