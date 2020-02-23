<p align="center">
        <img src="https://user-images.githubusercontent.com/17708702/75109618-f18f6100-561c-11ea-8c78-335b843f9388.png" alt="repo image" width="400" height="80" />
   <h1 align="center">Mattermost Plugin Netlify</h1>
  <h5 align="center"><i>A two way integration plugin between Mattermost and Netlify.</i></h5>
</p>

## Table of content
- [Installation](#installation)
- [Setting up](#setting-up)
  * [Setup at Netlify](#setting-up-at-netlify)
  * [Configure settings](#setting-up-at-mattermost)
- [Running](#running-up)
- [Features](#features)
  * [Slash commands](#slash-commands)
      + [Connect](#connect-command)
      + [Disconnect](#disconnect-command)
      + [Help](#help-command)
      + [List](#list-command)
- [Hackathon](https://www.hackerearth.com/challenges/hackathon/mattermost-bot-hackfest)
   * [Ideation draft](https://github.com/M-ZubairAhmed/mattermost-plugin-netlify/blob/master/proposal.md)
   * [Repository submission](https://github.com/mattermost/mattermost-hackathon-hackerearth-jan2020/blob/master/hackathon-submissions/m-zubairahmed-mattermost-plugin-netlify.md)
   * [Pull request link](https://github.com/mattermost/mattermost-hackathon-hackerearth-jan2020/pull/3)

## Installation
Please download the latest version of the [release](https://github.com/M-ZubairAhmed/mattermost-plugin-netlify/releases) directory. Headover to `System Console` and drop the latest release in plugins section. For more help on how to install a custom plugin refers [installing custom plugin docs](https://docs.mattermost.com/administration/plugins.html#custom-plugins).

*Currently unstable due to active development, should be used for testing purpose only*.

## Setting up

### Setting up at Netlify

1. Head over to your Netlify account, and proceed to *User settings*. Find the tab *Applications* and under *OAuth applications* section; click on *New OAuth app* button. ![Screenshot_2020-02-23 Sites XOXOXO's team](https://user-images.githubusercontent.com/17708702/75109559-60b88580-561c-11ea-9a2d-a4e318251135.png)

1. Enter the **Application Name**. The **Redirect URI** should be of *<site url>/plugins/netlify/auth/redirect* form. The site url is where you mattermost app is hosted. Example if you are installing it at the community the redirect url can be https://community.mattermost.com/plugins/netlify/auth/redirect . Description is optional field. Hit save button when fields are filled. ![Screenshot_2020-02-23 OAuth applications Netlify](https://user-images.githubusercontent.com/17708702/75109197-1386e480-5619-11ea-823d-9c63eefb0fe3.png)

1. You can see the new oauth app of the *Application Name* you entered, created under *OAuth applications*. Expand it to view **Redirect URL**, **Client ID** and **Secret**. ![Screenshot_2020-02-23 Mohammed Zubair Ahmed Applications](https://user-images.githubusercontent.com/17708702/75109198-1a155c00-5619-11ea-87f2-4b7ab49dcd2a.png)

1. Keep all the information above handy to be copied at plugin settings. ie.
    - Application Name
    - Redirect URI
    - Client ID
    - Secret

### Setting up at Mattermost

1. Head back to your Mattermost application. And open *Sytem Console* and select *Netlify* under *Plugins* tab on the left.
![Screenshot_2020-02-23 System Console - Mattermost](https://user-images.githubusercontent.com/17708702/75110364-328b7380-5625-11ea-8d13-3e15fc77432d.png)

1. Copy over the fields from Netlify
    - **Netlify Application Name** also refered as **Application Name** at Netlify.
    - **Netlify Client ID** aslo refered as **Client ID** at Netlify
    - **Netlify Secret** also refered as **Secret** at Netlify
    - **Plugin Encryption Key** can be generated by hitting over *Regerate* button below it.
    
1. Hit *Save* button in the footer to save your settings.
1. Restart the plugin to propogate the effect. ![Screenshot_2020-02-23 System Console - Mattermostsas](https://user-images.githubusercontent.com/17708702/75110455-3d92d380-5626-11ea-9b63-37726d41ddae.png)

## Running up
1. This plugin works with the helps of simple commands and interactive dialoges. But first to be able to access any of the plugins commands, Netlify and Mattermost needs to be connected. Go to any channel and write the first slash command of netlify plugin `/netlify connect`. Netlify bot then posts a link through which you can authenticate your netlify account. ![Screenshot_2020-02-23 Any Channel - AQQQ Mattermost](https://user-images.githubusercontent.com/17708702/75111311-615b1700-5630-11ea-8489-7e3eadc3d844.png)

1. Follow the link and you will be taken to Netlify login site. After logging in you will be presented with a permission screen. Hit approve button. ![Screenshot_2020-02-23 Authorize Application Netlify](https://user-images.githubusercontent.com/17708702/75111404-4f2da880-5631-11ea-89e8-8f5d7db4efaf.png)

1. If successful, You will be redirected to successfully authenticated page. This page can be safely closed then.  ![Screenshot_2020-02-23 Screenshot](https://user-images.githubusercontent.com/17708702/75111435-b77c8a00-5631-11ea-871a-ddb7aeabaab0.png)

1. A new message from *netlify* Bot is also posted on the DM stating the same. With this Netlify Mattermost plugin is configured and ready to use. ![Screenshot_2020-02-23 netlify - AQQQ Mattermost](https://user-images.githubusercontent.com/17708702/75111440-c4997900-5631-11ea-9e18-f31c5f1502d2.png)

## Features

### Slash commands
With the series of slash commands at its disposal, Netlify plugin can be used to manage or change resources up at Netlify account.

### Connect command
`/netlify connect`

By executing this command, the bot will post a link into the channel. Folllowing which authentication with Netlify can be performed. Any of the belows commands requests that connection being made first before execution. After successfull authentication access token is stored in encrypted form at mattermost database. Which is then used to perform various operations via netlify api. 

![connect-gif](https://user-images.githubusercontent.com/17708702/75114678-1d780a00-5650-11ea-8e88-dc3369e4fa20.gif)

### Disconnect command
`/netlify disconnect`

This commands clears out authentication between netlify and mattermost. All the notifications are also unsubscribbed from Mattermost.

![disconnect-gif](https://user-images.githubusercontent.com/17708702/75114689-341e6100-5650-11ea-864c-c10614a22797.gif)

### Help command
`/netlify help`

It shows all the commands which are available for user to interact with Netlify bot.

### List command
`/netlify list`

It tabulates all the sites information of Netlify account. It lists name, url, custom domain, repository, deployed branch, managed by team, last updated of the site.

![list-gif](https://user-images.githubusercontent.com/17708702/75114697-41d3e680-5650-11ea-9684-c95a44a5f12d.gif)
