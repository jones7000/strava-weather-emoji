# strava weather
Adds weather info to strava activity, emoji to title and temperature to description

* using ngrok to access local container

### how to build
1) Add env variables for ngrok
```
export NGROK_AUTHTOKEN=<TOKEN>
export NGROK_URL=<URL>
```
2) Build container
```
docker-compose up --build
```

# strava api
* [Getting Started Auth](https://developers.strava.com/docs/getting-started/#account)
* [Swagger Playground](https://developers.strava.com/playground/)
* [Developer API Description and Datatypes](https://developers.strava.com/docs/reference/#api-Routes-getRouteById)

## strava webhook

### test
```sh
curl -X POST <URL> \
  -H 'Content-Type: application/json' \
  -d '{
      "aspect_type": "create",
      "object_id": 9999999,
      "object_type": "activity",
      "owner_id": 9999999
    }'

```

### show
```bash
curl -G https://www.strava.com/api/v3/push_subscriptions \
-d client_id=<ID> \
-d client_secret=<SECRET> 
```
### delete
```sh
curl -X DELETE https://www.strava.com/api/v3/push_subscriptions/[WebhookID]\?client_id=<ID>\&client_secret=<SECRET>
```
## add
```sh
curl -X POST https://www.strava.com/api/v3/push_subscriptions \ 
-d client_id=<ID> \
-d client_secret=<SECRET> \
-d callback_url=<URL> \
-d verify_token=<TOKEN>
```

# TODO
- [ ] ...