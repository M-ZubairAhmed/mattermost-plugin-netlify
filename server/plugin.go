package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sync"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/pkg/errors"
)

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin

	// User ID of Netlify Bot
	BotUserID string

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration
}

// OnActivate is invoked when the plugin is activated. If an error is returned, the plugin will be terminated.
// The plugin will not receive hooks until after OnActivate returns without error.
// OnConfigurationChange will be called once before OnActivate.
// https://developers.mattermost.com/extend/plugins/server/reference/#Hooks.OnActivate
func (p *Plugin) OnActivate() error {
	// Retrieves the active configuration under lock.
	config := p.getConfiguration()

	err := config.IsValid()
	if err != nil {
		return err
	}

	// RegisterCommand registers a custom slash command.
	// When the command is triggered, your plugin can fulfill it via the ExecuteCommand hook.
	err = p.API.RegisterCommand(getCommand())
	if err != nil {
		return errors.Wrap(err, "Failed to register slash command")
	}

	fmt.Print("I started")

	netlifyBot := &model.Bot{
		Username:    "netlify",
		DisplayName: "Netlify",
		Description: "Created by Mattermost Netlify Plugin",
	}

	// If not present create a Netlify Bot
	botID, err := p.Helpers.EnsureBot(netlifyBot)
	if err != nil {
		return errors.Wrap(err, "Failed to ensure Netlify bot")
	}

	// Store created ID in Plugin struct
	p.BotUserID = botID

	// Get the plugin file path
	bundlePath, err := p.API.GetBundlePath()
	if err != nil {
		return errors.Wrap(err, "Could not get bundle path")
	}

	botProfileImageName := "profile.png"

	// Retrieve Bot profile image from assets file folder
	botProfileImage, err := ioutil.ReadFile(filepath.Join(bundlePath, "assets", botProfileImageName))
	if err != nil {
		return errors.Wrap(err, "Could not get the profile image")
	}

	// Set the profile image to bot via API
	errInSetProfileImage := p.API.SetProfileImage(botID, botProfileImage)
	if errInSetProfileImage != nil {
		return errors.Wrap(err, "Could not set the profile image")
	}

	return nil
}
