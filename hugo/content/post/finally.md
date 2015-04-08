+++
date = "2015-04-07T18:02:55-07:00"
draft = true
title = "Finally, now we can get started"

+++

### Posting new content under Docker

Time is so hard to find. Getting pulled three ways from Sunday at work has left me outside of a solid work-life balance. As such, it I have not had the opportunity to get back to playing with this setup.

Well, the time has come. I finally got everything figured out between Digital Ocean, Hugo, and Docker.

## The Set Up

Check out [this repo](http://www.github.com/sethgrid/me) for the complete digs; there you will find everything. Here are the basics. I used `fig` mostly "just because". I wanted to see how it orchestrated docker containers together. I did not get too deep into that, but, hey, there is a redis server running on this box. It does not do anything, but it is there!

## The Problem

After getting everything basically set up, I ran into an issue where the Hugo theme was not being applied. Looking at the source of the generated pages, it became quickly evident what was going on: all the links to resources were to the localhost. For whatever reason, the Hugo config was not picking up the `baseURL` config value.

The way I got around that was to use one of Hugo's flags, `baseURL`, as seen in my `fig.yml`:

```
  command: hugo server --theme=nofancy --buildDrafts --baseUrl=http://104.131.9.167:1313
```

## Finishing Up

After scp'ing my repo up to my Digital Ocean instance, I installed fig (Docker was already there). To handle the need for `sudo fig ...`, I created a symlink for `/usr/bin/fig` to `/usr/sbin/fig`. All that was needed was to `fig build && fig up`. Presto, done.

## Todo

I want to get the container size down. It is sitting at about 350mb. I'll be looking into tips and tricks to see what I can do in that regard. Additionally, I'm creating the docker image locally on the Digital Instance. I'd like to generate it on my laptop and scp it up. We will see.