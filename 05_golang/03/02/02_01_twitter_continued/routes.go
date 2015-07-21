package main

import (
	"net/http"
	"time"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/user"
)

func init() {
	http.HandleFunc("/", home)
	http.HandleFunc("/tweet", tweet)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/profile/", profile)
	http.Handle("/favicon.ico", http.NotFoundHandler())
	http.Handle("/public/", http.StripPrefix("/public", http.FileServer(http.Dir("public/"))))
}

func home(res http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		profile(res, req)
		return
	}

	ctx := appengine.NewContext(req)
	u := user.Current(ctx)
	log.Infof(ctx, "user: ", u)
	// pointers can be NIL so don't use a Profile * Profile here:
	var model struct {
		Profile   Profile
		TweetList []*Tweet
	}

	if u != nil {
		profile, err := getProfileByEmail(ctx, u.Email)
		if err != nil {
			http.Redirect(res, req, "/login", 302)
			return
		}
		model.Profile = *profile
	}

	var someErr error
	model.TweetList, someErr = getLatestTweets(ctx)
	if someErr != nil {
		log.Infof(ctx, "Error getting latest tweets: %v", someErr)
		return
	}

	// TODO: get recent tweets

	renderTemplate(res, "home.html", model)
}

func tweet(res http.ResponseWriter, req *http.Request) {
	ctx := appengine.NewContext(req)
	u := user.Current(ctx)
	currentUser, err := getProfileByEmail(ctx, u.Email)
	if err != nil {
		http.Redirect(res, req, "/login", 302)
		return
	}
	testTweet := Tweet{
		Username:   currentUser.Username,
		Message:    "Get the biggest, most muscular neck ever. BIG NECK. ", //Temporary message until AJAX stuff is complete
		TimePosted: time.Now(),
	}
	putTweet(req, &testTweet)
	http.Redirect(res, req, "/", 302)
}

func login(res http.ResponseWriter, req *http.Request) {
	ctx := appengine.NewContext(req)
	u := user.Current(ctx)

	// look for the user's profile
	profile, err := getProfileByEmail(ctx, u.Email)
	// if it exists redirect
	if err == nil && profile.Username != "" {
		http.Redirect(res, req, "/"+profile.Username, 302)
		return
	}

	var model struct {
		Profile *Profile
		Error   string
	}
	model.Profile = &Profile{Email: u.Email}

	// create the profile
	username := req.FormValue("username")
	if username != "" {
		_, err = getProfileByUsername(req, username)
		// if the username is already taken
		if err == nil {
			model.Error = "username is not available"
			model.Profile.Username = username
		} else {
			model.Profile.Username = username
			// try to create the profile
			err = createProfile(req, model.Profile)
			if err != nil {
				model.Error = err.Error()
			} else {
				// on success redirect to the user's timeline
				waitForProfile(req, username)
				http.SetCookie(res, &http.Cookie{Name: "logged_in", Value: "true"})
				http.Redirect(res, req, "/"+username, 302)
				return
			}
		}
	}
	// render the login template
	renderTemplate(res, "login.html", model)
}

func profile(res http.ResponseWriter, req *http.Request) {
	// TODO: fetch recent tweets

	ctx := appengine.NewContext(req)
	u := user.Current(ctx)
	log.Infof(ctx, "user: ", u)

	var model struct {
		Profile Profile
	}

	if u != nil {
		profile, err := getProfileByEmail(ctx, u.Email)
		if err != nil {
			http.Redirect(res, req, "/login", 302)
			return
		}
		model.Profile = *profile
	}

	renderTemplate(res, "profile.html", model)

}

func logout(res http.ResponseWriter, req *http.Request) {
	http.SetCookie(res, &http.Cookie{Name: "dev_appserver_login", Value: ""})
	http.SetCookie(res, &http.Cookie{Name: "logged_in", Value: ""})
	http.Redirect(res, req, "/", 302)
}