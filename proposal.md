<p align="center">
        <img src="https://user-images.githubusercontent.com/17708702/73535585-ce0d3800-441b-11ea-8187-1ea83a9cce32.png" alt="repo image" width="400" height="80" />
   <h1 align="center">Mattermost Netlify Bot</h1>
  <h5 align="center"><i>Proposal for Mattermost Netlify Bot for Mattermost Hackathon</i></h5>
    
</p>


Competition : [Mattermost Bot Hackfest Jan 2020](https://www.hackerearth.com/challenges/hackathon/mattermost-bot-hackfest)

Team : [TWR](https://www.hackerearth.com/challenges/hackathon/mattermost-bot-hackfest/dashboard/2abe565/team/)

Team Members : [@M-ZubairAhmed](https://github.com/M-ZubairAhmed)

Theme : ChatOps

Proposal : [Hackerearth's idea section](https://www.hackerearth.com/challenges/hackathon/mattermost-bot-hackfest/dashboard/2abe565/idea/)

Submission : [TBA](https://www.hackerearth.com/challenges/hackathon/mattermost-bot-hackfest/dashboard/2abe565/submission/)

![blue-line](https://i.imgur.com/cETzBqq.png)

## Summary :page_facing_up:
Mattermost Netlify bot is an intermediary agent between your netlify and mattermost account. It makes it easy to monitor and interact with your netlify's resources all within your team's channel. Once integrated with your mattermost channel, team can start recieving various Netlify notifications such as netlify form submissions, build failures etc and can run commands to redeploy, see build stats, create hooks and much more.

## Problem Statement :rotating_light:

- Familiar interface : System admins can manage netlify configuration right from the chat window with which they are familiar with.
- Critical notification to team : Concerned teams are notified of the issue which makes it easier to plan and execute the solution rapidly.
- System health and monitoring on the fly.

## Features :sparkles:

All commands start with prefix *netlify*

``` txt
/netlify command-name
```

#### :pencil2: Command : `/connect`
Connects your Mattermost account with your Netlif's account

#### :pencil2: Command : `/disconnect`
Disconnects your Netlif's account from Mattermost

#### :pencil2: Command : `/subscribe`
Manages webhook notifications of Netlify's sites to post in Mattermost channels.

#### :pencil2: Command : `/build`
Triggers build for Netlify site

#### :pencil2: Command : `/logs`
Gets build logs of the site.

#### :pencil2: Command : `/notifications`
Create and manage notifications for events such as deploy-started, deploy-failed etc for Netlify's site

#### :pencil2: Command : `/site`
Manage basic site settings such as SSL, DNS, processing settings etc.

#### :bell: Deploy Notifications
Get notified when build is started, fails or succeeds.

#### :bell: Form Notifications
Netlify form enabled sites can inform when there are new form submissions

## Development approach :wrench:
Bot will be developed on Mattermost platform via [Mattermost Plugin](https://developers.mattermost.com/extend/plugins/). API's of Netlify will be integrated by [Netlify Go API Client](https://github.com/netlify/open-api#go-client)
