package main

import (
	"fmt"
	"strings"
	"time"

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
		AutoCompleteDesc: "Available commands: connect, disconnect, list, list detail, me, help",
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
		} else if len(parameters) == 1 && parameters[0] == "detail" {
			return p.handleListCommand(args, true)
		} else {
			return p.handleUnknownCommand(c, args, action+" "+strings.Join(parameters, " "))
		}
	}

	// "/netlify me"
	if action == "me" {
		return p.handleMeCommand(args)
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
	p.sendBotEphemeralPostWithMessage(args, fmt.Sprintf("[Click here to link your Netlify account.](%s/plugins/netlify/auth/connect)", *siteURL))

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
	userID := args.UserId

	// Unique identifier
	accessTokenIdentifier := userID + NetlifyAuthTokenKVIdentifier

	// Delete the access token from KV store
	err := p.API.KVDelete(accessTokenIdentifier)
	if err != nil {
		p.sendBotEphemeralPostWithMessage(args, fmt.Sprintf("Couldnt disconnect to Netlify services : %v", err.Error()))
		return &model.CommandResponse{}, nil
	}

	// Send success disconnect message
	p.sendBotEphemeralPostWithMessage(args, fmt.Sprint("Mattermost Netlify plugin is now disconnected"))
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
