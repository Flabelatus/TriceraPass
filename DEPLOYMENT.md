
# Auth Service 🔐
**A guildeline to deploy the TriceraPass auth service API 📒**

---

## **Table of Contents** 📖  

1. [Auth Service 🔐](#auth-service-)  
2. [Production App Configs ⚙️](#production-app-configs-️)  
3. [Setup of CORS for Production ⚙️](#setup-of-cors-for-production-️)  
4. [Build the App 🛠️](#build-the-app-️)  
5. [Copying Files to Server 🗄️](#copying-files-to-server-️)  
6. [Reverse Proxy Configurations 🌐](#reverse-proxy-configurations-️)  
7. [Populate the Database 🗂️](#populate-the-database-️)  
8. [Application Runtime 🚀](#application-runtime-️)


---

## Production App Configs ⚙️
 - **Ensure you have added the domain of your service** 
 based on your DNS setup in the `settings.yml` file under the `allowed_origins` field.
 - **Set the values of final production app** 
 in your settings e.g. jwt_secret, database, emailserver etc.
 ```yaml
 api:
        # These parameters are configured for the production server
        allowed_origins:
                - https://hva-robotlab.nl

        # These parameters are configured for the production server
        email_server:
                api_key: your_mailgun_api_key
                domain: your_mailgun_domain

        # These parameters are configured for the production server
        database:
                type: postgres
                host: postgres
                port: 5432
                user: user
                password: password
                dbname: dbname
                sslmode: disable
                timezone: UTC
                connect_timeout: 5
```
- **Export the settings to environment variables**
```bash
#script to export all the environment variables needed for production
./generate-env.sh 
```
## Setup of CORS for the production ⚙️
- **Ensure the following snippet is used**  
in the middleware setup `/cmd/api/application/middleware.go`

**Snippet:**
``` go
func (app *Application) EnableCORS(h http.Handler) http.Handler {

	err := godotenv.Load()
	if err != nil {
		log.Fatal(fmt.Printf("Cannot locate the env file: %v", err))
	}

	allowedOrigins := os.Getenv("CORS")

	if allowedOrigins == "" {
		log.Fatal("No allowed origins found in the environment")
	}

	allowOriginList := strings.Split(allowedOrigins, ",")

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
}
```

---

## Build the app 🛠️
- Once ready, to build the app run this command:

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o robotlabAuth ./cmd/api
```

---

## Copying Files to Server 🗄️
- Copy the compiled app file to the server
Use `scp` to copy the binary file to the server
```bash
scp robotlabAuth .env settings.yml root@152.42.131.10:/var/www/api.robotlabAuth
scp -r static template root@152.42.131.10:/var/www/api.robotlabAuth
```

- **Dumping the Postgresql file**
Since its running in the docker compose follow this command to dump the postgres database file
```bash
docker-compose exec postgres pg_dump --no-owner -h localhost -p 5432 -U robotlab_auth database > database.sql
```

- **Copy the database file to the server**
Use `scp` to copy the sql file to the server
```bash
scp database.sql root@152.42.131.10:/home/robotlabAuth
```

---

## Reverse Proxy Configurations 🌐
- **Install Caddy server**
To install caddy simple run:
```bash
apt install caddy
```
- **Configure the Caddyfile**
Then navigate to `/etc/caddy/` and add the following directives to the `Caddyfile` using `vim` or `nano`

```nginx
{
    email j.jooshesh@hva.nl
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
                root * /var/www/api.robotlabAuth
                file_server
        }

        # Serve static assets (like images, CSS, JS)
        handle_path /assets/* {
                root * /var/www/api.robotlabAuth/template/docs/assets
                file_server
        }

        # Serve documentation HTML files
        handle_path /auth/api/docs/* {
                root * /var/www/api.robotlabAuth/template/docs
                file_server
                try_files {path} /index.html  # Serve index.html if the requested file is missing
        }

        # Reverse proxy godoc server
        handle_path /auth/api/godocs/* {
                uri strip_prefix /auth/api/godocs
                reverse_proxy localhost:6060
        }

        # Fix missing GoDocs styles (CSS, JS)
        handle_path /lib/go-1.22/* {
                reverse_proxy localhost:6060
        }

        handle /auth/api/* {
                reverse_proxy localhost:1993
        }
}

www.hva-robotlab.nl {
        redir https://hva-robotlab.nl

}
```

Then Run the caddy by the following command:
```bash
sudo systemctl start caddy
sudo systemctl status caddy
```

---

## Populate the database 🗂️
- **Install PostgreSQL**
Run the following commands to install the Postgres database in your server
```bash
sudo apt update
sudo apt install postgresql postgresql-contrib
sudo systemctl start postgresql.service
``` 

- **Create a database**
You can sign in as a postgres user and then via `psql` create a database:
        1. log in as postgres
```bash
# Login as postgres user
sudo -i -u postgres
psql -u postgres
```
Then create the database
```SQL
CREATE DATABASE robotlabAuth;
```
After that you can check the connection:
```
\conninfo
\q
```
This will create a database which we can connect. 

- **Generate the DB schema**
Now you would need to use the file which was dumped from our Postgres db in the development repo and generate our schema from
Run the following commands:
```bash
# Navigate to where the dumped SQL file is located
cd /home/robotlabAuth
# Run the psql command using the user, database name and the dump file
sudo -u postgres psql -d robotlabAuth -f database.sql
```
Now your database is populated with the schema

---

## Application Runtime 🚀
- **Install supervisor**
```bash
sudo apt install supervisor
```

- **Go to the directory of supervisor**
- Inside the configurations in the supervisor file we would have to add the apiAuth configuration file as the next step
```bash
cd /etc/supervisor/conf.d
```

- **Create the config file** 
Inside it create a conf file running this. 
```bash
sudo vim apiAuth.conf
```

- **Run the application**
Then enter the following instructions:

```ini
[program:api]
command=/var/www/api.robotlabAuth/robotlabAuth -dsn='host=localhost port=5432 user=postgres password=password dbname=database sslmode=disable' -jwt-secret='sercret' -jwt-issuer='hva-robotlab.nl' -jwt-audience='hva-robotlab.nl' -cookie-domain=''
directory=/var/www/api.robotlabAuth
autorestart=true
autostart=true
stdout_logfile=/var/www/logs/api.logs
```

- **Restart the supervisor** 
After copying the new API application
```bash
sudo systemctl restart supervisor
```