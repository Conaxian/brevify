# Brevify

Brevify is a simple URL shortener.

## Installation

Before starting the server, make sure that the `PORT` and `DB_URI` environment variables are set. `PORT` specifies which port should the HTTP server running on `localhost` listen to, and `DB_URI` specifies which database should be used for storing link data.

Brevify can also load environment variables from an `.env` file.

If you're running Brevify with a new database for the first time, run the `init.sql` script first.

Brevify is written in Go. The source code and package files are located in `src/`, run `go build` and execute the resulting binary to start the server.

## Usage

Brevify provides two routes:

### `GET /a/:id`

Produces a `302 Found` response when the database contains an entry corresponding to `:id`. The `Location` header will contain a redirect to the URL specified in the entry.

Otherwise, returns `404 Not Found`.

### `POST /a`

The request content type must be `application/json`. If this is false, returns `415 Unsupported Media Type`.

The request body must be a valid JSON object containing an entry `url`, whose value must be a valid URL. If this is false, returns `400 Bad Request`.

The returned response will be a JSON object containing a single entry `id`, whose value is the link ID. The shortened URL can then be accessed by `GET /a/:id`.
