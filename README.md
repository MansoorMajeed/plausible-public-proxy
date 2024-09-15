# Plausible Proxy with no authentication

## What

This is a proxy for Plausible Stats API, which currently only returns page views for a page

## Why

I use a hugo static site generator for my blog and I wanted to show each post's views.
But Plausible API does not support narrowed scopes, so I can't directly call the API using the JavaScript
code from the browser.

So, this proxy is a solution for that. It can be accessed without any authentication and will show the the pageviews
for the configured site.

## How to use it

```
export PLAUSIBLE_API_KEY=<plausible api token>
export PLAUSIBLE_SITE_ID=M<example.com>
export PLAUSIBLE_URL=https://plausible.domain.com
export PORT=1200

go build

./plausible-proxy

curl 'localhost:1200/pageviews?page=/blog'
{"pageviews":231,"page":"/blog"}

```


