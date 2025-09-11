# Prototype Game Hello World App

This repository provides a minimal React "Hello World" application served by an Express server and deployed with Fly.io.

## Local Development

```bash
npm install
npm test
npm run build
npm start
```

Then visit <http://localhost:3000> and <http://localhost:3000/healthz> for the health check.

## Docker

```bash
docker build -t prototype-game .
docker run -p 3000:3000 prototype-game
```

## Deployment

The repo includes a `fly.toml` configuration for Fly.io.

To deploy manually:

1. Install the Fly.io CLI and log in (`flyctl auth login`).
2. Create a Fly.io app and update `fly.toml` with your app name.
3. Run `flyctl deploy`.

After deployment, your app will be available at `https://<your-app-name>.fly.dev`.

## Health Endpoint

The Express server exposes `/healthz` which returns `200 OK`.

## Notes

This setup can serve as a template for future API and networking services using the same deployment pattern.
