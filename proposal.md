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
Mattermost Netlify bot is an intermediary agent between your Netlify and mattermost account. It makes it easy to monitor and interact with your Netlify's resources all within your team's channel. Once integrated with your Mattermost channel, team can start recieving various Netlify notifications such as Netlify form submissions, build failures etc and can run commands to redeploy, see build stats, create hooks and much more.

## Problem Statement :rotating_light:

- Familiar interface : System admins can manage Netlify configuration right from the chat window with which they are familiar with.
- Critical notification to team : Concerned teams are notified of the issue which makes it easier to plan and execute the solution rapidly.
- System health and monitoring on the fly.

## Features :sparkles:

All commands start with prefix *netlify*

``` txt
/netlify command-name
```

#### :pencil2: Command : `/connect`
Connects your Netlify's account into Mattermost.
![Screenshot from 2020-02-01 07-35-02](https://user-images.githubusercontent.com/17708702/73585246-4fee7700-4497-11ea-862f-4baec768d00b.png)

#### :pencil2: Command : `/disconnect`
Disconnects your Netlify's account from Mattermost.
![Screenshot from 2020-02-01 07-42-18](https://user-images.githubusercontent.com/17708702/73585336-4c0f2480-4498-11ea-92b7-8735763f41d2.png)

#### :pencil2: Command : `/notifications`
Create and manage notifications for events such as deploy-started, deploy-failed and more for your site.
![Screenshot from 2020-02-01 07-58-56](https://user-images.githubusercontent.com/17708702/73585601-9ee9db80-449a-11ea-95ae-19319992829c.png)

#### :pencil2: Command : `/subscribe`
Manages Mattermost channels which are subscribed to recieve notifications from Netlify.
![Screenshot from 2020-02-01 08-03-59](https://user-images.githubusercontent.com/17708702/73585670-52eb6680-449b-11ea-9e91-54f78e3e1733.png)

#### :pencil2: Command : `/build`
Triggers build for sites.
![Screenshot from 2020-02-01 08-05-17](https://user-images.githubusercontent.com/17708702/73585695-8201d800-449b-11ea-9ae3-29bf275aa1c8.png)

#### :pencil2: Command : `/site`
Manage site settings such as SSL, DNS, processing settings etc.
![Screenshot from 2020-02-01 08-25-45](https://user-images.githubusercontent.com/17708702/73585941-68ae5b00-449e-11ea-8b63-4b2c863b763a.png)

#### :bell: Deploy Notifications
Get notified when build is starts, fails or succeeds.
![Screenshot from 2020-02-01 08-27-51](https://user-images.githubusercontent.com/17708702/73585948-a8754280-449e-11ea-8fe1-882f1f574a30.png)

#### :bell: Form Notifications
Netlify form enabled sites can notify when there are new form submissions.
![Screenshot from 2020-02-01 08-28-23](https://user-images.githubusercontent.com/17708702/73585952-c17df380-449e-11ea-9b05-6799f53971a8.png)

## Development approach :wrench:
Bot will be developed on Mattermost platform via [Mattermost Plugin](https://developers.mattermost.com/extend/plugins/). API's of Netlify will be integrated by [Netlify Go API Client](https://github.com/netlify/open-api#go-client)
