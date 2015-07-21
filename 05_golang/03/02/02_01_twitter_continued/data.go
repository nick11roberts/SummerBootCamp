package main

import (
	"fmt"
	"net/http"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/user"
)

type Profile struct {
	Email    string
	Username string
}

type Tweet struct {
	Username   string
	Message    string
	TimePosted time.Time
}

// get profile by username
func getProfileByUsername(req *http.Request, username string) (*Profile, error) {
	ctx := appengine.NewContext(req)
	q := datastore.NewQuery("Profile").Filter("Username =", username).Limit(1)
	var profiles []Profile
	_, err := q.GetAll(ctx, &profiles)
	if err != nil {
		return nil, err
	}
	if len(profiles) == 0 {
		return nil, fmt.Errorf("profile not found")
	}
	return &profiles[0], nil
}

// get profile by email
func getProfileByEmail(ctx context.Context, email string) (*Profile, error) {
	key := datastore.NewKey(ctx, "Profile", email, 0, nil)
	var profile Profile
	err := datastore.Get(ctx, key, &profile)
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

// get profile
func getProfile(ctx context.Context) (*Profile, error) {
	u := user.Current(ctx)
	return getProfileByEmail(ctx, u.Email)
}

// get latest tweets
func getLatestTweets(ctx context.Context) ([]*Tweet, error) {

	// The Query type and its methods are used to construct a query.
	q := datastore.NewQuery("Tweet").Order("-TimePosted")

	// To retrieve the results,
	// you must execute the Query using its GetAll or Run methods.
	var tweets []*Tweet
	_, err := q.GetAll(ctx, &tweets)
	// handle error
	if err != nil {
		return nil, err
	}
	return tweets, nil
}

// get latest tweets
func getLatestTweetsByProfile(ctx context.Context) ([]*Tweet, error) {

	// The Query type and its methods are used to construct a query.
	q := datastore.NewQuery("Tweet").Order("-TimePosted")

	// To retrieve the results,
	// you must execute the Query using its GetAll or Run methods.
	var tweets []*Tweet
	_, err := q.GetAll(ctx, &tweets)
	// handle error
	if err != nil {
		return nil, err
	}
	return tweets, nil
}

// insert tweet
func putTweet(req *http.Request, tweet *Tweet) error {
	ctx := appengine.NewContext(req)
	key := datastore.NewIncompleteKey(ctx, "Tweet", nil)
	_, err := datastore.Put(ctx, key, tweet)
	return err
	// you can use memcache also to improve your consistency
}

// create profile
func createProfile(req *http.Request, profile *Profile) error {
	ctx := appengine.NewContext(req)
	key := datastore.NewKey(ctx, "Profile", profile.Email, 0, nil)
	_, err := datastore.Put(ctx, key, profile)
	return err
	// you can use memcache also to improve your consistency
}

// for eventual consistency thing
// assumption: user will show up in 10 seconds on datastore
func waitForProfile(req *http.Request, username string) error {
	deadline := time.Now().Add(time.Second * 10)
	for time.Now().Before(deadline) {
		_, err := getProfileByUsername(req, username)
		if err == nil {
			return nil
		}
		time.Sleep(time.Second * 1)
	}
	return nil
}
