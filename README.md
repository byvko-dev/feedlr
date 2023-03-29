# Feedlr

Feedlr is a simple Discord Bot + Backend Service that allows you to add RSS Feeds to your Discord Server.

### Setup

1. Clone the repository
2. Install Taskfile (https://taskfile.dev/#/installation)
3. Copy `.env.example` to `.env` and fill in the values

### Development
1. Run `task prisma` to build the Prisma Client and push the Prisma Schema to your database
2. Run `task bot-dev` to start the Discord Bot in development mode
3. Run `task server-dev` to start the Backend Service in development mode

You can also run `task dev` to start both the Discord Bot and the Backend Service in development mode.

### Deployment
This project is designed to be deployed to Railway (https://railway.app/).
*Note: Deploying MongoDB in Railway did not work with Prisma (It seems to require a multi-node cluster), though I gave it very little effort. Feel free to submit a PR if you would like to fix this.*

Pre-requisites:
- Discord Bot Token
- Railway Account
- MongoDB Deployment
- Manager/Self hosted AMQP Server

Deployment Steps:
1. Create a new project in Railway
2. Configure shared environment variables, include all variables from `.env.example`  
3. **For each service** - bot, worker, scheduler:  
    1. Create a new Service in the project and point it to this repository
    2. Add all shared environment variables to the Service in Settings
    3. If the option is available, set the config path to `railway.[service name].toml`, e.g. `railway.bot.toml`
        - If the option is not available, you will need to manually change the Settings to match the config file
    4. Deploy the Service
