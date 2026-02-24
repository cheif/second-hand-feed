# Second Hand - feed

This is an utility that allows you to create an atom feed (consumable from most rss/feed readers) based on one or more searches on second hand sites.

## Setup

This is a very simple app, it keeps it's state in a single `.json` file (the only argument when running), and it can be launched by just building the `Dockerfile`, and then running it while exposing port `:8080`.

### Usage

It currently supports these sites:
- https://www.blocket.se
- https://www.vinted.se (Should most probably work for other localized sites as well, but I haven't tried).

Just go to the site, and search for what you want, using builtin filters should also work. When you're happy, just copy the URL and paste it into the UI on `:8080`, the query should pop up in the list.

Now you should point your feed-reader towards the same url (you might have to append `/atom.xml`, but hopefully not), and it should discover a the feed, and you should see the same item's that you got from your search. How/when the feeds are updated are a bit of black magick to me (and probably reader dependent), but they seem to update reasonably often for me at least.

### Credits

This is heavily inspired by [Vinted-Notifications](https://github.com/Fuyucch1/Vinted-Notifications/tree/main), which does a similar thing, but with more bells and whistles, and only supports Vinted.
