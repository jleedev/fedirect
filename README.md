This is a replacement for `wikidata-externalid-url` for looking up addresses on Mastodon (Fediverse, Lemmy, etc.), which can't be done as a pure function but requires asking the server using a protocol called WebFinger. Per [arthurpsmith/wikidata-tools#30](https://github.com/arthurpsmith/wikidata-tools/issues/30), this requirement places this address resolution out of scope, so this is a separate server to do just that. This can be plugged in to [p:P4033](http://www.wikidata.org/entity/P4033) as a third-party formatter URL.

Why is this a requirement? Some specific examples:

The username `osi@opensource.org` would naively be resolved to `https://opensource.org/@osi`, which returns a 404; the correct URL is `https://social.opensource.org/@osi`. The URL for `macstories@macstories.net` is `https://mastodon.macstories.net/@macstories`. And so on.

Responses are cached for as long as the web server runs. Mastodon uses 3 days on its responses.
