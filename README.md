<p align="center">
        <img src="https://user-images.githubusercontent.com/17708702/75109618-f18f6100-561c-11ea-8c78-335b843f9388.png" alt="repo image" width="400" height="80" />
   <h1 align="center">Mattermost Plugin Netlify</h1>
  <h5 align="center"><i>A two way integration plugin between Mattermost and Netlify.</i></h5>
</p>

![blue-line](https://user-images.githubusercontent.com/17708702/75109655-3b784700-561d-11ea-8fef-da11ec7dbc2e.png)

Note : *Currently unstable due to active development, to be used for testing purpose only*.

> Started as a part of [Mattermost Bot Hackathon Jan/Feb '20](https://www.hackerearth.com/challenges/hackathon/mattermost-bot-hackfest/custom-tab/submission-guideline/#Submission%20Guideline), Ideation draft is available [here](https://github.com/M-ZubairAhmed/mattermost-plugin-netlify/blob/master/proposal.md).


## Installing plugin
Please download the latest version of the [release](https://github.com/M-ZubairAhmed/mattermost-plugin-netlify/releases) directory. Headover to `System Console` and drop the latest release in plugins section. For more help on how to install a custom plugin refers [installing custom plugin docs](https://docs.mattermost.com/administration/plugins.html#custom-plugins).

## Setting up
### Setting up at Netlify
1. Head over to your Netlify account, and proceed to *User settings*. Find the tab *Applications* and under *OAuth applications* section; click on *New OAuth app* button. ![Screenshot_2020-02-23 Sites XOXOXO's team](https://user-images.githubusercontent.com/17708702/75109559-60b88580-561c-11ea-9a2d-a4e318251135.png)

1. Enter the **Application Name**. The **Redirect URI** should be of *<site url>/plugins/netlify/auth/redirect* form. The site url is where you mattermost app is hosted. Example if you are installing it at the community the redirect url can be https://community.mattermost.com/plugins/netlify/auth/redirect . Description is optional field. Hit save button when fields are filled. ![Screenshot_2020-02-23 OAuth applications Netlify](https://user-images.githubusercontent.com/17708702/75109197-1386e480-5619-11ea-823d-9c63eefb0fe3.png)

1. You can see the new oauth app of the *Application Name* you entered, created under *OAuth applications*. Expand it to view **Redirect URL**, **Client ID** and **Secret**. ![Screenshot_2020-02-23 Mohammed Zubair Ahmed Applications](https://user-images.githubusercontent.com/17708702/75109198-1a155c00-5619-11ea-87f2-4b7ab49dcd2a.png)

1. Keep all the information above handy to be copied at plugin settings. ie.
    - Application Name
    - Redirect URI
    - Client ID
    - Secret (a.k.a Client Secret)

### Setting up at Mattermost
