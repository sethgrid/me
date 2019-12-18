+++
date = "2013-04-28T18:02:55-07:00"
draft = false 
title = "Settings Backup"

+++

### Settings Backup and Restoration with GitHub

A new Ubuntu version is coming out and the last thing I want to do is manually copy over my myriad of different config and settings files. I have my PS1 function set the way I like it, my Terminator terminal all set up with my keybindings and colors, some handy aliases, and some other rc_goodies. Thanks to the awesome power of GitHub, I fret of this aspect of upgrading no longer. I set up a repo that has a script that pulls in all my latest and greatest changes and can push those changes back out to my system.

### The Backup/Restore Process

I have a python script that has a list of file names. For each file, it `shutil.copyfile()`’s the file to my current working directory under `files/filename`. There is one caveat. I change the filename from `/path/to/file` to `slash__path__slash__to__slash__file`. Upon restoration, I run the same script with a different flag for restoration. It changes those `__slash` tokens back to a good ol’ forward-slash and copies the file back into the place it belongs. Prior to restoring the file, the script checks to see if the file already exists, and if it does, replace it with filename.BAK.

### On a New System

All I have to do is git clone my repo, and run `grab_settings.py restore`. Vim works the way I like. Terminator works the way I like. My aliases are all present. Done and done.
