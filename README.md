# UALA Tweets

A Twitter-like microservice built with Go, PostgreSQL, Redis, and Kafka.

## 🚀 Features

- User registration and management
- Tweet creation and retrieval
- User following/followers system
- Timeline generation using fan-out approach
- Real-time updates using Kafka
- RESTful API with Swagger documentation
- Containerized with Docker

## 🛠 Prerequisites

- Go 1.19+
- Docker and Docker Compose
- Make (optional, but recommended)

## 📚 Documentation

- [System Architecture](./docs/architecture.md) - High-level overview of the architecture and components.

## 🏗 Project Structure

```
.
├── cmd/                  # Application entry points
├── docs/                 # Documentation
├── internal/
│   ├── adapters/         # External implementations (DB, Kafka, etc.)
│   ├── application/      # Business logic and use cases
│   ├── domain/           # Core business entities and interfaces
│   └── interfaces/       # API handlers and web layer
├── migrations/           # Database migrations
├── scripts/              # Utility scripts
├── .env.example         # Example environment variables
├── docker-compose.yml    # Main Docker Compose file
├── docker-compose.test.yml # Test environment Docker Compose
└── Makefile             # Common tasks
```

## 🚀 Getting Started

### 1. Clone the repository

```bash
git clone https://github.com/fedeberco/uala-tweets.git
cd uala-tweets
```

### 2. Set up environment variables

Copy the example environment file and update the values as needed:

```bash
cp .env.example .env
```

### 3. Start services with Docker

Start all required services (PostgreSQL, Redis, Kafka, ZooKeeper):

```bash
make start-docker
```

### 4. Start the application

#### Option A: Run locally (recommended for development)

```bash
make start-local
```

#### Option B: Run in Docker

```bash
docker-compose up --build
```

The application will be available at http://localhost:8000

## 📚 API Documentation

Once the application is running, access the Swagger UI at:

```
http://localhost:8000/swagger/index.html
```

## 🧪 Running Tests

### Unit Tests

```bash
make test
```


## 🛠 Development

### Common Tasks

#### Local Development
| Command                  | Description                                   |
|--------------------------|-----------------------------------------------|
| `make start-local`       | Start all services in Docker and run app locally |
| `make test`             | Run all tests with test database               |
| `make install-deps`     | Install Go dependencies                       |
| `make swagger`          | Generate Swagger API documentation            |
| `make build`            | Build the application binary                  |
| `make clean`            | Clean build artifacts                        |


#### Docker Management
| Command                  | Description                                   |
|--------------------------|-----------------------------------------------|
| `make start-docker`     | Start all services in Docker                  |
| `make stop-docker`      | Stop all Docker services                      |
| `make docker-logs`      | View container logs                          |
| `make docker-build`     | Build Docker images                          |
| `make docker-up`        | Start services in detached mode               |
| `make docker-down`      | Stop and remove containers                    |
| `make docker-restart`   | Restart all services                         |

#### Database Migrations
| Command                  | Description                                   |
|--------------------------|-----------------------------------------------|
| `make migrate-up`       | Run database migrations                      |
| `make migrate-down`     | Rollback the last migration                  |
| `make migrate-down-all` | Rollback all migrations                      |

#### Test Environment
| Command                  | Description                                   |
|--------------------------|-----------------------------------------------|
| `make test-env-up`      | Start test environment with test database     |
| `make test-env-down`    | Stop test environment                        |


## 🔧 Environment Variables

Key environment variables:

- `PORT`: Application port (default: 8000)
- `DB_URL`: PostgreSQL connection string
- `REDIS_ADDR`: Redis address (default: localhost:6379)
- `KAFKA_BROKER`: Kafka broker address (default: localhost:9092)

## 📦 Dependencies

- **Web Framework**: [Gin](https://github.com/gin-gonic/gin)
- **Database**: PostgreSQL with [lib/pq](https://github.com/lib/pq)
- **Caching**: [Redis](https://redis.io/)
- **Message Broker**: [Kafka](https://kafka.apache.org/)
- **Testing**: [testify](https://github.com/stretchr/testify)
- **Documentation**: [Swagger](https://swagger.io/)

