+++
date = "2015-03-01T08:27:46-08:00"
draft = false
title = "Moving to Hugo"

+++

### Now on Hugo

I've been looking lately to move off my shared hosting (I know, right?) and I have been seeking an excuse to move all my stuff over to Digital Ocean. I've finally gotten off the pot, and decided that I would expore Docker using Digital Ocean as the backing hardware.

Previously, I had a homegrown site. As I don't spend my freetime making hot designs anymore, that site quickly fell out of anything resembling something polished. I switched over to the first blog-like management utility I could quickly set up on the shared hosting. It could not handle code snippets and that limited my drive to keep posting. Being so busy, I just let it slide.

Fast foward to today, and I finally decided it was time to get my web-presence back up-to-snuff. Knowing that I was to play with Docker and Digital Ocean, coupled with knowing that I wanted something easily able to write small posts about and show code snippits (and the fact that I am thuroughly enjoying Go), it seemed like Hugo was a great choice.

So, here it is. My first post in Hugo.

{{< highlight go >}}
package main

import "log"

func main(){
        log.Println("Digital Ocean + Docker + Hugo = Posting Again")
}
{{< /highlight >}}