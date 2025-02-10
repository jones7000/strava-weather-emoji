# Auto Strava Weather Emoji

### Image bauen und starten

```
docker-compose up --build # build sorgt dafür, dass das Image neu gebaut wird.
```

```

* Manuell starten mit:
```
docker run -d -p 8080:8080 --name go-app mein-go-app:latest
```


* Image nur bauen:
```
docker-compose build
```

* Danach starten mit:
```
docker-compose up -d
```