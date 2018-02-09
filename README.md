## sethammons.com

This is the hugo project running sethammons.com. It is an excuse to try some docker things and play with Hugo.

If you have docker:
```
$ docker run --rm -it -v $PWD:/src -p 1313:1313 -u hugo jguyomard/hugo-builder hugo server -w --bind=0.0.0.0 --theme temple
```

To create a new post:
```
docker run --rm -it -v $PWD:/src -p 1313:1313 -u hugo jguyomard/hugo-builder hugo new posts/my-post.md
```

A handy alias:
```
alias hugo='docker run --rm -it -v $PWD:/src -u hugo jguyomard/hugo-builder hugo'
```
