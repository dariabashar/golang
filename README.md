## Assignment 4 – Go API + Postgres in Docker

This project is a simple Go REST API with PostgreSQL, containerized with a **multi‑stage Dockerfile** and orchestrated via **docker-compose**.

### Services

- **web-app**: Go backend running on port `8080`
- **db**: PostgreSQL `15-alpine` with a named volume and healthcheck

### How to run

1. Make sure Docker and Docker Compose are installed.
2. From the project root, show that nothing is running yet:
   - `docker ps -a`
3. Build and start everything:
   - `docker compose up --build -d`
   - or simply `make up`
4. Check containers:
   - `docker ps -a`
5. Follow logs and see how the app waits for DB healthcheck and then prints **"Starting the Server on :8080"**:
   - `docker compose logs -f`

The API will be available on `http://localhost:8080`.

### API endpoints (CRUD using DB data)

- `GET /items` – list items
- `POST /items` – create item. JSON body:

  ```json
  {
    "name": "Example",
    "description": "Example item"
  }
  ```

- `GET /items/{id}` – get single item
- `PUT /items/{id}` – update item
- `DELETE /items/{id}` – delete item

You can test these with Postman as required in the assignment.

### Database configuration

Configured in `docker-compose.yml`:

- **Service name**: `db` (used as `DB_HOST`)
- **Env vars** (no credentials hardcoded in Go):
  - `DB_HOST=db`
  - `DB_PORT=5432`
  - `DB_USER=appuser`
  - `DB_PASSWORD=appsecret`
  - `DB_NAME=appdb`
  - `DB_SSLMODE=disable`

Schema is initialized from `init.sql` mounted to `/docker-entrypoint-initdb.d/init.sql`, which creates the `items` table.

### Persistence proof

1. Add data via `POST /items`.
2. Run:
   - `docker compose down`
3. Start again:
   - `docker compose up -d`
4. Call `GET /items` – the data is still there because of the named volume `db_data`.

### Image size note

The backend image is built with a **multi‑stage Dockerfile**:

- First stage: `golang:1.25.5-alpine` builds the binary.
- Final stage: small `alpine:3.19` image that contains only the compiled binary.

You can compare image sizes with:

- `docker images`

