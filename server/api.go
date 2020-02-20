package main

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"golang.org/x/oauth2"
)

// ServeHTTP is starting point for plugin API starting from /plugins/netlify/XX
func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	// Set the header for json type
	w.Header().Set("Content-Type", "application/json")

	// Identify unique routes of the API
	route := r.URL.Path

	// oauth/connect : When user execute /connect go to netlify auth page
	if route == "/auth/connect" {
		p.redirectUserToNetlifyAuthPage(w, r)
	}
	if route == "/auth/redirect" {
		fmt.Fprint(w, "Successfully Authenticated!, you may close this tab and return to your Mattermost application. Thank you for using Netlify plugin.")
	}
}

func (p *Plugin) getOAuthConfig() *oauth2.Config {
	// Parse the Netlify Auth URL https://docs.netlify.com/api/get-started/#authentication
	authURL, _ := url.Parse("https://app.netlify.com/authorize")

	config := p.getConfiguration()

	siteURL := p.API.GetConfig().ServiceSettings.SiteURL

	// oauthConfig contains all the information for OAuth flow
	oauthConfig := &oauth2.Config{
		// ClientID is the application's ID.
		ClientID: config.NetlifyOAuthClientID,

		// ClientSecret is the application's secret.
		ClientSecret: config.NetlifyOAuthSecret,

		// Endpoint contains the resource server's token endpoint
		// URLs. These are constants specific to each server and are
		// often available via site-specific packages, such as
		// google.Endpoint or github.Endpoint.
		Endpoint: oauth2.Endpoint{
			AuthURL: authURL.String(),
		},

		// RedirectURL is the URL to redirect users going through
		// the OAuth flow, after the resource owner's URLs.
		RedirectURL: fmt.Sprintf("%s/plugins/netlify/auth/redirect", *siteURL),
	}

	return oauthConfig
}

func (p *Plugin) redirectUserToNetlifyAuthPage(w http.ResponseWriter, r *http.Request) {
	// Check if this url was reached from within Mattermost app
	userID := r.Header.Get("Mattermost-User-ID")
	if userID == "" {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
	}

	// Create a unique ID generated to protect against CSRF attach while auth.
	antiCSRFToken := fmt.Sprintf("%v_%v", model.NewId()[0:15], userID)

	// Store that uniqueState for later validations in redirect from oauth
	p.API.KVSet(antiCSRFToken, []byte(antiCSRFToken))

	// Get OAuth configuration
	conf := p.getOAuthConfig()

	netlifyAuthURL := conf.AuthCodeURL(antiCSRFToken)

	http.Redirect(w, r, netlifyAuthURL, http.StatusFound)
}
