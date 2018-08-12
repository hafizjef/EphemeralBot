package twitter

import (
	"context"
	"ephemeral/modules/runners"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"

	"github.com/ChimeraCoder/anaconda"
	"github.com/dghubble/oauth1"
	"github.com/gin-gonic/gin"
	"gopkg.in/appleboy/gin-jwt.v2"
)

var clientMap = make(map[string]*anaconda.TwitterApi)
var jobs = runners.NewCancelMap()

// Config Oauth Configuration
var Config = &oauth1.Config{
	ConsumerKey:    os.Getenv("CONSUMER_KEY"),
	ConsumerSecret: os.Getenv("CONSUMER_SECRET"),
	CallbackURL:    os.Getenv("CALLBACK_URL"),
	Endpoint: oauth1.Endpoint{
		RequestTokenURL: "https://api.twitter.com/oauth/request_token",
		AuthorizeURL:    "https://api.twitter.com/oauth/authorize",
		AccessTokenURL:  "https://api.twitter.com/oauth/access_token",
	},
}

func init() {

	if Config.ConsumerKey == "" || Config.ConsumerSecret == "" || Config.CallbackURL == "" {
		panic("Cannot start with empty config")
	}

	anaconda.SetConsumerKey(Config.ConsumerKey)
	anaconda.SetConsumerSecret(Config.ConsumerSecret)
}

// NewClient return new twitter client
func NewClient(t, s string) (twitterAPI *anaconda.TwitterApi, user *anaconda.User, err error) {
	client := anaconda.NewTwitterApi(t, s)
	self, err := client.GetSelf(nil)
	if err != nil {
		return nil, nil, err
	}
	clientMap[self.ScreenName] = client
	return client, &self, nil
}

// GetClient return created client from globalmap
func GetClient(c *gin.Context) (twitterAPI *anaconda.TwitterApi, err error) {
	claims := jwt.ExtractClaims(c)
	key := claims["id"].(string)
	if client, ok := clientMap[key]; ok {
		return client, nil
	}
	return nil, errors.New("Client Not Found")
}

// GetUser return currently self logined user
func GetUser(c *gin.Context) (twitterUser *anaconda.User, err error) {
	client, err := GetClient(c)
	if err != nil {
		return &anaconda.User{}, err
	}
	user, err := client.GetSelf(nil)
	if err != nil {
		return &anaconda.User{}, err
	}
	return &user, nil
}

// GetTimeline return all user tweets
func GetTimeline(retweets bool, c *gin.Context) ([]anaconda.Tweet, error) {

	client, err := GetClient(c)
	if err != nil {
		return make([]anaconda.Tweet, 0), err
	}

	args := url.Values{}
	args.Add("count", "3200")
	args.Add("include_rts", strconv.FormatBool(retweets))

	timeline, err := client.GetUserTimeline(args)
	if err != nil {
		return make([]anaconda.Tweet, 0), err
	}
	return timeline, nil
}

func refreshTimeline(api *anaconda.TwitterApi) ([]anaconda.Tweet, error) {
	args := url.Values{}
	args.Add("count", "3200")       // Twitter only returns most recent 20 tweets by default, so override
	args.Add("include_rts", "true") // When using count argument, RTs are excluded, so include them as recommended
	timeline, err := api.GetUserTimeline(args)
	if err != nil {
		return make([]anaconda.Tweet, 0), err
	}
	return timeline, nil
}

func deleteTweets(ctx context.Context, c *gin.Context, jobID string) error {

	client, err := GetClient(c)
	if err != nil {
		jobs.Delete(jobID)
		return err
	}

	timeline, err := refreshTimeline(client)
	if err != nil {
		jobs.Delete(jobID)
		return err
	}

	//log.Println("Timeline len:", len(timeline))

	for len(timeline) > 0 {
		for _, t := range timeline {
			select {
			case <-ctx.Done():
				jobs.Delete(jobID)
				return nil
			default:
				// Delete tweet
				_, err := client.DeleteTweet(t.Id, true)
				if err != nil {
					return err
				}
			}
			// Refresh timeline to get latest tweet
			// while running the goroutine
			timeline, _ = refreshTimeline(client)
		}
		// All tweets deleted successfully
		jobs.Delete(jobID)
		return nil
	}

	// Nothing to delete | timeline = 0
	jobs.Delete(jobID)
	return nil
}

// RunDeleteJob to start delete tweets
func RunDeleteJob(c *gin.Context) error {
	user, err := GetUser(c)
	if err != nil {
		return err
	}

	jobID := user.ScreenName

	if _, ok := jobs.Get(jobID); ok {
		return fmt.Errorf("Job id: %s is already running", user.ScreenName)
	}

	ctx, cancel := context.WithCancel(context.Background())
	jobs.Set(user.ScreenName, cancel)

	go deleteTweets(ctx, c, jobID)

	return nil
}

// StopDeleteJob to stop delete job
func StopDeleteJob(c *gin.Context) error {
	user, err := GetUser(c)
	if err != nil {
		return err
	}
	jobID := user.ScreenName

	cancel, found := jobs.Get(jobID)
	if !found {
		return fmt.Errorf("Job id: %s is not running", jobID)
	}

	cancel()
	return nil
}
