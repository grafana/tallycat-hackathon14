# 🤝 Contributing to TallyCat

Thanks for your interest in contributing to **TallyCat**!  
This guide will help you get the project running locally so you can explore, test, and contribute.

---

## 🧰 Prerequisites

Make sure you have the following tools installed:

- ✅ Go 1.21 or later  
- 🐥 [DuckDB](https://duckdb.org/) — used as the local metadata store  
- 🐳 Docker or Podman  
- 📦 OpenTelemetry Collector (optional, for local signal testing)

---

## 🏃‍♂️ Running the TallyCat Server

### 1. Clone the repository

```sh
git clone https://github.com/tallycat/tallycat.git
cd tallycat
```

### 2. Install Go dependencies

```sh
go mod download
```

### 3. Start the server

```sh
go run main.go server
```

This will start the TallyCat server locally at [http://localhost:8080](http://localhost:8080).

---

## 📡 Testing with OpenTelemetry Collector

You can use the provided Docker Compose setup to test with real telemetry:

### 1. Start the OpenTelemetry Collector

```sh
docker compose up -f examples/docker-compose.yml
```

This uses `otel-collector-config.yaml` to route telemetry into TallyCat.

### 2. Stop the services

```sh
docker compose down -f examples/docker-compose.yml
```

---

## 🖥️ Running the UI Locally

### 1. Navigate to the UI directory

```sh
cd ui
```

### 2. Install frontend dependencies

```sh
npm install
```

### 3. Start the development server

```sh
npm run dev
```

The UI will be available at [http://localhost:3000](http://localhost:3000).

### 4. Build the UI for production

```sh
npm run build
```

Output will be written to the `dist/` directory.

---

## 📚 Additional Resources

- [OpenTelemetry Documentation](https://opentelemetry.io/docs/)
- [DuckDB Documentation](https://duckdb.org/docs/)
- [TallyCat GitHub Issues](https://github.com/tallycat/tallycat/issues)

---

## ❓ Need Help?

Found a bug? Have a question?  
Open an issue or start a discussion on the [GitHub repo](https://github.com/tallycat/tallycat/issues). We’d love to hear from you!