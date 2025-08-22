# Futuna

Futuna fetches daily AI-generated recommendations for all HOSE tickers and exposes them via a web dashboard.

## Features
* Daily (weekdays 08:00 GMT+7) analysis of every HOSE stock using OpenAI with deterministic prompts and automatic web search.
* Results stored in PostgreSQL with full history.
* REST API and Next.js dashboard with a modern Tabulator table for filters, search and color‑coded recommendations.
* Kubernetes manifests for Postgres, analyzer CronJob and web server.

## Running locally
```

# edit .env with your keys and database URL

# install front-end deps once
npm --prefix web install

# fetch analyses, start API and dashboard (runs migrations automatically)
./scripts/run-dev.sh

# or run components separately (each runs migrations)
./scripts/run-analyzer.sh
./scripts/run-web.sh   # API server
./scripts/run-client.sh   # front-end
```

## Kubernetes
See manifests in `k8s/`. Build and push the Docker image, then deploy:
```
kubectl apply -f k8s/postgres.yaml
kubectl apply -f k8s/web.yaml
kubectl apply -f k8s/analyzer-cronjob.yaml
```
