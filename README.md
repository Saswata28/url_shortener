## Description
A Simple URL Shortener using Golang with Redis


## Set up
To use it you need docker and docker-compose:
Install Docker from [here](https://docs.docker.com/engine/install/)

Or if you are on Linux, you can just simply install Docker and Docker-compose with:
```
sudo apt install docker.io docker-compose-plugin
```
Now clone this repository:
```
git clone https://github.com/Saswata28/url_shortener.git
```
Move to the repo:
```
cd url_shortener
```

Now make a `.env` file and enter the following inside that `.env` file:

```
DB_ADDR="db:6379" #redis port
DB_PASS=""
APP_PORT=":3000" #main app port
DOMAIN="localhost:3000"
API_QUOTA=10
```

api quota is set as 10, you can change it to any number.

Now to install the dependencies run:
```
go mod tidy
```

`go.mod` file is given but the command will install the dependencies and generate `go.sum` file.


## Run
And run docker compose:
```
docker compose up -d
```
This will pull Redis alpine and Golang alpine image and make 2 images named url_shortener-api and url_shortener-db and set up 2 containers named url_shortener-api-1 and url_shortener-db-1

To stop the containers:
```
docker compose down
```
This will stop and remove the containers, but the images won't be removed.

## Usage
You can use any software to tools to send request like postman, Thunder Client or Curl.

Send a json post request to `http://localhost:3000/api/v1/` , the data should be: `{"url":"https://www.youtube.com/watch?v=tKe4HLTeuNUuu"}`

Example with curl:
```
curl -X POST http://localhost:3000/api/v1/ -H "Content-Type: application/json" -d '{"url":"any_link"}'
```
This will response will be in json like this:
`{"url":"requested_url","short":"shortened_url","expiry":24,"rate_limit":9,"rate_limit_reset":60}`


## Modify
You can modify the rate limit, api quota, time reset, `DOMAIN` variable value in `.env` file to get custom shortened domains instead of getting `localhost` in the shortened url.


## Author
- [Saswata28](https://github.com/Saswata28)


## License
[MIT License](LICENSE)