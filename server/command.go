package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/mattermost/mattermost-server/v5/mlog"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

// Custom slash commands to setup
func getCommand() *model.Command {
	return &model.Command{
		Trigger:          "netlify",
		DisplayName:      "Netlify",
		Description:      "Integration with Netlify",
		AutoComplete:     true,
		AutoCompleteDesc: "Available commands: connect, disconnect, list, list id, deploy, rollback, subscribe, unsubscribe, subscriptions, me, help",
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
		return p.handleDeployCommand(args)
	}

	// "/netlify rollback"
	if action == "rollback" {
		return p.handleRollbackCommand(args)
	}

	if action == "subscribe" {
		return p.handleSubscribeCommand(args)
	}

	if action == "unsubscribe" {
		return p.handleUnsubscribeCommand(args)
	}

	if action == "subscriptions" {
		return p.handleSubscriptionsCommand(args)
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
	p.sendMessageFromBot(args.ChannelId, "", false, "#### Netlify commands with description:\n"+HelpPost)
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
		Type: model.POST_ACTION_TYPE_BUTTON,
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
		Type: model.POST_ACTION_TYPE_BUTTON,
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

func (p *Plugin) handleDeployCommand(args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	userID := args.UserId
	actionSecret := p.getConfiguration().EncryptionKey

	// Check if SiteURL is defined in the app
	siteURL := p.API.GetConfig().ServiceSettings.SiteURL
	if siteURL == nil {
		p.sendBotEphemeralPostWithMessage(args, "Error! Site URL is not defined in the App")
		return &model.CommandResponse{}, nil
	}

	// Make command responsive and convey that we are doing work
	waitPost := &model.Post{
		UserId:    p.BotUserID,
		ChannelId: args.ChannelId,
		Message:   ":hourglass: Please hang on while we get the list of all deployable sites from your Netlify account",
	}

	p.API.SendEphemeralPost(userID, waitPost)

	netlifyClientCredentials, err := p.getNetlifyClientCredentials(userID)
	if err != nil {
		p.sendBotEphemeralPostWithMessage(args, fmt.Sprintf(
			":exclamation: Authentication failed\n"+
				"*Error : %v*", err.Error()))
		return &model.CommandResponse{}, nil
	}

	netlifyClient, _ := p.getNetlifyClient()

	// Execute list site func from netlify library
	listSitesResponse, err := netlifyClient.Operations.ListSites(nil, netlifyClientCredentials)
	if err != nil {
		p.sendBotEphemeralPostWithMessage(args, fmt.Sprintf("Failed to receive sites list from Netlify : %v", err.Error()))
		return &model.CommandResponse{}, nil
	}

	sites := listSitesResponse.GetPayload()

	// If user has no netlify sites
	if len(sites) == 0 {
		p.sendBotEphemeralPostWithMessage(args, fmt.Sprintf(":white_flag: You don't seem to have any Netlify sites"))
		return &model.CommandResponse{}, nil
	}

	// Create an empty array of options we will be using for dropdown
	var sitesListDropdown []*model.PostActionOptions

	// Loop over all the sites
	for _, site := range sites {
		siteOption := &model.PostActionOptions{
			Text:  fmt.Sprintf("%v (%v branch)", site.Name, site.BuildSettings.RepoBranch),
			Value: fmt.Sprintf("%v %v %v", site.ID, site.Name, site.BuildSettings.RepoBranch),
		}
		// Store name, id and branch information of all the sites inside the dropdown option
		sitesListDropdown = append(sitesListDropdown, siteOption)
	}

	// Construct a dropdown
	sitesDropdown := &model.PostAction{
		Type:     model.POST_ACTION_TYPE_SELECT,
		Name:     "Select a site",
		Disabled: false,
		Options:  sitesListDropdown,
		Integration: &model.PostActionIntegration{
			URL: fmt.Sprintf("%s/plugins/netlify/command/deploy", *siteURL),
			Context: map[string]interface{}{
				"actionSecret": actionSecret,
			},
		},
	}

	deployCommandInteractiveMessage := &model.SlackAttachment{
		Title:   "Deploy your Netlify sites",
		Text:    "Select a site to deploy or redeploy from the list of sites below:\n",
		Actions: []*model.PostAction{sitesDropdown},
		Footer:  "Selecting a site from the dropdown will be the final selection. Please be sure before selecting.",
	}

	deployCommandPost := &model.Post{
		UserId:    p.BotUserID,
		ChannelId: args.ChannelId,
		Props: map[string]interface{}{
			"attachments": []*model.SlackAttachment{deployCommandInteractiveMessage},
		},
	}

	// Present the user with the site dropdown
	p.API.SendEphemeralPost(userID, deployCommandPost)

	return &model.CommandResponse{}, nil
}

func (p *Plugin) handleRollbackCommand(args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	userID := args.UserId
	actionSecret := p.getConfiguration().EncryptionKey

	// Check if SiteURL is defined in the app
	siteURL := p.API.GetConfig().ServiceSettings.SiteURL
	if siteURL == nil {
		p.sendBotEphemeralPostWithMessage(args, "Error! Site URL is not defined in the App")
		return &model.CommandResponse{}, nil
	}

	// Make command responsive and convey that we are doing work
	waitPost := &model.Post{
		UserId:    p.BotUserID,
		ChannelId: args.ChannelId,
		Message:   ":hourglass: Please hang on while we get the list of all rollback enabled sites from your Netlify account",
	}

	p.API.SendEphemeralPost(userID, waitPost)

	netlifyClientCredentials, err := p.getNetlifyClientCredentials(userID)
	if err != nil {
		p.sendBotEphemeralPostWithMessage(args, fmt.Sprintf(
			":exclamation: Authentication failed\n"+
				"*Error : %v*", err.Error()))
		return &model.CommandResponse{}, nil
	}

	netlifyClient, _ := p.getNetlifyClient()

	// Execute list site func from netlify library
	listSitesResponse, err := netlifyClient.Operations.ListSites(nil, netlifyClientCredentials)
	if err != nil {
		p.sendBotEphemeralPostWithMessage(args, fmt.Sprintf("Failed to receive sites list from Netlify : %v", err.Error()))
		return &model.CommandResponse{}, nil
	}

	sites := listSitesResponse.GetPayload()

	// If user has no netlify sites
	if len(sites) == 0 {
		p.sendBotEphemeralPostWithMessage(args, fmt.Sprintf(":white_flag: You don't seem to have any Netlify sites"))
		return &model.CommandResponse{}, nil
	}

	// Create an empty array of options we will be using for dropdown
	var sitesDropdownOptions []*model.PostActionOptions

	// Loop over all the sites
	for _, site := range sites {
		siteOption := &model.PostActionOptions{
			Text:  fmt.Sprintf("%v", site.Name),
			Value: fmt.Sprintf("%v %v", site.ID, site.Name),
		}
		// Store name, id information of all the sites inside the dropdown option
		sitesDropdownOptions = append(sitesDropdownOptions, siteOption)
	}

	// Construct a dropdown
	sitesDropdown := &model.PostAction{
		Type:     model.POST_ACTION_TYPE_SELECT,
		Name:     "Select a site",
		Disabled: false,
		Options:  sitesDropdownOptions,
		Integration: &model.PostActionIntegration{
			URL: fmt.Sprintf("%s/plugins/netlify/command/rollback-builds", *siteURL),
			Context: map[string]interface{}{
				"actionSecret": actionSecret,
			},
		},
	}

	rollbackCommandInteractiveMessage := &model.SlackAttachment{
		Title:   "Rollback your Netlify sites to previous versions",
		Text:    "Select a site which you would like to rollback from the list of sites below:\n",
		Actions: []*model.PostAction{sitesDropdown},
		Footer:  "After the selection is made, you will be shown latest of maximum 5 deploys of the selected site to rollback",
	}

	rollbackCommandPost := &model.Post{
		UserId:    p.BotUserID,
		ChannelId: args.ChannelId,
		Props: map[string]interface{}{
			"attachments": []*model.SlackAttachment{rollbackCommandInteractiveMessage},
		},
	}

	// Present the user with the site dropdown
	p.API.SendEphemeralPost(userID, rollbackCommandPost)

	return &model.CommandResponse{}, nil
}

func (p *Plugin) handleSubscribeCommand(args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	channelID := args.ChannelId
	userID := args.UserId
	actionSecret := p.getConfiguration().EncryptionKey
	siteURL := p.API.GetConfig().ServiceSettings.SiteURL

	// Get the Netlify library client for interacting with netlify api
	netlifyClient, _ := p.getNetlifyClient()

	// Get Netlify credentials
	netlifyCredentials, err := p.getNetlifyClientCredentials(userID)
	if err != nil {
		p.sendMessageFromBot(channelID, userID, true,
			fmt.Sprintf("Authentication failed : %v", err.Error()),
		)
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

	// Create an empty array of options we will be using for dropdown
	var sitesDropdownOptions []*model.PostActionOptions

	// Loop over all the sites
	for _, site := range sites {
		siteOption := &model.PostActionOptions{
			Text:  fmt.Sprintf("%v", site.Name),
			Value: fmt.Sprintf("%v %v", site.ID, site.Name),
		}
		// Store name, id information of all the sites inside the dropdown option
		sitesDropdownOptions = append(sitesDropdownOptions, siteOption)
	}

	// Construct a dropdown
	sitesDropdown := &model.PostAction{
		Type:     model.POST_ACTION_TYPE_SELECT,
		Name:     "Select a site",
		Disabled: false,
		Options:  sitesDropdownOptions,
		Integration: &model.PostActionIntegration{
			URL: fmt.Sprintf("%s/plugins/netlify/command/subscribe", *siteURL),
			Context: map[string]interface{}{
				"actionSecret": actionSecret,
			},
		},
	}

	subscribeCommandAttachment := &model.SlackAttachment{
		Pretext: "Subscribe to Netlify notifications",
		Title:   "Select a site you want to subscribe build notifications for",
		Text:    "Selecting a site will subscribe current channel for build start, success and fail notifications.\n",
		Actions: []*model.PostAction{sitesDropdown},
		Footer:  "If however you don't wish to subscribe, hit the (x) cross icon on the right to dismiss this message",
	}

	subscribeCommandPost := &model.Post{
		UserId:    p.BotUserID,
		ChannelId: args.ChannelId,
		Props: map[string]interface{}{
			"attachments": []*model.SlackAttachment{subscribeCommandAttachment},
		},
	}

	// Present the user with the site dropdown
	p.API.SendEphemeralPost(userID, subscribeCommandPost)

	return &model.CommandResponse{}, nil
}

func (p *Plugin) handleUnsubscribeCommand(args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	channelID := args.ChannelId
	userID := args.UserId

	// Get the Netlify library client for interacting with netlify api
	netlifyClient, _ := p.getNetlifyClient()

	// Get Netlify credentials
	netlifyCredentials, err := p.getNetlifyClientCredentials(userID)
	if err != nil {
		p.sendMessageFromBot(channelID, userID, true,
			fmt.Sprintf("Authentication failed : %v", err.Error()),
		)
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

	// For each site, check in KV store if it has subscriptions
	for _, site := range sites {
		err := p.snapWebhookSubscriptionForSite(site.ID, channelID)
		if err != nil {
			mlog.Err(fmt.Errorf("Could not unsubscribe channel with %v site", site.Name))
			p.API.SendEphemeralPost(userID, &model.Post{
				UserId:    p.BotUserID,
				ChannelId: channelID,
				Message: fmt.Sprintf(
					":exclamation: Could not unsubscribe channel with %v site\n"+
						"*Error : %v*", site.Name, err.Error()),
			})
			return nil, &model.AppError{Message: err.Error()}
		}
	}

	p.API.CreatePost(&model.Post{
		UserId:    p.BotUserID,
		ChannelId: channelID,
		Message: fmt.Sprintf(
			":no_bell:  Successfully unsubscribed this channel for build notifications from all your netlify sites.\n"),
	})

	return &model.CommandResponse{}, nil
}

func (p *Plugin) handleSubscriptionsCommand(args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	channelID := args.ChannelId
	userID := args.UserId

	// Get the Netlify library client for interacting with netlify api
	netlifyClient, _ := p.getNetlifyClient()

	// Get Netlify credentials
	netlifyCredentials, err := p.getNetlifyClientCredentials(userID)
	if err != nil {
		p.sendMessageFromBot(channelID, userID, true,
			fmt.Sprintf("Authentication failed : %v", err.Error()),
		)
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

	postMessageForSubscriptions := "### List of subscriptions of your Netlify sites with channels\n"

	for _, site := range sites {
		channelsSubscribed, err := p.getWebhookSubscriptionForSite(site.ID)
		if err != nil {
			p.API.SendEphemeralPost(userID, &model.Post{
				UserId:    p.BotUserID,
				ChannelId: channelID,
				Message: fmt.Sprintf(
					":exclamation: Could not get subscriptions for channels with %v site\n"+
						"*Error : %v*", site.Name, err.Error()),
			})
			return nil, &model.AppError{Message: err.Error()}
		}
		if len(channelsSubscribed) != 0 {
			var siteSubscriptionBlock string = fmt.Sprintf("#### %v :\n", site.Name)
			for _, channelSubscribed := range channelsSubscribed {
				channel, err := p.API.GetChannel(channelSubscribed)
				if err != nil {
					p.API.SendEphemeralPost(userID, &model.Post{
						UserId:    p.BotUserID,
						ChannelId: channelID,
						Message: fmt.Sprintf(
							":exclamation: Could not get subscriptions for channels with %v site\n"+
								"*Error : %v*", site.Name, err.Error()),
					})
					return nil, &model.AppError{Message: err.Error()}
				}
				siteSubscriptionBlock = siteSubscriptionBlock + fmt.Sprintf("- %v\n", channel.Name)
			}
			postMessageForSubscriptions = postMessageForSubscriptions + siteSubscriptionBlock
		}
	}

	p.API.CreatePost(&model.Post{
		UserId:    p.BotUserID,
		ChannelId: channelID,
		Message:   postMessageForSubscriptions,
	})

	return &model.CommandResponse{}, nil
}
