package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	netlifyModels "github.com/netlify/open-api/go/models"
	netlifyPlumbingModels "github.com/netlify/open-api/go/plumbing/operations"
)

// Custom slash commands to setup
func getCommand() *model.Command {
	return &model.Command{
		Trigger:          "netlify",
		DisplayName:      "Netlify",
		Description:      "Integration with Netlify",
		AutoComplete:     true,
		AutoCompleteDesc: "Available commands: connect, disconnect, list, list id, deploy, me, help",
		AutoCompleteHint: "[command]",
	}
}

// ExecuteCommand executes the commands registered on getCommand() via RegisterCommand hook
func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	// Obtain basecommand and its associated action and parameters
	baseCommand, action, parameters := p.transformCommandToAction(args.Command)

	// Reject any command not prefixed `netlify`
	if baseCommand != "/netlify" {
		return &model.CommandResponse{}, nil
	}

	// "/netlify connect"
	if action == "connect" {
		return p.handleConnectCommand(c, args)
	}

	// "/netlify help" or "/netlify"
	if action == "help" || action == "" {
		return p.handleHelpCommand(c, args)
	}

	// Before executing any of below commands check if user account is connected
	accessToken, err := p.getNetlifyUserAccessTokenFromStore(args.UserId)
	if err != nil || len(accessToken) == 0 {
		p.sendBotEphemeralPostWithMessage(args, fmt.Sprintf("You must connect your Netlify account first.\nPlease run `/netlify connect`"))
		return &model.CommandResponse{}, nil
	}

	// "/netlify disconnect"
	if action == "disconnect" {
		return p.handleDisconnectCommand(c, args)
	}

	// "/netlify list"
	if action == "list" {
		if len(parameters) == 0 {
			return p.handleListCommand(args, false)
		} else if len(parameters) == 1 && parameters[0] == "id" {
			return p.handleListCommand(args, true)
		} else {
			return p.handleUnknownCommand(c, args, action+" "+strings.Join(parameters, " "))
		}
	}

	// "/netlify me"
	if action == "me" {
		return p.handleMeCommand(args)
	}

	// "/netlify deploy"
	if action == "deploy" {
		return p.handleBuildCommand(args, parameters)
	}

	// "/netlify xyz"
	return p.handleUnknownCommand(c, args, action)

}

func (p *Plugin) transformCommandToAction(command string) (string, string, []string) {
	// Split the entered command based on white space
	arguments := strings.Fields(command)

	// Eg. "netlify" in command "/netlify"
	baseCommand := arguments[0]

	// Eg "connect" in command "/netlify list"
	action := ""
	if len(arguments) > 1 {
		action = arguments[1]
	}

	// Eg "detail" in command "/netlify list detail"
	parameters := []string{}
	if len(arguments) > 2 {
		parameters = arguments[2:]
	}

	return baseCommand, action, parameters
}

func (p *Plugin) handleConnectCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	// Check if SiteURL is defined in the app
	siteURL := p.API.GetConfig().ServiceSettings.SiteURL
	if siteURL == nil {
		p.sendBotEphemeralPostWithMessage(args, "Error! Site URL is not defined in the App")
		return &model.CommandResponse{}, nil
	}

	// Send an ephemeral post with the link to connect netlify
	p.sendBotEphemeralPostWithMessage(args, fmt.Sprintf("[Click here to connect your Netlify account with Mattermost.](%s/plugins/netlify/auth/connect)", *siteURL))

	return &model.CommandResponse{}, nil
}

func (p *Plugin) handleHelpCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	p.sendBotEphemeralPostWithMessage(args, fmt.Sprintf("Help"))
	return &model.CommandResponse{}, nil
}

func (p *Plugin) handleUnknownCommand(c *plugin.Context, args *model.CommandArgs, action string) (*model.CommandResponse, *model.AppError) {
	p.sendBotEphemeralPostWithMessage(args, fmt.Sprintf("Unknown command `/netlify %v`\nTo see list of commands type `/netlify help`", action))
	return &model.CommandResponse{}, nil
}

func (p *Plugin) handleDisconnectCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	// Check if SiteURL is defined in the app
	siteURL := p.API.GetConfig().ServiceSettings.SiteURL
	if siteURL == nil {
		p.sendBotEphemeralPostWithMessage(args, "Error! Site URL is not defined in the App")
		return &model.CommandResponse{}, nil
	}

	actionSecret := p.getConfiguration().EncryptionKey

	deleteButton := &model.PostAction{
		Type: "button",
		Name: "Disconnect",
		Integration: &model.PostActionIntegration{
			URL: fmt.Sprintf("%s/plugins/netlify/command/disconnect", *siteURL),
			Context: map[string]interface{}{
				"action":       ActionDisconnectPlugin,
				"actionSecret": actionSecret,
			},
		},
	}

	cancelButton := &model.PostAction{
		Type: "button",
		Name: "Cancel",
		Integration: &model.PostActionIntegration{
			URL: fmt.Sprintf("%s/plugins/netlify/command/disconnect", *siteURL),
			Context: map[string]interface{}{
				"action":       ActionCancel,
				"actionSecret": actionSecret,
			},
		},
	}

	deleteMessageAttachment := &model.SlackAttachment{
		Title: "Disconnect Netlify plugin",
		Text: ":scissors: Are you sure you would like to disconnect Netlify from Mattermost?\n" +
			"If you have any question or concerns please [report](https://github.com/M-ZubairAhmed/mattermost-plugin-netlify/issues/new)",
		Actions: []*model.PostAction{deleteButton, cancelButton},
	}

	deletePost := &model.Post{
		UserId:    p.BotUserID,
		ChannelId: args.ChannelId,
		// Message:   text,
		Props: map[string]interface{}{
			"attachments": []*model.SlackAttachment{deleteMessageAttachment},
		},
	}

	p.API.SendEphemeralPost(args.UserId, deletePost)

	return &model.CommandResponse{}, nil
}

func (p *Plugin) handleListCommand(args *model.CommandArgs, listInDetail bool) (*model.CommandResponse, *model.AppError) {
	userID := args.UserId

	// Get the Netlify library client for interacting with netlify api
	netlifyClient, _ := p.getNetlifyClient()

	// Get Netlify credentials
	netlifyCredentials, err := p.getNetlifyClientCredentials(userID)
	if err != nil {
		p.sendBotEphemeralPostWithMessage(args, fmt.Sprintf("Authentication failed : %v", err.Error()))
		return &model.CommandResponse{}, nil
	}

	if err != nil {
		p.sendBotEphemeralPostWithMessage(args, fmt.Sprintf("Could not list Netlify sites : %v", err.Error()))
		return &model.CommandResponse{}, nil
	}

	// Execute list site func from netlify library
	listSitesResponse, err := netlifyClient.Operations.ListSites(nil, netlifyCredentials)
	if err != nil {
		p.sendBotEphemeralPostWithMessage(args, fmt.Sprintf("Failed to receive sites list from Netlify : %v", err.Error()))
		return &model.CommandResponse{}, nil
	}

	// Get all sites from the response payload
	sites := listSitesResponse.GetPayload()

	// If user has no netlify sites
	if len(sites) == 0 {
		p.sendBotEphemeralPostWithMessage(args, fmt.Sprintf("You don't seem to have any Netlify sites"))
		return &model.CommandResponse{}, nil
	}

	// Create a table with just the header, rows will fill up in the loop
	var markdownTable string = MarkdownSiteListTableHeader
	if listInDetail == true {
		markdownTable = MarkdownSiteListDetailTableHeader
	}

	// Loop over all sites and make a row to add to table
	for _, site := range sites {
		name := "-"
		if len(site.Name) != 0 {
			name = site.Name
		}

		id := "-"
		if len(site.ID) != 0 {
			id = site.ID
		}

		url := "-"
		if len(site.URL) != 0 {
			url = site.URL
		}

		customDomain := "*none*"
		if len(site.CustomDomain) != 0 {
			customDomain = site.CustomDomain
		}

		repo := "-"
		if len(site.BuildSettings.RepoURL) != 0 {
			repo = site.BuildSettings.RepoURL
		}

		branch := "-"
		if len(site.BuildSettings.RepoBranch) != 0 {
			branch = site.BuildSettings.RepoBranch
		}

		team := "-"
		if len(site.AccountName) != 0 {
			team = site.AccountName
		}

		lastUpdatedAt := "*failed to obtain*"
		if len(site.UpdatedAt) != 0 {
			lastUpdatedAtParsed, err := time.Parse(NetlifyDateLayout, site.UpdatedAt)
			if err == nil {
				lastUpdatedAt = lastUpdatedAtParsed.Format(time.RFC822)
			}
		}

		var tableRow string = fmt.Sprintf("| %v | %v | %v | %v | %v | %v | %v |", name, url, customDomain, repo, branch, team, lastUpdatedAt)
		if listInDetail == true {
			tableRow = fmt.Sprintf("| %v | %v |", name, id)
		}

		markdownTable = fmt.Sprintf("%v\n%v", markdownTable, tableRow)
	}

	p.sendBotPostOnChannel(args, markdownTable)

	return &model.CommandResponse{}, nil
}

func (p *Plugin) handleMeCommand(args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	userID := args.UserId

	// Get the Netlify library client for interacting with netlify api
	netlifyClient, _ := p.getNetlifyClient()

	// Get Netlify credentials
	netlifyCredentials, err := p.getNetlifyClientCredentials(userID)
	if err != nil {
		p.sendBotEphemeralPostWithMessage(args, fmt.Sprintf("Authentication failed : %v", err.Error()))
		return &model.CommandResponse{}, nil
	}

	// Execute netlify api to get account details
	currentUserResponse, err := netlifyClient.Operations.GetAccount(nil, netlifyCredentials)
	if err != nil {
		p.sendBotEphemeralPostWithMessage(args, fmt.Sprintf("Failed to get current user: %v", err.Error()))
		return &model.CommandResponse{}, nil
	}

	// Get the user account details payload from response
	currentUser := currentUserResponse.GetPayload()[0]

	// Parse the strings into dates
	userCreatedDate := "not available"
	userCreatedDateParsed, err := time.Parse(NetlifyDateLayout, currentUser.CreatedAt)
	if err == nil {
		userCreatedDate = userCreatedDateParsed.Format(time.RFC822)
	}

	userUpdatedDate := "not available"
	userUpdatedDateParsed, err := time.Parse(NetlifyDateLayout, currentUser.UpdatedAt)
	if err == nil {
		userUpdatedDate = userUpdatedDateParsed.Format(time.RFC822)
	}

	// Construct the message with account details
	currentUserMessage := fmt.Sprintf("Details of Netlify account attached with Mattermost:\n"+
		"***\n"+
		"### Primary details\n"+
		"*Name* : **%v**\n"+
		"*Email* : **%v**\n"+
		"*Account Type* : **%v** - **%v**\n\n"+
		"### Misc details\n"+
		"*ID* : %v\n"+
		"*Roles allowed* : %v\n"+
		"*Created at* : %v\n"+
		"*Last updated* : %v\n"+
		"***",
		currentUser.Name, currentUser.BillingEmail, currentUser.Type, currentUser.TypeName,
		currentUser.ID, strings.Join(currentUser.RolesAllowed, " "), userCreatedDate, userUpdatedDate)

	p.sendBotPostOnChannel(args, currentUserMessage)

	return &model.CommandResponse{}, nil
}

func (p *Plugin) handleBuildCommand(args *model.CommandArgs, parameters []string) (*model.CommandResponse, *model.AppError) {
	// Need to input at least one site id
	if len(parameters) == 0 {
		p.sendBotEphemeralPostWithMessage(args, fmt.Sprintf(
			":warning: Please mention site id after build command\n"+
				"Eg. `/netlify deploy <site id>` , for more details run help command"))
		return &model.CommandResponse{}, nil
	}

	// Cannot build more than 1 sites at a time
	if len(parameters) > 1 {
		p.sendBotEphemeralPostWithMessage(args, fmt.Sprintf(
			":warning: Please mention only one site id for a command.\n"+
				"Eg. `/netlify deploy <site id>` , for more details run help command"))
		return &model.CommandResponse{}, nil
	}

	// Get site ID from parameters
	siteID := parameters[0]
	userID := args.UserId

	netlifyClientCredentials, err := p.getNetlifyClientCredentials(userID)
	if err != nil {
		p.sendBotEphemeralPostWithMessage(args, fmt.Sprintf(
			":exclamation: Authentication failed\n"+
				"*Error : %v*", err.Error()))
		return &model.CommandResponse{}, nil
	}

	netlifyClient, ctx := p.getNetlifyClient()

	// Construct request parameters
	getSiteParameters := &netlifyPlumbingModels.GetSiteParams{
		SiteID:  siteID,
		Context: ctx,
	}

	// Get site details
	getSiteResponse, err := netlifyClient.Operations.GetSite(getSiteParameters, netlifyClientCredentials)
	if err != nil {
		p.sendBotEphemeralPostWithMessage(args, fmt.Sprintf(
			":exclamation: Failed to get site details\n"+
				"*Error : %v*", err.Error()))
		return &model.CommandResponse{}, nil
	}

	// Site details of the passed in site
	site := getSiteResponse.GetPayload()

	siteBranch := site.BuildSettings.RepoBranch
	siteName := site.Name

	// Give a message saying we are preparing to deploy
	p.sendBotEphemeralPostWithMessage(args, fmt.Sprintf(
		":loudspeaker: Mattermost Netlify Bot is preparing to deploy **%v** branch of **%v** site.", siteBranch, siteName))

	// Check if build hook from Mattermost already exist
	listBuildHooksParams := &netlifyPlumbingModels.ListSiteBuildHooksParams{
		SiteID:  siteID,
		Context: ctx,
	}
	listBuildHooksResponse, err := netlifyClient.Operations.ListSiteBuildHooks(listBuildHooksParams, netlifyClientCredentials)
	if err != nil {
		p.sendBotEphemeralPostWithMessage(args, fmt.Sprintf(
			":exclamation: Failed to get **%v** site build hooks.\n"+
				"*Error : %v*", siteName, err.Error()))
		return &model.CommandResponse{}, nil
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
		err := p.sendWebhookForSiteBuild(existingBuildHookURL, siteBranch)
		if err != nil {
			p.sendBotEphemeralPostWithMessage(args, fmt.Sprintf(
				":exclamation: Failed to deploy **%v** site with Mattermost build hook.\n"+
					"*Error : %v*", siteName, err.Error()))
			return &model.CommandResponse{}, nil
		}

		p.sendBotPostOnChannel(args, fmt.Sprintf(
			":satellite: Mattermost Netlify Bot has successfully asked Netlify to deploy **%v** branch of **%v** site.", siteBranch, siteName))
		return &model.CommandResponse{}, nil
	}

	createSiteBuildHookParams := &netlifyPlumbingModels.CreateSiteBuildHookParams{
		SiteID: siteID,
		BuildHook: &netlifyModels.BuildHook{
			Title:  MattermostNetlifyBuildHookTitle,
			Branch: siteBranch,
		},
		Context: ctx,
	}

	// Create a MM webhook if no existing MM build hook is present
	createdSiteBuildHookResponse, err := netlifyClient.Operations.CreateSiteBuildHook(createSiteBuildHookParams, netlifyClientCredentials)
	if err != nil {
		p.sendBotEphemeralPostWithMessage(args, fmt.Sprintf(
			":exclamation: Failed to create a deploy hook for **%v** site.\n"+
				"*Error : %v*", siteName, err.Error()))
		return &model.CommandResponse{}, nil
	}

	mmBuildHookCreated := createdSiteBuildHookResponse.GetPayload()

	// Send the webhook request for new deploy on newly created webhook of MM
	err = p.sendWebhookForSiteBuild(mmBuildHookCreated.URL, siteBranch)
	if err != nil {
		p.sendBotEphemeralPostWithMessage(args, fmt.Sprintf(
			":exclamation: Failed to deploy **%v** site with Mattermost build hook.\n"+
				"*Error : %v*", siteName, err.Error()))
		return &model.CommandResponse{}, nil
	}

	p.sendBotPostOnChannel(args, fmt.Sprintf(
		":satellite: Mattermost Netlify Bot has successfully asked Netlify to deploy **%v** branch of **%v** site.", siteBranch, siteName))

	return &model.CommandResponse{}, nil
}

func (p *Plugin) sendWebhookForSiteBuild(baseWebhookURL string, branch string) error {
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
