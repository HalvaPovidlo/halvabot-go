// Package docs GENERATED BY SWAG; DO NOT EDIT
// This file was generated by swaggo/swag
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "license": {
            "name": "Apache 2.0",
            "url": "http://www.apache.org/licenses/LICENSE-2.0.html"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/login": {
            "post": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Validates your login and password. Returns JWT.",
                "parameters": [
                    {
                        "description": "Login and password",
                        "name": "query",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/login.Credentials"
                        }
                    }
                ],
                "responses": {}
            }
        },
        "/music/enqueue": {
            "post": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "music",
                    "JWTAuth"
                ],
                "summary": "Play the song from YouTube by name or url",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Bearer",
                        "name": "Authorization",
                        "in": "header",
                        "required": true
                    },
                    {
                        "description": "Song name or url",
                        "name": "query",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/rest.songQuery"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "The song that was added to the queue",
                        "schema": {
                            "$ref": "#/definitions/rest.EnqueueResponse"
                        }
                    },
                    "400": {
                        "description": "Incorrect input",
                        "schema": {
                            "$ref": "#/definitions/rest.Response"
                        }
                    },
                    "500": {
                        "description": "Internal error. This does not necessarily mean that the song will not play. For example, if there is a database error, the song will still be added to the queue.",
                        "schema": {
                            "$ref": "#/definitions/rest.Response"
                        }
                    }
                }
            }
        },
        "/music/loopstatus": {
            "get": {
                "produces": [
                    "text/plain"
                ],
                "tags": [
                    "music"
                ],
                "summary": "Is loop mode enabled",
                "responses": {
                    "200": {
                        "description": "Returns true or false as string",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/music/now": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "music"
                ],
                "summary": "Song that is playing now",
                "responses": {
                    "200": {
                        "description": "The song that is playing right now",
                        "schema": {
                            "$ref": "#/definitions/pkg.Song"
                        }
                    }
                }
            }
        },
        "/music/radiostatus": {
            "get": {
                "produces": [
                    "text/plain"
                ],
                "tags": [
                    "music"
                ],
                "summary": "Is radio mode enabled",
                "responses": {
                    "200": {
                        "description": "Returns true or false as string",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/music/setloop": {
            "post": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "music",
                    "JWTAuth"
                ],
                "summary": "Set loop mode",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Bearer",
                        "name": "Authorization",
                        "in": "header",
                        "required": true
                    },
                    {
                        "description": "Send true to enable and false to disable",
                        "name": "query",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/rest.enableQuery"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Incorrect input",
                        "schema": {
                            "$ref": "#/definitions/rest.Response"
                        }
                    }
                }
            }
        },
        "/music/setradio": {
            "post": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "music",
                    "JWTAuth"
                ],
                "summary": "Set radio mode",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Bearer",
                        "name": "Authorization",
                        "in": "header",
                        "required": true
                    },
                    {
                        "description": "Send true to enable and false to disable",
                        "name": "query",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/rest.enableQuery"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Incorrect input",
                        "schema": {
                            "$ref": "#/definitions/rest.Response"
                        }
                    }
                }
            }
        },
        "/music/skip": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "music",
                    "JWTAuth"
                ],
                "summary": "Skip the current song and play next from the queue",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Bearer",
                        "name": "Authorization",
                        "in": "header",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/music/songstatus": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "music"
                ],
                "summary": "SongStatus of player on the current song",
                "responses": {
                    "200": {
                        "description": "The song that is playing right now",
                        "schema": {
                            "$ref": "#/definitions/pkg.SessionStats"
                        }
                    }
                }
            }
        },
        "/music/status": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "music"
                ],
                "summary": "Status of the player",
                "responses": {
                    "200": {
                        "description": "Status of the player",
                        "schema": {
                            "$ref": "#/definitions/pkg.PlayerStatus"
                        }
                    }
                }
            }
        },
        "/music/stop": {
            "get": {
                "produces": [
                    "text/plain"
                ],
                "tags": [
                    "music",
                    "JWTAuth"
                ],
                "summary": "Skip the current song and play next from the queue",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Bearer",
                        "name": "Authorization",
                        "in": "header",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "login.Credentials": {
            "type": "object",
            "required": [
                "login",
                "password"
            ],
            "properties": {
                "login": {
                    "type": "string"
                },
                "password": {
                    "type": "string"
                }
            }
        },
        "pkg.PlayDate": {
            "type": "object",
            "properties": {
                "time.Time": {
                    "type": "string"
                }
            }
        },
        "pkg.PlayerStatus": {
            "type": "object",
            "properties": {
                "loop": {
                    "type": "boolean"
                },
                "now": {
                    "$ref": "#/definitions/pkg.Song"
                },
                "radio": {
                    "type": "boolean"
                },
                "song": {
                    "$ref": "#/definitions/pkg.SessionStats"
                }
            }
        },
        "pkg.SessionStats": {
            "type": "object",
            "properties": {
                "duration": {
                    "description": "seconds",
                    "type": "number"
                },
                "position": {
                    "description": "seconds",
                    "type": "number"
                }
            }
        },
        "pkg.Song": {
            "type": "object",
            "properties": {
                "artist_name": {
                    "type": "string"
                },
                "artist_url": {
                    "type": "string"
                },
                "artwork_url": {
                    "type": "string"
                },
                "last_play": {
                    "$ref": "#/definitions/pkg.PlayDate"
                },
                "playbacks": {
                    "type": "integer"
                },
                "service": {
                    "type": "string"
                },
                "thumbnail_url": {
                    "type": "string"
                },
                "title": {
                    "type": "string"
                },
                "url": {
                    "type": "string"
                }
            }
        },
        "rest.EnqueueResponse": {
            "type": "object",
            "properties": {
                "playbacks_count": {
                    "type": "integer"
                },
                "song": {
                    "$ref": "#/definitions/pkg.Song"
                }
            }
        },
        "rest.Response": {
            "type": "object",
            "properties": {
                "message": {
                    "type": "string"
                }
            }
        },
        "rest.enableQuery": {
            "type": "object",
            "properties": {
                "enable": {
                    "type": "boolean"
                }
            }
        },
        "rest.songQuery": {
            "type": "object",
            "required": [
                "song"
            ],
            "properties": {
                "song": {
                    "type": "string"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "localhost:9091",
	BasePath:         "/api/v1",
	Schemes:          []string{},
	Title:            "HalvaBot for Discord",
	Description:      "A music discord bot.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
