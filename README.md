# Subsonic Sanitizer Proxy

A proxy that intercepts and reduces the size of `getPlaylist` and `getPlaylists` responses from a SubSonic API, primarily to workaround an issue syncing the [SubMusic app](https://apps.garmin.com/en-US/apps/600bd75f-6ccf-4ca5-bc7a-0a4fcfdcf794) on Garmin Watches.

## Why

There is a limit to the response size of network requests that Garmin watches can/will process (that limit varies by model and device memory). This prevents the SubMusic app from syncing playslists larger than a handful of songs, because the SubSonic API returns responses that are too large  resulting in `402 NETWORK_RESPONSE_TOO_LARGE` errors.

## How

The SubSonic API returns a lot more information than SubMusic consumes. This is a simple reverse proxy server that intercepts playlist requests and strips out the JSON data that is not used by SubMusic, significantly reducing the size responses (by ~80%). This allows syncing larger playlists to your watch without issue.

## Installation/usage



### Docker

`joekrill/subsonic-sanitizer`

Set `TARGET_URL` environment variable to your SubSonic instance (i.e. "http://mynavidrom:5433")

### Docker Compose

See [compose-example.yml](./compose-example.yml) for an example of proxying requests to Navidrome

## Links

- [SubMusic Garmin App](https://apps.garmin.com/en-US/apps/600bd75f-6ccf-4ca5-bc7a-0a4fcfdcf794)
- [SubMusic Github Repo](https://github.com/memen45/SubMusic)