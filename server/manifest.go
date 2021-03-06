// This file is automatically generated. Do not modify it manually.

package main

import (
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"
)

var manifest *model.Manifest

const manifestStr = `
{
  "id": "netlify",
  "name": "Netlify",
  "description": "Netlify's plugin for Mattermost",
  "homepage_url": "https://github.com/m-zubairahmed/mattermost-plugin-netlify",
  "support_url": "https://github.com/m-zubairahmed/mattermost-plugin-netlify/issues",
  "icon_path": "assets/icon.svg",
  "version": "0.4.0",
  "min_server_version": "5.14.0",
  "server": {
    "executables": {
      "linux-amd64": "server/dist/plugin-linux-amd64",
      "darwin-amd64": "server/dist/plugin-darwin-amd64",
      "windows-amd64": "server/dist/plugin-windows-amd64.exe"
    },
    "executable": ""
  },
  "settings_schema": {
    "header": "The Netlify plugin for Mattermost allows users to control your netlify sites right from Mattermost",
    "footer": "Made with Love and Support from Mattermost Team by Md Zubair Ahmed",
    "settings": [
      {
        "key": "NetlifyOAuthAppName",
        "display_name": "Netlify Application Name",
        "type": "text",
        "help_text": "The name given to the your OAuth application in Netlify, Please remember to add the Redirect URL in Netlify application as : \u003csiteURL\u003e/plugins/netlify/auth/redirect",
        "placeholder": "Please copy the same application name in Netlify OAuth app",
        "default": "Mattermost-Netlify-Bot"
      },
      {
        "key": "NetlifyOAuthClientID",
        "display_name": "Netlify Client ID",
        "type": "text",
        "help_text": "The OAuth Client ID generated by Netlify for your OAuth application",
        "placeholder": "Please copy client ID over from Netlify OAuth application",
        "default": null
      },
      {
        "key": "NetlifyOAuthSecret",
        "display_name": "Netlify Secret",
        "type": "text",
        "help_text": "The OAuth Secret generated by Netlify for your OAuth application",
        "placeholder": "Please copy secret over from Netlify OAuth application",
        "default": null
      },
      {
        "key": "EncryptionKey",
        "display_name": "Plugin Encryption Key",
        "type": "generated",
        "help_text": "The AES encryption key internally used in plugin to encrypt stored access tokens.",
        "placeholder": "Generate the key and store before connecting the account",
        "default": null
      },
      {
        "key": "WebhookSecret",
        "display_name": "Webhook Secret Key",
        "type": "generated",
        "help_text": "This Secret key will be used to uniquely identify incoming webhook requests from Netlify.",
        "placeholder": "Generate the key and store before connecting the account",
        "default": null
      }
    ]
  }
}
`

func init() {
	manifest = model.ManifestFromJson(strings.NewReader(manifestStr))
}
