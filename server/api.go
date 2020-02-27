package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	netlifyModels "github.com/netlify/open-api/go/models"
	netlifyPlumbingModels "github.com/netlify/open-api/go/plumbing/operations"
	"golang.org/x/oauth2"
)

// ServeHTTP is starting point for plugin API starting from /plugins/netlify/XX
func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	// Set the header for json type
	w.Header().Set("Content-Type", "application/json")

	// Identify unique routes of the API
	route := r.URL.Path

	if route == "/webhook" {
		p.handleWebhooks(w, r)
	}
	// When user execute /connect go to netlify auth page
	if route == "/auth/connect" {
		p.handleRedirectUserToNetlifyAuthPage(w, r)
	}
	// When user is redirected back from Netlify to MM for access token extraction
	if route == "/auth/redirect" {
		p.handleAuthRedirectFromNetlify(w, r)
	}
	// When user selects a button from the options when user enters disconnect command
	if route == "/command/disconnect" {
		p.handleDisconnectCommandResponse(w, r)
	}
	// When user selects a site from options provided when deploy command is executed
	if route == "/command/deploy" {
		p.handleDeployCommandResponse(w, r)
	}
	// When user selects a site from lists of sites, visible when rollback command is executed
	if route == "/command/rollback-builds" {
		p.handleRollbackCommandResponse(w, r)
	}
	// When user selects a build from list of builds of a site, coming in after selecting a site from rollback command
	if route == "/command/rollback" {
		p.handleRollbackBuildSelectResponse(w, r)
	}
}

func (p *Plugin) getOAuthConfig() *oauth2.Config {
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
			// Netlify Auth URL https://docs.netlify.com/api/get-started/#authentication
			AuthURL:  NetlifyAuthURL,
			TokenURL: NetlifyTokenURL,
		},

		// RedirectURL is the URL to redirect users going through
		// the OAuth flow, after the resource owner's URLs.
		RedirectURL: fmt.Sprintf("%s/plugins/netlify/auth/redirect", *siteURL),
	}

	return oauthConfig
}

func (p *Plugin) handleRedirectUserToNetlifyAuthPage(w http.ResponseWriter, r *http.Request) {
	// Check if this url was reached from within Mattermost app
	userID := r.Header.Get("Mattermost-User-ID")
	if userID == "" {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	// Create a unique ID generated to protect against CSRF attach while auth.
	antiCSRFToken := fmt.Sprintf("%v_%v", model.NewId()[0:15], userID)

	// Store that uniqueState for later validations in redirect from oauth
	p.API.KVSet(antiCSRFToken, []byte(antiCSRFToken))

	// Get OAuth configuration
	oAuthconfig := p.getOAuthConfig()

	// Redirect user to Netlify auth URL for authentication
	http.Redirect(w, r, oAuthconfig.AuthCodeURL(antiCSRFToken), http.StatusFound)
}

func (p *Plugin) handleAuthRedirectFromNetlify(w http.ResponseWriter, r *http.Request) {
	// Check if we were redirected from MM pages
	authUserID := r.Header.Get("Mattermost-User-ID")
	if authUserID == "" {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	// Get the state "antiCSRFToken" we passes in earlier when redirecting to Netlify auth URL from redirect URL
	antiCSRFTokenInURL := r.URL.Query().Get("state")

	// Check if antiCSRFToken is the same in redirect URL as to which we passed in earlier
	antiCSRFTokenPassedEarlier, err := p.API.KVGet(antiCSRFTokenInURL)
	if err != nil {
		http.Error(w, "AntiCSRF state not found", http.StatusBadRequest)
		return
	}

	if string(antiCSRFTokenPassedEarlier) != antiCSRFTokenInURL || len(antiCSRFTokenInURL) == 0 {
		http.Error(w, "Cross-site request forgery", http.StatusForbidden)
		return
	}

	// Extract user id from the state
	userID := strings.Split(antiCSRFTokenInURL, "_")[1]

	// and then clear the KVStore off the CSRF token
	p.API.KVDelete(antiCSRFTokenInURL)

	// Check if the same user in App authenticated with Netlify
	if userID != authUserID {
		http.Error(w, "Incorrect user while authentication", http.StatusUnauthorized)
		return
	}

	// Extract the access code from the redirected url
	accessCode := r.URL.Query().Get("code")

	// Create a context
	ctx := context.Background()

	oauthConf := p.getOAuthConfig()

	// Exchange the access code for access token from netlify token url
	token, appErr := oauthConf.Exchange(ctx, accessCode)
	if appErr != nil {
		http.Error(w, appErr.Error(), http.StatusInternalServerError)
		return
	}

	// Store the accesstoken into KV store with a unique identifier i.e userid_netlifyToken
	appErr = p.setNetlifyUserAccessTokenToStore(token, authUserID)
	if appErr != nil {
		http.Error(w, "Could not store netlify credentials", http.StatusInternalServerError)
		return
	}

	// Send a welcome message via Bot
	p.sendBotPostOnDM(authUserID, SuccessfullyNetlifyConnectedMessage)

	// Get the plugin file path
	bundlePath, bundleErr := p.API.GetBundlePath()

	// Get the HTML of the page which should be shown once auth is completed
	redirectedOAuthPageHTML, fileErr := ioutil.ReadFile(filepath.Join(bundlePath, "assets", "auth-redirect.html"))

	// If any error then fallback to default HTML
	if bundleErr != nil || fileErr != nil {
		defaultRedirectedOAuthPageHTML := `
		<!DOCTYPE html>
		<html>
			<head>
			</head>
			<body>
				<p>You can safely close this page and head back to your Mattermost app</p>
			</body>
		</html>
		`
		redirectedOAuthPageHTML = []byte(defaultRedirectedOAuthPageHTML)
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write(redirectedOAuthPageHTML)
}

func (p *Plugin) handleDisconnectCommandResponse(w http.ResponseWriter, r *http.Request) {
	// Check if this was passed within Mattermost
	authUserID := r.Header.Get("Mattermost-User-ID")
	if authUserID == "" {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	// Get the information from Body which contain the interactive Message Attachment we sent from /disconnect command
	intergrationResponseFromCommand := model.PostActionIntegrationRequestFromJson(r.Body)

	userID := intergrationResponseFromCommand.UserId
	actionToBeTaken := intergrationResponseFromCommand.Context["action"].(string)
	channelID := intergrationResponseFromCommand.ChannelId
	originalPostID := intergrationResponseFromCommand.PostId
	actionSecret := p.getConfiguration().EncryptionKey
	actionSecretPassed := intergrationResponseFromCommand.Context["actionSecret"].(string)

	if actionToBeTaken == ActionDisconnectPlugin && actionSecret == actionSecretPassed {
		// Unique identifier
		accessTokenIdentifier := userID + NetlifyAuthTokenKVIdentifier

		// Delete the access token from KV store
		err := p.API.KVDelete(accessTokenIdentifier)
		if err != nil {
			p.API.DeleteEphemeralPost(userID, originalPostID)
			p.sendBotEphemeralPostWithMessageInChannel(channelID, userID, fmt.Sprintf("Couldn't disconnect to Netlify services : %v", err.Error()))
			return
		}

		// Send and override success disconnect message
		p.API.UpdateEphemeralPost(userID, &model.Post{
			Id:        originalPostID,
			UserId:    p.BotUserID,
			ChannelId: channelID,
			Message: fmt.Sprint(
				":zzz: Mattermost Netlify plugin is now disconnected\n" +
					"If you ever want to connect again, just run `/netlify connect`"),
		})
		return
	}

	if actionToBeTaken == ActionCancel && actionSecret == actionSecretPassed {
		p.API.UpdateEphemeralPost(userID, &model.Post{
			Id:        originalPostID,
			UserId:    p.BotUserID,
			ChannelId: channelID,
			Message:   fmt.Sprint(channelID + " "),
		})
		return
	}

	// If secret don't match or action is not the one we want.
	http.Error(w, "Unauthorized or unknown disconnect action detected", http.StatusInternalServerError)
	p.API.DeleteEphemeralPost(userID, originalPostID)
}

func (p *Plugin) sendBuildhookForSiteDeploy(baseWebhookURL string, branch string) error {
	emptyBody := bytes.NewBuffer([]byte{})

	webhookURL, err := url.Parse(baseWebhookURL)
	if err != nil {
		return err
	}

	// Add build url parameters
	webhookParams := url.Values{}
	webhookParams.Add("trigger_branch", branch)
	webhookParams.Add("trigger_title", MattermostNetlifyBuildHookMessage)
	webhookURL.RawQuery = webhookParams.Encode()

	request, err := http.NewRequest("POST", webhookURL.String(), emptyBody)
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")

	httpClient := p.getHTTPClient()

	response, err := httpClient.Do(request)
	if err != nil || response.StatusCode != 200 {
		return err
	}

	defer response.Body.Close()

	return nil
}

func (p *Plugin) handleDeployCommandResponse(w http.ResponseWriter, r *http.Request) {
	// Check if this was passed within Mattermost
	authUserID := r.Header.Get("Mattermost-User-ID")
	if authUserID == "" {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	// Parse the JSON
	intergrationResponseFromCommand := model.PostActionIntegrationRequestFromJson(r.Body)

	userID := intergrationResponseFromCommand.UserId
	channelID := intergrationResponseFromCommand.ChannelId

	// Get the netlify client
	netlifyClient, ctx := p.getNetlifyClient()
	netlifyClientCredentials, err := p.getNetlifyClientCredentials(userID)
	if err != nil {
		p.API.SendEphemeralPost(userID, &model.Post{
			UserId:    p.BotUserID,
			ChannelId: channelID,
			Message: fmt.Sprintf(
				":exclamation: Authentication failed\n"+
					"*Error : %v*", err.Error()),
		})
		return
	}

	// Comprises of id name branch
	selectedOption := intergrationResponseFromCommand.Context["selected_option"].(string)

	// Get the information from Body which contain the interactive Message Attachment we sent from /disconnect command
	selectedOptionsValue := strings.Fields(selectedOption)

	// Extract the selected site information
	siteID := selectedOptionsValue[0]
	siteName := selectedOptionsValue[1]
	siteBranch := selectedOptionsValue[2]

	actionSecret := p.getConfiguration().EncryptionKey
	actionSecretPassed := intergrationResponseFromCommand.Context["actionSecret"].(string)

	// If action was not initiated from within MM
	if actionSecret != actionSecretPassed {
		p.API.SendEphemeralPost(userID, &model.Post{
			UserId:    p.BotUserID,
			ChannelId: channelID,
			Message: fmt.Sprintf(
				":exclamation: Authentication failed\n"),
		})
		return
	}

	// Check if any is empty
	if len(selectedOptionsValue) == 0 || len(siteID) == 0 || len(siteName) == 0 || len(siteBranch) == 0 {
		p.API.SendEphemeralPost(userID, &model.Post{
			UserId:    p.BotUserID,
			ChannelId: channelID,
			Message: fmt.Sprintf(
				":exclamation: One of more values while selecting from dropdown were empty"),
		})
		return
	}

	// Update the message of original dropdown message post
	p.API.UpdateEphemeralPost(intergrationResponseFromCommand.UserId, &model.Post{
		Id:        intergrationResponseFromCommand.PostId,
		UserId:    p.BotUserID,
		ChannelId: intergrationResponseFromCommand.ChannelId,
		Message:   fmt.Sprintf(":loudspeaker: Mattermost Netlify Bot is preparing to deploy **%v** branch of **%v** site.", siteBranch, siteName),
	})

	// Check if build hook from Mattermost already exist
	listBuildHooksParams := &netlifyPlumbingModels.ListSiteBuildHooksParams{
		SiteID:  siteID,
		Context: ctx,
	}
	listBuildHooksResponse, err := netlifyClient.Operations.ListSiteBuildHooks(listBuildHooksParams, netlifyClientCredentials)
	if err != nil {
		p.API.SendEphemeralPost(userID, &model.Post{
			UserId:    p.BotUserID,
			ChannelId: channelID,
			Message: fmt.Sprintf(
				":exclamation: Failed to get **%v** site build hooks.\n"+
					"*Error : %v*", siteName, err.Error()),
		})
		return
	}

	// All of the build hooks available with the site
	listBuildHooks := listBuildHooksResponse.GetPayload()

	// Loop over hooks available to check if MM specific hook exists
	var mmBuildHookExists bool = false
	var existingBuildHookURL string
	for _, buildHook := range listBuildHooks {
		if buildHook.Title == MattermostNetlifyBuildHookTitle {
			mmBuildHookExists = true
			existingBuildHookURL = buildHook.URL
			break
		}
	}

	// If build hook already exists then send webhook event
	if mmBuildHookExists == true {
		err := p.sendBuildhookForSiteDeploy(existingBuildHookURL, siteBranch)
		if err != nil {
			p.API.SendEphemeralPost(userID, &model.Post{
				UserId:    p.BotUserID,
				ChannelId: channelID,
				Message: fmt.Sprintf(
					":exclamation: Failed to deploy **%v** site with Mattermost build hook.\n"+
						"*Error : %v*", siteName, err.Error()),
			})
			return
		}

		// Successfully post a message saying existing MM webhook was used to deploy site
		p.API.CreatePost(&model.Post{
			UserId:    p.BotUserID,
			ChannelId: channelID,
			Message: fmt.Sprintf(
				":satellite: Mattermost Netlify Bot has successfully asked Netlify to deploy **%v** branch of **%v** site."+
					"If you have configured notifications, you should be seeing one soon.", siteBranch, siteName),
		})
		return
	}

	// Create a MM webhook if no existing MM build hook is present
	createSiteBuildHookParams := &netlifyPlumbingModels.CreateSiteBuildHookParams{
		SiteID: siteID,
		BuildHook: &netlifyModels.BuildHook{
			Title:  MattermostNetlifyBuildHookTitle,
			Branch: siteBranch,
		},
		Context: ctx,
	}
	createdSiteBuildHookResponse, err := netlifyClient.Operations.CreateSiteBuildHook(createSiteBuildHookParams, netlifyClientCredentials)
	if err != nil {
		p.API.SendEphemeralPost(userID, &model.Post{
			UserId:    p.BotUserID,
			ChannelId: channelID,
			Message: fmt.Sprintf(
				":exclamation: Failed to create a deploy hook for **%v** site.\n"+
					"*Error : %v*", siteName, err.Error()),
		})
		return
	}

	mmBuildHookCreated := createdSiteBuildHookResponse.GetPayload()

	// Send the webhook request for new deploy on newly created webhook of MM
	err = p.sendBuildhookForSiteDeploy(mmBuildHookCreated.URL, siteBranch)
	if err != nil {
		p.API.SendEphemeralPost(userID, &model.Post{
			UserId:    p.BotUserID,
			ChannelId: channelID,
			Message: fmt.Sprintf(
				":exclamation: Failed to deploy **%v** site with Mattermost build hook.\n"+
					"*Error : %v*", siteName, err.Error()),
		})
		return
	}

	p.createBotPost(channelID, fmt.Sprintf(
		":satellite: Mattermost Netlify Bot has successfully asked Netlify to deploy **%v** branch of **%v** site.\n"+
			"If you have configured notifications, you should be seeing one soon.", siteBranch, siteName))
}

func (p *Plugin) handleRollbackCommandResponse(w http.ResponseWriter, r *http.Request) {
	// Check if this was passed within Mattermost
	authUserID := r.Header.Get("Mattermost-User-ID")
	if authUserID == "" {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	// Parse the JSON
	intergrationResponseFromCommand := model.PostActionIntegrationRequestFromJson(r.Body)

	userID := intergrationResponseFromCommand.UserId
	channelID := intergrationResponseFromCommand.ChannelId

	actionSecretPassed := intergrationResponseFromCommand.Context["actionSecret"].(string)
	actionSecret := p.getConfiguration().EncryptionKey

	// If action was not initiated from within MM
	if actionSecret != actionSecretPassed {
		p.API.SendEphemeralPost(userID, &model.Post{
			UserId:    p.BotUserID,
			ChannelId: channelID,
			Message: fmt.Sprintf(
				":exclamation: Authentication failed\n"),
		})
		return
	}

	// Check if SiteURL is defined in the app
	siteURL := p.API.GetConfig().ServiceSettings.SiteURL
	if siteURL == nil {
		p.API.SendEphemeralPost(userID, &model.Post{
			UserId:    p.BotUserID,
			ChannelId: channelID,
			Message: fmt.Sprintf(
				":exclamation: Error! Site URL is not defined in the App\n"),
		})
		return
	}

	selectedOption := intergrationResponseFromCommand.Context["selected_option"].(string)
	// Get the information from Body which contain the interactive Message Attachment we sent from /disconnect command
	selectedOptionsValue := strings.Fields(selectedOption)

	// Extract the selected site information
	siteID := selectedOptionsValue[0]
	siteName := selectedOptionsValue[1]

	// Check if any is empty
	if len(selectedOptionsValue) == 0 || len(siteID) == 0 || len(siteName) == 0 {
		p.API.SendEphemeralPost(userID, &model.Post{
			UserId:    p.BotUserID,
			ChannelId: channelID,
			Message: fmt.Sprintf(
				":exclamation: One of more values while selecting from dropdown were empty"),
		})
		return
	}

	// Get the netlify client
	netlifyClient, ctx := p.getNetlifyClient()
	netlifyClientCredentials, err := p.getNetlifyClientCredentials(userID)
	if err != nil {
		p.API.SendEphemeralPost(userID, &model.Post{
			UserId:    p.BotUserID,
			ChannelId: channelID,
			Message: fmt.Sprintf(
				":exclamation: Authentication failed\n"+
					"*Error : %v*", err.Error()),
		})
		return
	}

	originalPostID := intergrationResponseFromCommand.PostId

	// Update the message of original dropdown message post
	p.API.UpdateEphemeralPost(userID, &model.Post{
		Id:        originalPostID,
		UserId:    p.BotUserID,
		ChannelId: channelID,
		Message:   fmt.Sprintf(":one: Fetching list of 5 most recent deploys of **%v** site.", siteName),
	})

	listSiteBuildsParams := &netlifyPlumbingModels.ListSiteBuildsParams{
		SiteID:  siteID,
		Context: ctx,
	}

	listSiteDeploysResponse, err := netlifyClient.Operations.ListSiteBuilds(listSiteBuildsParams, netlifyClientCredentials)
	if err != nil {
		p.API.SendEphemeralPost(userID, &model.Post{
			UserId:    p.BotUserID,
			ChannelId: channelID,
			Message: fmt.Sprintf(
				":exclamation: Failed to get **%v** site recent deploys.\n"+
					"*Error : %v*", siteName, err.Error()),
		})
		return
	}

	// All of the previous builds of the site
	listSiteDeploys := listSiteDeploysResponse.GetPayload()

	// Create an empty array of options we will be using for dropdown
	var sitesDeployListDropdownOptions []*model.PostActionOptions

	// Create a table with just the header, rows will fill up in the loop
	var deployMarkdownTable string = MarkdownDeployListTableHeader

	var i int = 1
	for index, deploy := range listSiteDeploys {
		// Take only successfull deploys into consideration
		if deploy.Done == true && len(deploy.Error) == 0 {
			if i <= 5 {
				// To restrict only to 5 entries
				i = i + 1

				var deploySHA string = "*Deployed via webhook or manually*"
				if deploy.Sha != "" {
					deploySHA = deploy.Sha
				}

				var deployedAt string = "-"
				if deploy.CreatedAt != "" {
					lastDeployedAtParsed, err := time.Parse(NetlifyDateLayout, deploy.CreatedAt)
					if err == nil {
						deployedAt = lastDeployedAtParsed.Format(time.RFC822)
					}
				}

				// Construct table for details of deploy
				var deployTableRow string = fmt.Sprintf("| %v | %v | %v | %v |", index, deploySHA, deployedAt, deploy.DeployID)
				deployMarkdownTable = fmt.Sprintf("%v\n%v", deployMarkdownTable, deployTableRow)

				siteDeployOption := &model.PostActionOptions{
					Text:  fmt.Sprintf("To sequence No.%v", index),
					Value: fmt.Sprintf("%v %v %v", siteID, siteName, deploy.DeployID),
				}
				// Store name, id information of all the sites inside the dropdown option
				sitesDeployListDropdownOptions = append(sitesDeployListDropdownOptions, siteDeployOption)
			} else {
				break
			}
		}
	}

	// If there are no deploys, give the error message
	if len(sitesDeployListDropdownOptions) == 0 {
		p.API.SendEphemeralPost(userID, &model.Post{
			UserId:    p.BotUserID,
			ChannelId: channelID,
			Message: fmt.Sprintf(
				":white-flag: There are no valid deploys with **%v** site.\n", siteName),
		})
		return
	}

	// Post a message with detail info of top 5 builds
	p.createBotPost(channelID, fmt.Sprintf(":chains: List of latest 5 releases of **%v** Netlify site\n%v", siteName, deployMarkdownTable))

	// Construct a dropdown
	sitesDeployListDropdown := &model.PostAction{
		Type:     model.POST_ACTION_TYPE_SELECT,
		Name:     "Select a previous deploy version",
		Disabled: false,
		Options:  sitesDeployListDropdownOptions,
		Integration: &model.PostActionIntegration{
			// When the user selects an option following route will be handeled
			URL: fmt.Sprintf("%s/plugins/netlify/command/rollback", *siteURL),
			Context: map[string]interface{}{
				"actionSecret": actionSecret,
			},
		},
	}

	sitesDeployListCommandInteractiveMessage := &model.SlackAttachment{
		Title:   fmt.Sprintf("Rollback *%v* sites to previous versions", siteName),
		Text:    fmt.Sprintf("Select a deploy version of %v site which you would like to rollback to:\n", siteName),
		Actions: []*model.PostAction{sitesDeployListDropdown},
		Footer:  "Before proceding refer the table of the most recent successful deploys posted above and then make your selection",
	}

	sitesDeployListCommandPost := &model.Post{
		UserId:    p.BotUserID,
		ChannelId: channelID,
		Props: map[string]interface{}{
			"attachments": []*model.SlackAttachment{sitesDeployListCommandInteractiveMessage},
		},
	}

	p.API.SendEphemeralPost(userID, sitesDeployListCommandPost)
}

func (p *Plugin) handleRollbackBuildSelectResponse(w http.ResponseWriter, r *http.Request) {
	// Check if this was passed within Mattermost
	authUserID := r.Header.Get("Mattermost-User-ID")
	if authUserID == "" {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	// Parse the JSON
	intergrationResponseFromCommand := model.PostActionIntegrationRequestFromJson(r.Body)

	userID := intergrationResponseFromCommand.UserId
	channelID := intergrationResponseFromCommand.ChannelId

	actionSecretPassed := intergrationResponseFromCommand.Context["actionSecret"].(string)
	actionSecret := p.getConfiguration().EncryptionKey

	// If action was not initiated from within MM
	if actionSecret != actionSecretPassed {
		p.API.SendEphemeralPost(userID, &model.Post{
			UserId:    p.BotUserID,
			ChannelId: channelID,
			Message: fmt.Sprintf(
				":exclamation: Authentication failed\n"),
		})
		return
	}

	// Check if SiteURL is defined in the app
	siteURL := p.API.GetConfig().ServiceSettings.SiteURL
	if siteURL == nil {
		p.API.SendEphemeralPost(userID, &model.Post{
			UserId:    p.BotUserID,
			ChannelId: channelID,
			Message: fmt.Sprintf(
				":exclamation: Error! Site URL is not defined in the App\n"),
		})
		return
	}

	selectedOption := intergrationResponseFromCommand.Context["selected_option"].(string)
	// Get the information from Body which contain the interactive Message Attachment we sent from /disconnect command
	selectedOptionsValue := strings.Fields(selectedOption)

	// Extract the selected site information
	siteID := selectedOptionsValue[0]
	siteName := selectedOptionsValue[1]
	siteDeployID := selectedOptionsValue[2]

	// Check if any is empty
	if len(selectedOptionsValue) == 0 || len(siteID) == 0 || len(siteDeployID) == 0 {
		p.API.SendEphemeralPost(userID, &model.Post{
			UserId:    p.BotUserID,
			ChannelId: channelID,
			Message: fmt.Sprintf(
				":exclamation: One of more values while selecting from dropdown were empty"),
		})
		return
	}

	// Get the netlify client
	netlifyClient, ctx := p.getNetlifyClient()
	netlifyClientCredentials, err := p.getNetlifyClientCredentials(userID)
	if err != nil {
		p.API.SendEphemeralPost(userID, &model.Post{
			UserId:    p.BotUserID,
			ChannelId: channelID,
			Message: fmt.Sprintf(
				":exclamation: Authentication failed\n"+
					"*Error : %v*", err.Error()),
		})
		return
	}

	originalPostID := intergrationResponseFromCommand.PostId

	// Update the message of original dropdown message post
	p.API.UpdateEphemeralPost(userID, &model.Post{
		Id:        originalPostID,
		UserId:    p.BotUserID,
		ChannelId: channelID,
		Message:   fmt.Sprintf(":two: Preparing to rollback %v site to %v deploy id state", siteName, siteDeployID),
	})

	// Restore site to prev x state
	restoreSiteDeployParams := &netlifyPlumbingModels.RestoreSiteDeployParams{
		DeployID: siteDeployID,
		SiteID:   siteID,
		Context:  ctx,
	}
	netlifyClient.Operations.RestoreSiteDeploy(restoreSiteDeployParams, netlifyClientCredentials)

	// Successfully post a message we asked netlify to re deploy
	p.API.CreatePost(&model.Post{
		UserId:    p.BotUserID,
		ChannelId: channelID,
		Message: fmt.Sprintf(
			":satellite: Mattermost Netlify Bot has successfully asked Netlify to rollback **%v** site to a previously version by ID %v.\n"+
				"*Since this is an update, you probably will not receive a build notification, You can visit the URL to see if its rolled back.*", siteName, siteDeployID),
	})
}
