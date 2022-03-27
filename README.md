# DiscordBotGo

## Enable Discord

You need to sign up at [Discord Developer Portal](https://discord.com/developers/applications)

And join our HPDevelopment team.

Create file `secret_config.json`

```json
{
  "discord": {
    "token": "***",
    "bot": "HalvaBot",
    "id": 746726055259406426,
    "prefix": "$"
  },
  "sheets": {
    "id": "***",
    "film": "***"
  },
  "vk": {
    "login": "***",
    "password": "***"
  },
  "lichess" : {
    "token": "***"
  },
  "general" : {
    "debug": true
  }
}
```
Replace `token` with token from Discord Developer Portal.

Applications -> HalvaBot -> Bot -> Click to reveal token

**Don't pass this token on to anyone!!!**
