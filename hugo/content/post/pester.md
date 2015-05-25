+++
date = "2015-05-25T10:10:55-07:00"
draft = true
title = "Pester: Go (Golang) HTTP retries, backoff, and concurrency"

+++

### Resiliency in HTTP requests

Having worked for a few years on high scale and high availability systems, I've learned there are a few tricks that you can employ to make sure your requests make it back from over the network. There are two basic approaches in this arena: fail and recover fast (ala [Hystrix](https://github.com/Netflix/Hystrix)) and the bit older method of handling retires, backoff, and making concurrent requests. Hystrix seems pretty dang awesome, but may be a little heavy handed. I would like to introduce a small library wrapper I put together to handle the other side of the coin for retries, backoff, and concurrency: [Pester](https://github.com/sethgrid/pester).

### Retries

Some requests go out and hit a transitory error. Here is a recent example I've run into. We built an authorization system that creates a few hundred permission entries in our datastore when a user is created, adding to the millions of other records. Other services in our system can hit our authorization system on behalf of the new user nearly immediately, but due to replication lag, we may not have the data available just yet. More concretely, we could have a situation where `authorizationURL/{:user_id}` does not exist one moment and 404's, but the next moment will 200 and return data.

How can we deal with this? One option is to allow the service to just have an error and give a crap user experience. I'm not a fan of this approach. I'd wager we want to retry. Now we wrap our calls to the authorization service with special 404 handling to wait some amount of time and then try again. What if we've suffered a network partition? What if we could have multiple seconds (or worse, minutes) of down time for a given service, and we find we want to keep trying the downed endpoint? We can keep retrying but add some backoff to each request.

### Backoff

To help prevent us from knocking down the downstream service with requests (maybe it is responding so slowly due to increased load) we can introduce a backoff strategy. The strategy you use depends on your use case. It may make sense to wait one second between retries, or it may make sense to increase the backoffs from 1 second to 10 to several minutes or longer. Some of the systems I work on have retries that can happen hours later.

If we find ourselves making metric craptonnes of calls, we don't want to find all of our backoffs in sync. If our errors are due to load, we want to make sure our backoffs are staggered to avoid kicking the error can down the street. One solution to this problem is to add jitter to your backoffs by adding or subtracting some random amount of time to each request's scheduled backoff. This can work to ensure that we don't have a traveling herd of requests and errors.

### Concurrency

We've gone over ways to increase our system resiliency with retries and backoff, but what about getting our data back more quickly? If we are hitting our load balancer and it is routing calls for us, we may have calls that find their way to less busy servers. If we send out three requests, and they would come back at 200ms, 18ms, and 60ms. By making these three requests concurrently, we can get back our fastest 18ms response and continue about our day.

### Pester

Go's standard library is super solid. Building atop of it, I put together a simple library that has built-in retries, backoff, and concurrency. With Pester, I made sure that anyone who wanted to use it would not have to change any of their logic in their apps that deal with http client calls. Simply replace your http client calls with the pester version and you're set. Pester provides an http client that matches the standlib and supports `.Get`, `.Post`,`.PostForm`,`.Head`, and `.Do`. Each of which that can make automatic use of retries and backoff, and idempotent calls (aka GET) can make use of making concurrent requests. I hope your project can benefit from it!

```
{ // drop in replacement for http.Get and other client methods
    resp, err := pester.Get(SomeURL)
    if err != nil {
        log.Fatalf("error GETing default", err)
    }
    defer resp.Body.Close()

    log.Printf("GET: %s", resp.Status)
}


{ // control the resiliency
    client := pester.New()
    client.Concurrency = 3
    client.MaxRetries = 5
    client.Backoff = pester.ExponentialJitterBackoff
    client.KeepLog = true

    resp, err := client.Get(SomeURL)
    if err != nil {
        log.Fatalf("error GETing with all options, %s", client.LogString())
    }
    defer resp.Body.Close()

    log.Printf("Full Example: %s [request %d, retry %d]", resp.Status, client.SuccessReqNum, client.SuccessRetryNum)
}
```