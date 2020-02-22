package main

import (
	"fmt"
	"strings"

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
		AutoCompleteDesc: "Available commands: connect, help",
		AutoCompleteHint: "[command]",
	}
}

// ExecuteCommand executes the commands registered on getCommand() via RegisterCommand hook
func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	// Obtain basecommand and its associated action
	baseCommand, action := p.transformCommandToAction(args.Command)

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

	// "/netlify xyz"
	return p.handleUnknownCommand(c, args, action)

}

func (p *Plugin) transformCommandToAction(command string) (string, string) {
	// Split the entered command based on white space
	arguments := strings.Fields(command)

	// Eg. "netlify" in command "/netlify"
	baseCommand := arguments[0]

	// Eg "connect" in command "/netlify connect"
	action := ""
	if len(arguments) > 1 {
		action = arguments[1]
	}
	return baseCommand, action
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
