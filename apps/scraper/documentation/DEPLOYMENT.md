# Deployment Guide

This guide provides comprehensive instructions for deploying the QuakeWatch Scraper in various environments, from development to production.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Development Environment](#development-environment)
- [Production Deployment](#production-deployment)
- [Docker Deployment](#docker-deployment)
- [Systemd Service](#systemd-service)
- [Monitoring and Logging](#monitoring-and-logging)
- [Security Considerations](#security-considerations)
- [Troubleshooting](#troubleshooting)

## Prerequisites

### System Requirements

- **Operating System**: Linux (Ubuntu 20.04+, CentOS 8+, RHEL 8+), macOS 10.15+, Windows 10+
- **Go Version**: 1.24 or higher
- **Memory**: Minimum 512MB RAM, recommended 2GB+
- **Storage**: Minimum 1GB free space, recommended 10GB+
- **Network**: Internet access for API calls

### Database Requirements (Optional)

- **PostgreSQL**: 12.0 or higher
- **Memory**: Minimum 256MB RAM for PostgreSQL
- **Storage**: Additional storage based on data volume

### Dependencies

```bash
# Ubuntu/Debian
sudo apt update
sudo apt install -y git curl wget postgresql postgresql-contrib

# CentOS/RHEL
sudo yum install -y git curl wget postgresql postgresql-server

# macOS
brew install git curl wget postgresql

# Windows
# Install Git, curl, and PostgreSQL from official websites
```

## Development Environment

### Local Development Setup

1. **Clone the Repository**

```bash
git clone <repository-url>
cd quakewatch-scraper
```

2. **Install Go Dependencies**

```bash
make install
# or
go mod download
go mod tidy
```

3. **Build the Application**

```bash
make build
# or
go build -o bin/quakewatch-scraper cmd/scraper/main.go
```

4. **Set Up Configuration**

```bash
# Copy default configuration
cp configs/config.yaml .

# Edit configuration for your environment
nano config.yaml
```

5. **Set Up Database (Optional)**

```bash
# Start PostgreSQL service
sudo systemctl start postgresql
sudo systemctl enable postgresql

# Create database and user
sudo -u postgres createdb quakewatch
sudo -u postgres psql -c "CREATE USER quakewatch_user WITH PASSWORD 'your_password';"
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE quakewatch TO quakewatch_user;"

# Run database migrations
./bin/quakewatch-scraper db migrate up
```

6. **Test the Application**

```bash
# Check version
./bin/quakewatch-scraper version

# Check health
./bin/quakewatch-scraper health

# Test earthquake collection
./bin/quakewatch-scraper earthquakes recent --limit 10

# Test fault collection
./bin/quakewatch-scraper faults collect
```

### Docker Development Environment

1. **Create Docker Compose File**

```yaml
# docker-compose.dev.yml
version: '3.8'

services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: quakewatch
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5

  scraper:
    build:
      context: .
      dockerfile: Dockerfile
    environment:
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: postgres
      DB_PASSWORD: postgres
      DB_NAME: quakewatch
      DB_SSL_MODE: disable
    volumes:
      - ./data:/app/data
      - ./configs:/app/configs
    depends_on:
      postgres:
        condition: service_healthy
    command: ["sleep", "infinity"]

volumes:
  postgres_data:
```

2. **Create Dockerfile**

```dockerfile
# Dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o quakewatch-scraper cmd/scraper/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app/quakewatch-scraper .
COPY configs/ ./configs/

EXPOSE 8080
CMD ["./quakewatch-scraper"]
```

3. **Start Development Environment**

```bash
# Start services
docker-compose -f docker-compose.dev.yml up -d

# Initialize database
docker-compose -f docker-compose.dev.yml exec scraper ./quakewatch-scraper db init
docker-compose -f docker-compose.dev.yml exec scraper ./quakewatch-scraper db migrate up

# Test application
docker-compose -f docker-compose.dev.yml exec scraper ./quakewatch-scraper version
```

## Production Deployment

### Manual Installation

1. **Download and Install**

```bash
# Download latest release
wget https://github.com/your-org/quakewatch-scraper/releases/latest/download/quakewatch-scraper-linux-amd64
chmod +x quakewatch-scraper-linux-amd64
sudo mv quakewatch-scraper-linux-amd64 /usr/local/bin/quakewatch-scraper

# Create application directory
sudo mkdir -p /opt/quakewatch-scraper
sudo mkdir -p /var/log/quakewatch-scraper
sudo mkdir -p /var/lib/quakewatch-scraper/data
```

2. **Create Configuration**

```bash
sudo nano /opt/quakewatch-scraper/config.yaml
```

```yaml
api:
    emsc:
        base_url: https://www.emsc-csem.org/javascript
        timeout: 30s
    usgs:
        base_url: https://earthquake.usgs.gov/fdsnws/event/1
        rate_limit: 60
        timeout: 30s

collection:
    default_limit: 1000
    max_limit: 10000
    retry_attempts: 3
    retry_delay: 5s

database:
    type: postgres
    host: localhost
    port: 5432
    username: quakewatch_user
    password: your_secure_password
    database: quakewatch
    ssl_mode: require
    max_connections: 20
    connection_timeout: 30s
    enabled: true

logging:
    level: info
    format: json
    output: /var/log/quakewatch-scraper/app.log

storage:
    output_dir: /var/lib/quakewatch-scraper/data
    earthquakes_dir: earthquakes
    faults_dir: faults
```

3. **Create System User**

```bash
sudo useradd -r -s /bin/false quakewatch
sudo chown -R quakewatch:quakewatch /opt/quakewatch-scraper
sudo chown -R quakewatch:quakewatch /var/log/quakewatch-scraper
sudo chown -R quakewatch:quakewatch /var/lib/quakewatch-scraper
```

4. **Set Up Database**

```bash
# Install PostgreSQL
sudo apt install -y postgresql postgresql-contrib

# Create database and user
sudo -u postgres createdb quakewatch
sudo -u postgres psql -c "CREATE USER quakewatch_user WITH PASSWORD 'your_secure_password';"
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE quakewatch TO quakewatch_user;"

# Initialize database
sudo -u quakewatch quakewatch-scraper --config /opt/quakewatch-scraper/config.yaml db init
sudo -u quakewatch quakewatch-scraper --config /opt/quakewatch-scraper/config.yaml db migrate up
```

### Automated Installation Script

```bash
#!/bin/bash
# install.sh

set -e

# Configuration
APP_NAME="quakewatch-scraper"
APP_VERSION="1.0.0"
INSTALL_DIR="/opt/quakewatch-scraper"
LOG_DIR="/var/log/quakewatch-scraper"
DATA_DIR="/var/lib/quakewatch-scraper/data"
SERVICE_USER="quakewatch"
DB_NAME="quakewatch"
DB_USER="quakewatch_user"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Installing QuakeWatch Scraper...${NC}"

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   echo -e "${RED}This script must be run as root${NC}" 
   exit 1
fi

# Create directories
echo "Creating directories..."
mkdir -p $INSTALL_DIR
mkdir -p $LOG_DIR
mkdir -p $DATA_DIR

# Download and install binary
echo "Downloading application..."
wget -O /usr/local/bin/$APP_NAME \
    https://github.com/your-org/quakewatch-scraper/releases/download/v$APP_VERSION/$APP_NAME-linux-amd64
chmod +x /usr/local/bin/$APP_NAME

# Create system user
echo "Creating system user..."
useradd -r -s /bin/false $SERVICE_USER || true

# Set ownership
chown -R $SERVICE_USER:$SERVICE_USER $INSTALL_DIR
chown -R $SERVICE_USER:$SERVICE_USER $LOG_DIR
chown -R $SERVICE_USER:$SERVICE_USER $DATA_DIR

# Create configuration
cat > $INSTALL_DIR/config.yaml << EOF
api:
    emsc:
        base_url: https://www.emsc-csem.org/javascript
        timeout: 30s
    usgs:
        base_url: https://earthquake.usgs.gov/fdsnws/event/1
        rate_limit: 60
        timeout: 30s

collection:
    default_limit: 1000
    max_limit: 10000
    retry_attempts: 3
    retry_delay: 5s

database:
    type: postgres
    host: localhost
    port: 5432
    username: $DB_USER
    password: $(openssl rand -base64 32)
    database: $DB_NAME
    ssl_mode: require
    max_connections: 20
    connection_timeout: 30s
    enabled: true

logging:
    level: info
    format: json
    output: $LOG_DIR/app.log

storage:
    output_dir: $DATA_DIR
    earthquakes_dir: earthquakes
    faults_dir: faults
EOF

chown $SERVICE_USER:$SERVICE_USER $INSTALL_DIR/config.yaml

echo -e "${GREEN}Installation completed successfully!${NC}"
echo -e "${YELLOW}Next steps:${NC}"
echo "1. Configure database: sudo -u postgres createdb $DB_NAME"
echo "2. Initialize database: sudo -u $SERVICE_USER $APP_NAME --config $INSTALL_DIR/config.yaml db init"
echo "3. Set up systemd service: sudo systemctl enable quakewatch-scraper"
```

## Docker Deployment

### Production Docker Compose

```yaml
# docker-compose.yml
version: '3.8'

services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: quakewatch
      POSTGRES_USER: quakewatch_user
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./postgresql.conf:/etc/postgresql/postgresql.conf
    command: postgres -c config_file=/etc/postgresql/postgresql.conf
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U quakewatch_user -d quakewatch"]
      interval: 30s
      timeout: 10s
      retries: 3

  scraper:
    image: quakewatch/scraper:latest
    environment:
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: quakewatch_user
      DB_PASSWORD: ${DB_PASSWORD}
      DB_NAME: quakewatch
      DB_SSL_MODE: disable
      LOG_LEVEL: info
    volumes:
      - scraper_data:/app/data
      - scraper_logs:/app/logs
    depends_on:
      postgres:
        condition: service_healthy
    restart: unless-stopped
    command: ["./quakewatch-scraper", "interval", "earthquakes", "recent", "--interval", "5m", "--duration", "24h"]

  scheduler:
    image: quakewatch/scraper:latest
    environment:
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: quakewatch_user
      DB_PASSWORD: ${DB_PASSWORD}
      DB_NAME: quakewatch
      DB_SSL_MODE: disable
      LOG_LEVEL: info
    volumes:
      - scraper_data:/app/data
      - scraper_logs:/app/logs
    depends_on:
      postgres:
        condition: service_healthy
    restart: unless-stopped
    command: ["./quakewatch-scraper", "interval", "faults", "collect", "--interval", "24h", "--duration", "168h"]

volumes:
  postgres_data:
    driver: local
  scraper_data:
    driver: local
  scraper_logs:
    driver: local
```

### Production Dockerfile

```dockerfile
# Dockerfile.prod
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o quakewatch-scraper cmd/scraper/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app/quakewatch-scraper .

# Create non-root user
RUN addgroup -g 1001 -S quakewatch && \
    adduser -u 1001 -S quakewatch -G quakewatch

# Create directories
RUN mkdir -p /app/data /app/logs && \
    chown -R quakewatch:quakewatch /app

USER quakewatch

EXPOSE 8080
CMD ["./quakewatch-scraper"]
```

### Docker Deployment Commands

```bash
# Build and deploy
docker-compose build
docker-compose up -d

# Check status
docker-compose ps
docker-compose logs -f

# Scale services
docker-compose up -d --scale scraper=3

# Update application
docker-compose pull
docker-compose up -d
```

## Systemd Service

### Service Configuration

```ini
# /etc/systemd/system/quakewatch-scraper.service
[Unit]
Description=QuakeWatch Scraper
Documentation=https://github.com/your-org/quakewatch-scraper
After=network.target postgresql.service
Wants=postgresql.service

[Service]
Type=simple
User=quakewatch
Group=quakewatch
WorkingDirectory=/opt/quakewatch-scraper
ExecStart=/usr/local/bin/quakewatch-scraper --config /opt/quakewatch-scraper/config.yaml interval earthquakes recent --interval 5m --duration 24h
ExecReload=/bin/kill -HUP $MAINPID
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=quakewatch-scraper

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/quakewatch-scraper /var/log/quakewatch-scraper

# Resource limits
LimitNOFILE=65536
LimitNPROC=4096

[Install]
WantedBy=multi-user.target
```

### Service Management

```bash
# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable quakewatch-scraper
sudo systemctl start quakewatch-scraper

# Check status
sudo systemctl status quakewatch-scraper

# View logs
sudo journalctl -u quakewatch-scraper -f

# Restart service
sudo systemctl restart quakewatch-scraper

# Stop service
sudo systemctl stop quakewatch-scraper
```

### Multiple Service Instances

```ini
# /etc/systemd/system/quakewatch-scraper-earthquakes.service
[Unit]
Description=QuakeWatch Scraper - Earthquakes
After=network.target postgresql.service

[Service]
Type=simple
User=quakewatch
Group=quakewatch
WorkingDirectory=/opt/quakewatch-scraper
ExecStart=/usr/local/bin/quakewatch-scraper --config /opt/quakewatch-scraper/config.yaml interval earthquakes recent --interval 5m --duration 24h
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

```ini
# /etc/systemd/system/quakewatch-scraper-faults.service
[Unit]
Description=QuakeWatch Scraper - Faults
After=network.target postgresql.service

[Service]
Type=simple
User=quakewatch
Group=quakewatch
WorkingDirectory=/opt/quakewatch-scraper
ExecStart=/usr/local/bin/quakewatch-scraper --config /opt/quakewatch-scraper/config.yaml interval faults collect --interval 24h --duration 168h
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

## Monitoring and Logging

### Log Configuration

```yaml
# Enhanced logging configuration
logging:
    level: info
    format: json
    output: /var/log/quakewatch-scraper/app.log
    max_size: 100MB
    max_age: 30d
    max_backups: 10
    compress: true
```

### Log Rotation

```bash
# /etc/logrotate.d/quakewatch-scraper
/var/log/quakewatch-scraper/*.log {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    create 644 quakewatch quakewatch
    postrotate
        systemctl reload quakewatch-scraper
    endscript
}
```

### Health Monitoring

```bash
#!/bin/bash
# health-check.sh

# Check application health
if ! curl -f http://localhost:8080/health; then
    echo "Application health check failed"
    systemctl restart quakewatch-scraper
    exit 1
fi

# Check database connectivity
if ! sudo -u quakewatch quakewatch-scraper --config /opt/quakewatch-scraper/config.yaml health --check-db; then
    echo "Database health check failed"
    exit 1
fi

# Check disk space
DISK_USAGE=$(df /var/lib/quakewatch-scraper | tail -1 | awk '{print $5}' | sed 's/%//')
if [ $DISK_USAGE -gt 90 ]; then
    echo "Disk usage is high: ${DISK_USAGE}%"
    exit 1
fi

echo "Health check passed"
```

### Monitoring with Prometheus

```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'quakewatch-scraper'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
    scrape_interval: 30s
```

### Grafana Dashboard

```json
{
  "dashboard": {
    "title": "QuakeWatch Scraper",
    "panels": [
      {
        "title": "Collection Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(earthquakes_collected_total[5m])",
            "legendFormat": "Earthquakes/min"
          }
        ]
      },
      {
        "title": "Database Connections",
        "type": "graph",
        "targets": [
          {
            "expr": "database_connections_active",
            "legendFormat": "Active Connections"
          }
        ]
      }
    ]
  }
}
```

## Security Considerations

### Network Security

```bash
# Firewall configuration
sudo ufw allow 22/tcp
sudo ufw allow 5432/tcp
sudo ufw enable

# PostgreSQL network access
# /etc/postgresql/15/main/postgresql.conf
listen_addresses = 'localhost'
port = 5432

# /etc/postgresql/15/main/pg_hba.conf
local   all             postgres                                peer
local   all             quakewatch_user                         md5
host    quakewatch      quakewatch_user         127.0.0.1/32   md5
```

### Application Security

```yaml
# Security-enhanced configuration
security:
    tls:
        enabled: true
        cert_file: /etc/ssl/certs/quakewatch-scraper.crt
        key_file: /etc/ssl/private/quakewatch-scraper.key
    authentication:
        enabled: true
        api_key: ${API_KEY}
    rate_limiting:
        enabled: true
        requests_per_minute: 100
```

### Secrets Management

```bash
# Using environment variables
export DB_PASSWORD=$(openssl rand -base64 32)
export API_KEY=$(openssl rand -base64 32)

# Using Docker secrets
echo "your_secure_password" | docker secret create db_password -
echo "your_api_key" | docker secret create api_key -
```

## Troubleshooting

### Common Issues

#### 1. Database Connection Issues

```bash
# Check PostgreSQL status
sudo systemctl status postgresql

# Check connection
psql -h localhost -U quakewatch_user -d quakewatch -c "SELECT 1;"

# Check logs
sudo tail -f /var/log/postgresql/postgresql-15-main.log
```

#### 2. Permission Issues

```bash
# Fix file permissions
sudo chown -R quakewatch:quakewatch /opt/quakewatch-scraper
sudo chown -R quakewatch:quakewatch /var/log/quakewatch-scraper
sudo chown -R quakewatch:quakewatch /var/lib/quakewatch-scraper

# Check SELinux (if applicable)
sudo setsebool -P httpd_can_network_connect 1
```

#### 3. Memory Issues

```bash
# Check memory usage
free -h
ps aux | grep quakewatch-scraper

# Adjust PostgreSQL memory settings
# /etc/postgresql/15/main/postgresql.conf
shared_buffers = 256MB
effective_cache_size = 1GB
work_mem = 4MB
```

#### 4. Disk Space Issues

```bash
# Check disk usage
df -h
du -sh /var/lib/quakewatch-scraper/data/*

# Clean up old data
find /var/lib/quakewatch-scraper/data -name "*.json" -mtime +30 -delete
```

### Debug Mode

```bash
# Enable debug logging
sudo systemctl stop quakewatch-scraper
sudo -u quakewatch quakewatch-scraper --config /opt/quakewatch-scraper/config.yaml --verbose earthquakes recent --limit 5

# Check configuration
sudo -u quakewatch quakewatch-scraper --config /opt/quakewatch-scraper/config.yaml config
```

### Performance Tuning

```bash
# Database performance
sudo -u postgres psql -c "VACUUM ANALYZE earthquakes;"
sudo -u postgres psql -c "REINDEX TABLE earthquakes;"

# Application performance
# Adjust collection intervals and limits in configuration
```

This comprehensive deployment guide covers all aspects of deploying the QuakeWatch Scraper in various environments, from simple development setups to production-ready deployments with monitoring and security considerations. 