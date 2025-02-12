### Building the app
To build the app run this command:

`CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o robotlabAuth ./cmd/api`

### Copy the compiled app file to the server
Use `scp` to copy the binary file to the server
`scp robotlabAuth root@152.42.131.10:/var/www/api.robotlabAuth`

### Dumping the Postgresql file
Since its running in the docker compose follow this command to dump the postgres database file
`docker-compose exec postgres pg_dump --no-owner -h localhost -p 5432 -U robotlab_auth database > database.sql`

### Copy the database file to the server
Use `scp` to copy the sql file to the server
`scp database.sql root@152.42.131.10:/home/robotlabAuth`

### Setup of CORS for the production
``` go
return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if isOriginAllowed(origin, allowOriginList) {
			for _, url := range allowOriginList {
				fmt.Println(url)
				w.Header().Set("Access-Control-Allow-Origin", url)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
		}

		if r.Method == "OPTIONS" {
			// w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, X-Auth-Email, X-Auth-Key, X-CSRF-Token, Origin, X-Requested-With, Authorization")
			return
		} else {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			h.ServeHTTP(w, r)
		}
	})
```

## Reverse Proxy Configurations
### Using Caddy server
```
{
    email   <j.jooshesh@hva.nl>
}

(static) {
        @static {
                file
                path *.ico *.css *.js *.gif *.jpg *.jpeg *.png *.svg *.woff *.json *.pdf
        }
        header @static Cache-Control max-age=5184000
}

(security) {
        header {
                # enable HSTS
                Strict-Transport-Security max-age=31536000;
                # disable clients from sniffing the media type
                X-Content-Type-Options nosniff
                # keep referrer data off of HTTP connections
                Referrer-Policy no-referrer-when-downgrade
        }
}

hva-robotlab.nl {
        encode zstd gzip
        import static
        import security

        handle {
                root * /var/www/frontend-app/www.domain.com
                try_files {path} /index.html
                file_server
        }

        handle /auth/api/* {
                reverse_proxy localhost:8080
        }
}

www.hva-robotlab.nl {
        redir https://hva-robotlab.nl
}
```
## Using Supervisor for application runtime
### Install supervisor 
configurations in the supervisor file. Create a conf.d directory and inside it create a conf file running this `sudo vim api.conf`

Then enter the following instructions:

```
[program:api]
command=/var/www/api.robotlabAuth/robotlabAuth -dsn='host=localhost port=5432 user=postgres password=robotlab2025 dbname=database sslmode=disable' -jwt-secret='TRICERATOPLESS-eb5d5e9f-86ac-4766-93e2-d760cbb86e7d' -jwt-issuer='hva-robotlab.nl' -jwt-audience='hva-robotlab.nl' -cookie-domain=''
directory=/var/www/api.robotlabAuth
autorestart=true
autostart=true
stdout_logfile=/var/www/logs/api.logs
```

### Restart the supervisor after copying the new API application
```
sudo systemctl restart supervisor
```