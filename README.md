# Energy Metering API

Backend service untuk IoT energy metering dengan real-time monitoring dan billing calculations.

**Stack:** Go 1.24+ • Gin • PostgreSQL/TimescaleDB • RabbitMQ • WebSocket

## Quick Start

```bash
# Setup
cp .env.example .env  # Edit dengan credentials Anda
go mod download

# Run migrations
psql "$DATABASE_URL" -f migrations/*.sql

# Run
go run ./cmd/server
# atau: make run
```

## Configuration (.env)

```bash
DATABASE_URL=postgres://user:pass@host:port/db?sslmode=require
RABBITMQ_URL=amqp://user:pass@host:5672/
SERVICE_PORT=8080
```

## API Endpoints

- `GET /health` - Health check
- `GET /api/v1/billing` - Billing lengkap (client_id, start_date, end_date, format=pdf)
- `GET /api/v1/billing/preview` - Billing cepat
- `GET /api/v1/dashboard/summary` - Dashboard summary (client_id)
- `GET /api/v1/timeseries` - Data historis (client_id)
- `GET /api/v1/clients` - List clients
- `GET /api/v1/anomalies` - Data anomali (client_id, limit)
- `WS /ws/live` - Real-time updates

**Contoh:**
```bash
curl "localhost:8080/api/v1/billing/preview?client_id=c1&start_date=2024-01-01&end_date=2024-01-31"
```

## Docker

```bash
docker build -t energy-metering-api .
docker run -p 8080:8080 --env-file .env energy-metering-api
```

## Deploy to Fly.io

```bash
flyctl auth login
flyctl launch --no-deploy
flyctl secrets set DATABASE_URL="..." RABBITMQ_URL="..."
flyctl deploy
```

## Development

```bash
make help       # Lihat semua command
make build      # Build binary
make run        # Run server
make test       # Run tests
make deploy-fly # Deploy ke Fly.io
```

## Project Structure

```
cmd/server/      # main.go, controller.go, logger.go
internal/
  ├── billing/   # Billing logic & handlers
  ├── dashboard/ # Dashboard handlers
  ├── repository/ # Data access layer
  ├── websocket/ # WebSocket hub
  └── config/    # Configuration
migrations/      # SQL migrations (run in order)
```

## Migrations

Jalankan dengan `psql "$DATABASE_URL" -f migrations/XXXX.sql` sesuai urutan:
1. `0001_enable_extensions.sql` - Enable TimescaleDB
2. `0002_meter_clients.sql` - Table clients
3. `0003_meter_readings_raw.sql` - Hypertable untuk readings
4. `0004_meter_readings_daily.sql` - Continuous aggregates
5. `0005_billing_invoices.sql` - Table billing

## Production Checklist

- [ ] Set DATABASE_URL & RABBITMQ_URL sebagai secrets (bukan di .env)
- [ ] Gunakan SSL untuk database (sslmode=require)
- [ ] Restrict CORS origins (saat ini allow all `*`)
- [ ] Tambahkan authentication jika diperlukan
- [ ] Setup monitoring & database backups

---

Made with ❤️ for IoT Energy Metering
```

3. **Configure environment**
```bash
cp .env.example .env
# Edit .env with your database and RabbitMQ credentials
```

4. **Run migrations**
```bash
# Apply migrations in order
psql "$DATABASE_URL" -f migrations/0001_enable_extensions.sql
psql "$DATABASE_URL" -f migrations/0002_meter_clients.sql
psql "$DATABASE_URL" -f migrations/0003_meter_readings_raw.sql
psql "$DATABASE_URL" -f migrations/0004_meter_readings_daily.sql
psql "$DATABASE_URL" -f migrations/0005_billing_invoices.sql
```

5. **Run the application**
```bash
# Using Go
go run ./cmd/server

# Or using make
make run

# Or build and run
make build
./bin/server.exe
```

## 📡 API Endpoints

### Health Check
- `GET /health` - Service health status

### Billing
- `GET /api/v1/billing` - Comprehensive billing (requires: client_id, start_date, end_date)
- `GET /api/v1/billing/preview` - Billing preview (requires: client_id, start_date, end_date)

### Dashboard
- `GET /api/v1/dashboard/summary` - Dashboard summary (requires: client_id)

### Timeseries
- `GET /api/v1/timeseries` - Historical meter readings (requires: client_id)

### Clients
- `GET /api/v1/clients` - List all meter clients

### Anomalies
- `GET /api/v1/anomalies` - Anomaly readings (requires: client_id, optional: limit)

### WebSocket
- `WS /ws/live` - Real-time meter readings

## 🐳 Docker Deployment

### Build Docker Image
```bash
docker build -t energy-metering-api:latest .
```

### Run Docker Container
```bash
docker run -p 8080:8080 --env-file .env energy-metering-api:latest
```

## ☁️ Fly.io Deployment

See [DEPLOY.md](DEPLOY.md) for detailed deployment instructions.

### Quick Deploy
```bash
# Install flyctl
curl -L https://fly.io/install.sh | sh

# Login
flyctl auth login

# Deploy
flyctl launch --no-deploy
flyctl secrets set DATABASE_URL="your_db_url"
flyctl secrets set RABBITMQ_URL="your_rabbitmq_url"
flyctl deploy
```

## 🛠️ Development

### Using Makefile
```bash
make help          # Show all available commands
make build         # Build the application
make run           # Run the application
make test          # Run tests
make clean         # Clean build artifacts
make fmt           # Format code
make vet           # Run go vet
make tidy          # Tidy go modules
```

### Using Tasks
```bash
# Available VS Code tasks (Ctrl+Shift+P > Run Task)
- Build Server
- Run Server
- Run Tests
- Go Mod Tidy
```

## 📊 Database Migrations

Migrations are plain SQL files in the `migrations/` directory:

1. **0001_enable_extensions.sql** - Enable TimescaleDB extension
2. **0002_meter_clients.sql** - Create meter_clients table
3. **0003_meter_readings_raw.sql** - Create hypertable for raw readings
4. **0004_meter_readings_daily.sql** - Create continuous aggregates
5. **0005_billing_invoices.sql** - Create billing/invoice tables

**Note**: Migration 0003 enables TimescaleDB compression and retention policies. Ensure proper database privileges.

## 🔌 RabbitMQ Setup
Run the setup script to create exchanges, queues, and bindings (if available in your scripts folder).

## 📝 Logging

The application uses structured logging with Zap:
- All HTTP requests are logged with method, path, status, latency, and client IP
- Service operations are logged at appropriate levels (info, warn, error)
- Database queries can be toggled for debugging (currently commented out)

## 🧪 Testing

```bash
# Run all tests
go test -v ./...

# Or using make
make test

# Run tests with coverage
go test -cover ./...
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is proprietary and confidential.

## 🙋 Support

For issues and questions, please open an issue in the repository.

---

Made with ❤️ for IoT Energy Metering
````
# Make executable
chmod +x scripts/setup-rabbitmq.sh

# Using default (localhost)
./scripts/setup-rabbitmq.sh

# Custom host/credentials
export RABBITMQ_HOST="172.29.228.213"
export RABBITMQ_USER="admin"
export RABBITMQ_PASS="secret"
./scripts/setup-rabbitmq.sh
```

**Resources created:**
- Exchanges: `energy-metering.ingest.exchange`, `energy-metering.worker.events.exchange`
- Queues: `energy-metering.ingest.queue`, `energy-metering.ingest.dlq`, `energy-metering.worker.events.queue`, `energy-metering.ws.bridge`
- Bindings with appropriate routing keys

**Verify:** Open RabbitMQ Management UI at `http://localhost:15672` (guest/guest)

## Running locally
1. Set environment variables above.
2. Fetch dependencies: `go mod tidy`
3. Run:

```bash
go run ./cmd/server
```

## Endpoints
- `GET /api/v1/readings/latest?limit=` - latest raw readings
- `GET /api/v1/readings/timeseries?client_id=&metric=&from=&to=` - timeseries (RFC3339)
- `GET /api/v1/billing/invoice?from=YYYY-MM-DD&to=YYYY-MM-DD` - invoices
- `GET /ws/live` - websocket endpoint for real-time events

## Deployment notes
- Use a read-only DB role for this service.
- Ensure TimescaleDB extension is enabled before running migrations.
- Provide RabbitMQ URL with appropriate permissions to consume.
- For production, use a process supervisor and health checks.
