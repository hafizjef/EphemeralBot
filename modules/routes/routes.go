package routes

import (
	"ephemeral/modules/twitter"
	"ephemeral/modules/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/satori/go.uuid"
)

var config = *twitter.Config

// InjectRayID used to inject ray id for each request
func InjectRayID(c *gin.Context) {
	c.Header("X-Ray-Id", uuid.NewV4().String())
	c.Next()
}

// InjectOnce inject middleware only once
func InjectOnce() gin.HandlerFunc {
	// Init
	return func(c *gin.Context) {
		c.Next()
	}
}

// GetSelf return current logged in User information
func GetSelf(c *gin.Context) {
	user, err := twitter.GetUser(c)
	utils.ServeError(err, c)
	c.JSON(http.StatusOK, user)
}

// GetTweet test
func GetTweet(c *gin.Context) {

	type tweet struct {
		Text      string `json:"text"`
		Created   string `json:"createdAt"`
		Retweeted bool   `json:"isRT"`
	}

	tweets := make([]*tweet, 0)

	includeRts := false
	if c.Query("rt") == "true" {
		includeRts = true
	}

	timeline, err := twitter.GetTimeline(includeRts, c)
	utils.ServeError(err, c)

	for _, t := range timeline {
		s := &tweet{t.FullText, t.CreatedAt, t.Retweeted}
		tweets = append(tweets, s)
	}

	//log.Println(tweets)

	c.JSON(http.StatusOK, gin.H{
		"tweets":       tweets,
		"tweet_counts": len(tweets),
	})
}

// DeleteAllTweets delete all user tweets
func DeleteAllTweets(c *gin.Context) {
	if err := twitter.RunDeleteJob(c); err != nil {
		c.JSON(http.StatusAccepted, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Started Delete Job",
	})
}

// StopDelete to stop delete jobs
func StopDelete(c *gin.Context) {
	if err := twitter.StopDeleteJob(c); err != nil {
		c.AbortWithStatusJSON(http.StatusAccepted, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully stopped job",
	})
}

// Home Render static React Shell
// Client need to post to "/login" to create Twitter.Client
func Home(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"token":    c.Query("oauth_token"),
		"verifier": c.Query("oauth_verifier"),
	})
}

// RequestTwitterLogin redirect to Twitter auth page
// and start oauth phase
func RequestTwitterLogin(c *gin.Context) {
	requestToken, _, err := config.RequestToken()
	utils.CatchErr(err)

	authURL, err := config.AuthorizationURL(requestToken)
	utils.CatchErr(err)

	c.Redirect(http.StatusFound, authURL.String())
}
