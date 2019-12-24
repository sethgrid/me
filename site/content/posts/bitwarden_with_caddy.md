---
date: 2019-12-22T18:02:55-07:00
draft: false
title: "Bitwarden with Caddy 2 on Digital Ocean"
---

### The Preamble Ramble

I finally had to bite the bullet and go away from the antiquated way I was leveraging password management. I was maintaining a keepass database synced via Dropbox. This let me keep my phone, my work laptop, my desktop, and my wife's computer all in sync. However, when my wife re-installed her OS last week, Dropbox informed me that my license is only good for three computers. Well, time to finally make the move to Bitwarden.

I've heard good things about Bitwarden and I especially like the idea of self-hosting my password management solution. Having easy browser integration and mobile clients made looking into getting it running on my Digital Ocean box a no-brainer.

Unfortunately, as I was getting ready to sign into my Digital Ocean box, I realized something was wrong. I couldn't ssh, I couldn't open a web console. Nothing. Digital Oceans support is lack luster for emergencies like "I can't reach my node and it is DO's fault." Their self-service mentality assumes everything that can go wrong is the user's fault. Fast forward a bit, and I had to re-create my instance and start from scratch. Since I was starting from scratch, I decided I would give Caddy server a try too. And since Caddy 2 is in beta, I might as well use that!

## Caddy Configuration

Caddy 2 is functional, but still very beta. Docs are getting written and, for someone who is not familiar with it, there is a bit of a learning curve. While there is apparently some magic simple config format, the "real" config format is JSON, and if you want to have full control of your Caddy 2 instance, you are going to be writing some JSON.

Previously, I thought YAML was a pain in the butt. Man. Hand crafting JSON is terrible. That said, here is a Caddy configuration that allowed me to have logging, to run my normal site as a static file server, also run my son's site as a reverse proxy to a Go binary running on my server, and a reverse proxy to my Bitwarden install:

```json
{
    "logging": {
        "sink": {
            "writer": {
                "output": "file",
                "filename": "/var/log/caddy/sink.log"
            }
        },
        "logs": {
            "default": {
                "writer": {
                    "output": "file",
                    "filename": "/var/log/caddy/caddy.log"
                },
                "encoder": {
                    "format":"console"
                },
                "level": "info",
                "include": [],
                "exclude": []
            }
        }
    },
    "apps": {
        "http": {
            "servers": {
                "myserver": {
                    "listen": [
                        ":443"
                    ],
                    "routes": [
                         {
                            "match": [
                                {
                                    "host": [
                                        "bitwarden.sethammons.com"
                                    ]
                                }
                            ],
                            "handle": [
                                {
                                    "handler": "reverse_proxy",
                                    "transport": {
                                      "protocol": "http",
                                      "tls": {}
                                    },
                                    "upstreams": [
                                        {
                                            "dial": "bitwarden.sethammons.com:5443",
                                            "max_requests": 1000
                                        }
                                    ]
                                }
                            ]
                        },

                        {
                            "match": [
                                {
                                    "host": [
                                        "sethammons.com"
                                    ]
                                }
                            ],
                            "handle": [
                                {
                                    "handler": "file_server",
                                    "root": "/home/seth/projects/me/site/public"
                                }
                            ]
                        },
                        {
                            "match": [
                                {
                                    "host": [
                                        "grzlybr.com"
                                    ]
                                }
                            ],
                            "handle": [
                                {
                                    "handler": "reverse_proxy",
                                    "upstreams": [
                                        {
                                            "dial": "localhost:1414",
                                            "max_requests": 1000
                                        }
                                    ]
                                }
                            ]
                        }
                    ]
                }
            }
        }
    }
}
```

Whew. There is a bit going on there. The relevant section for Bitwarden is the handler:
```json
{
    "handler": "reverse_proxy",
    "transport": {
      "protocol": "http",
      "tls": {}
    },
    "upstreams": [
        {
            "dial": "bitwarden.sethammons.com:5443",
            "max_requests": 1000
        }
    ]
}
```

This shows us setting up a reverse proxy (ie, just forward the requests through Caddy to some other address). By specifying the transport, we are able to send the HTTPs requests stright through to let Bitwarden's secure address. To use this config, I run `caddy start --config caddy.conf`.


## Bitwarden

This took a bit more back and forth to get going. I think there is a better way to do this, but for the time being, this works. We install Bitwarden as normal, but we update the SSL port from the default `:443` to, in this case, `:5443`. Let Bitwarden set up its Let's Encrypt Certificate like normal. Note, Bitwarden tries to validate everything will work and will try to bind to `:443` during the install. This means that Caddy will have to be temporarily stopped so Bitwarden can do its thing.

After Bitwarden is installed, we can alter some of its configs. Start by updating `bwdata/env/global.override.env`. You will need to update all the URL paths to include your custom SSL port. There are multiple entries to update, but the basic gist is to add the port like so:
```
globalSettings__baseServiceUri__api=https://bitwarden.sethammons.com:5443/api
```

You will also need to tell Bitwarden that you have a custom SSL port. Edit  `/bwdata/config.yml` to have `https_port: 5443`.

Run `./bitwarden.sh rebuild` and `./bitwarden.sh restart`. Since Caddy was turned off, you can start it back up. I do so by pointing at my on disk config: `caddy start --config caddy.conf`.

# Validating your Installation

Go to your Bitwarden url you set up and bask in the glory of your self hosted Bitwarden instance. Create a new user and go to town setting up passwords and the like.

# Post Install - Prevent Bitwarden Sign Ups

Now that everything is working, I did not want to allow additional signups and needed to disable that. Simply update `bwdata/env/global.override.env`:
```
globalSettings__disableUserRegistration=true
```
And run `./bitwarden.sh updateconf`. You may need to restart the service with `./bitwarden.sh restart`.

## Todo

I'm pretty sure that this current set up is going to require me to stop the Caddy server every few months and update the Let's Encrypt Certificate that Bitwarden is using. That sucks. I need to dig in and likely point Bitwarden at the certificate used by Caddy. It would be great if Bitwarden provided a "I'm behind an SSL equiped load balancer, I don't need SSL" option, but that does not exist to my knowledge.
