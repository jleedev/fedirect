This is a replacement for `wikidata-externalid-url` for looking up addresses on Mastodon (Fediverse, Lemmy, etc.), which can't be done as a pure function but requires asking the server using a protocol called WebFinger. Per [arthurpsmith/wikidata-tools#30](https://github.com/arthurpsmith/wikidata-tools/issues/30), this requirement places this address resolution out of scope, so this is a separate server to do just that. This can be plugged in to [p:P4033](http://www.wikidata.org/entity/P4033) as a third-party formatter URL.

Why is this a requirement? Some specific examples:

The username `osi@opensource.org` would naively be resolved to `https://opensource.org/@osi`, which returns a 404; the correct URL is `https://social.opensource.org/@osi`. The URL for `macstories@macstories.net` is `https://mastodon.macstories.net/@macstories`. And so on.

Responses are cached for as long as the web server runs. Mastodon uses 3 days on its responses.

Syntax explanation:

- `?id=user@host` - Default lookup. A link with `"rel": "http[s]://webfinger.net/rel/profile-page"` will be used as the redirect, or else one with `"rel": "self"`. If there are multiple, the _last_ is used, as described by Lemmy and Mastodon.
- `?id=user@host&type=Person|Group` - A link having `"properties": { "https://www.w3.org/ns/activitystreams#type": "Group" }` will be used (case insensitive match), or an error will be returned.
- `?id=!user@host` - Internally converted to `type=Group`, as per Lemmy.
- `?id=@user@host` - Internally converted to `type=Person`, but will fall back to the default case if not found.

Everything should use the plain `user@host` request by default! The others are experiments to see if we can also handle [p:P11947](http://www.wikidata.org/entity/P11947) (Lemmy community ID). Kbin doesn't seem distinguish the type of the returned subject in the WebFinger response, only when you go and fetch the profile. I'm not interested in doing that, so Kbin will get an arbitrary redirect if multiple things share a username.
