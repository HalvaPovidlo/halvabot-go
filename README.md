# halvabot-go
[![Linter](https://github.com/HalvaPovidlo/halvabot-go/actions/workflows/linter.yml/badge.svg)](https://github.com/HalvaPovidlo/discordBotGo/actions/workflows/linter.yml) [![Test](https://github.com/HalvaPovidlo/halvabot-go/actions/workflows/test.yml/badge.svg)](https://github.com/HalvaPovidlo/halvabot-go/actions/workflows/test.yml)

## Enable Discord

You need to sign up at [Discord Developer Portal](https://discord.com/developers/applications)

And join our HPDevelopment team.

Create file `secret_config.json`

```json

{
  "general":{
    "debug":true
  },
  "host":{
    "ip": "***",
    "bot": "***",
    "mock": "***",
    "web": "***"
  },
  "discord":{
    "token":"***",
    "bot":"HalvaBot",
    "id":746726055259406426,
    "prefix":"$",
    "api": {
      "open": ["основной", "видосы", "плейлисты"],
      "status": ["music", "debug"]
    }
  },
  "youtube":{
    "download":false,
    "output":""
  }
}
```
Replace `token` with token from Discord Developer Portal.

Applications -> HalvaBot -> Bot -> Click to reveal token

**Don't pass this token on to anyone!!!**
