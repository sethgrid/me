+++
date = "2015-03-01T10:50:48-08:00"
draft = false
title = "Hugo Syntax Highlighting"

+++

So, a few minutes ago, I was all excited on how easy Hugo was to get started. While I really like what I am seeing in a lot of ways, what I do not like is the documentation for syntax highlighting.

From the [configuration page](http://gohugo.io/overview/configuration/):

```
# color-codes for highlighting derived from this style
pygmentsStyle:              "monokai"
# true: use pygments-css or false: color-codes directly
pygmentsUseClasses:         false
```

Using this, coupled with the [syntax highlighting page](http://gohugo.io/extras/highlighting/), I cannot get sytanx highlighting to work the way you would think it should.

Using defaults, I can do something like:

```
    // I have the `{` escaped because I can't figure
    // out how to show the handlebars in this view:(
    \{\{< highlight go >}}
    package main
    ...
    \{\{< /highlight >}}
```

And. as long as `pygmentsUseClasses` is defaulted to false, I am solid. However, I don't like the default black background. I figured it should be trivial to update to a different style. This is where things fell apart.

If I change `pygmentsStyle` to _anything_ other than `"monokai"`, the code block changes to just a `<p>` tag. If I change `pygmentsUseClasses` to `true`, I get no syntax highlighting. Period.

I figured this could be easily overcome. I went into the css that was I though should be loading. There is `public/css/syntax.css` and I should be able to just overwrite the css with the values I want. Apparently, I have no clue what I am doing. Adjusting the colors in that file would not change the way code looked on the screen.

At this point, I have other things to do. Like write this post. I have a lunch/dinner guests coming over, and what should have been very straight forward has turned out to be a time sink. All I wanted was pretty syntax highlighting.