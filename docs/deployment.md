# Deployment Guide

This guide covers deploying the Prototype Game Backend to various environments, from development to production.

## Table of Contents

- [Development Deployment](#development-deployment)
- [Staging Deployment](#staging-deployment)
- [Production Deployment](#production-deployment)
- [Cloud Deployment](#cloud-deployment)
- [Monitoring & Maintenance](#monitoring--maintenance)
- [Troubleshooting](#troubleshooting)

## Development Deployment

### Local Development

#### Quick Start with Make
```bash
# Clone and setup
git clone <repository-url>
cd prototype-game
./scripts/setup.sh

# Start services
make run

# Test connection
make login && make wsprobe TOKEN=$(make login)
```

#### Docker Development Environment
```bash
# Start development stack
docker-compose -f docker-compose.dev.yml up -d

# Check logs
docker-compose -f docker-compose.dev.yml logs -f

# Stop services
docker-compose -f docker-compose.dev.yml down
```

### Development Configuration

Environment variables for development:
```bash
export LOG_LEVEL=debug
export SPATIAL_CELL_SIZE=500
export TICK_RATE=20
export MAX_CONNECTIONS=100
```

## Staging Deployment

### Prerequisites

- Docker and Docker Compose
- PostgreSQL 15+
- Redis 7+
- Nginx (optional, for load balancing)

### Staging Setup

1. **Prepare Environment**
```bash
# Create staging directory
mkdir -p /opt/prototype-game-staging
cd /opt/prototype-game-staging

# Clone repository
git clone <repository-url> .
git checkout staging

# Copy staging configuration
cp configs/.env.production .env
# Edit .env with staging-specific values
```

2. **Configure Services**
```bash
# Edit docker-compose.yml for staging
# Update ports, volumes, and environment variables

# Set up SSL certificates (Let's Encrypt)
sudo apt install certbot
sudo certbot certonly --standalone -d staging.yourdomain.com
```

3. **Deploy Services**
```bash
# Build and start services
docker-compose up -d

# Verify deployment
curl -f http://localhost:8080/healthz
```

### Staging Monitoring

```bash
# View logs
docker-compose logs -f prototype-game

# Monitor resources
docker stats

# Check metrics
curl http://localhost:8080/metrics
```

## Production Deployment

### Infrastructure Requirements

#### Minimum Requirements
- **CPU**: 4 cores
- **Memory**: 8GB RAM
- **Storage**: 50GB SSD
- **Network**: 1Gbps connection

#### Recommended Requirements
- **CPU**: 8+ cores
- **Memory**: 16GB+ RAM
- **Storage**: 100GB+ SSD
- **Network**: 10Gbps connection
- **Load Balancer**: For high availability

### Production Setup

#### 1. Server Preparation

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Create application user
sudo useradd -m -s /bin/bash gameserver
sudo usermod -aG docker gameserver
```

#### 2. Application Deployment

```bash
# Switch to application user
sudo su - gameserver

# Create application directory
mkdir -p /home/gameserver/prototype-game
cd /home/gameserver/prototype-game

# Clone and configure
git clone <repository-url> .
git checkout main

# Set up production configuration
cp configs/.env.production .env
# Configure production-specific settings
```

#### 3. Production Configuration

**Environment Variables** (`.env`):
```bash
# Database
DATABASE_URL=postgres://gameuser:${DB_PASSWORD}@postgres:5432/gamedb
REDIS_URL=redis://redis:6379/0

# Security
JWT_SECRET=${JWT_SECRET}
CORS_ORIGINS=https://yourgame.com

# Performance
GOMAXPROCS=8
MAX_CONNECTIONS=10000
SPATIAL_CELL_SIZE=1000
TICK_RATE=30

# Monitoring
METRICS_ENABLED=true
LOG_LEVEL=info
LOG_FORMAT=json
```

#### 4. SSL/TLS Setup

```bash
# Install Certbot
sudo apt install certbot python3-certbot-nginx

# Obtain SSL certificate
sudo certbot certonly --nginx -d yourgame.com -d api.yourgame.com

# Auto-renewal
sudo crontab -e
# Add: 0 12 * * * /usr/bin/certbot renew --quiet
```

#### 5. Nginx Configuration

Create `/etc/nginx/sites-available/prototype-game`:
```nginx
upstream backend {
    server localhost:8080;
}

server {
    listen 80;
    server_name yourgame.com api.yourgame.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name yourgame.com api.yourgame.com;

    ssl_certificate /etc/letsencrypt/live/yourgame.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/yourgame.com/privkey.pem;

    # Security headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains";
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;
    add_header Referrer-Policy "strict-origin-when-cross-origin";

    # WebSocket support
    location /ws {
        proxy_pass http://backend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # API endpoints
    location / {
        proxy_pass http://backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

#### 6. Start Production Services

```bash
# Start services
docker-compose up -d

# Verify deployment
curl -f https://yourgame.com/healthz

# Check logs
docker-compose logs -f
```

### Production Monitoring

#### System Monitoring
```bash
# Install monitoring tools
sudo apt install htop iotop netstat-nat

# Monitor system resources
htop
iotop
ss -tulpn
```

#### Application Monitoring
```bash
# View application logs
docker-compose logs -f prototype-game

# Monitor metrics
curl https://yourgame.com/metrics

# Database monitoring
docker exec -it postgres_container psql -U gameuser -d gamedb -c "SELECT * FROM pg_stat_activity;"
```

## Cloud Deployment

### AWS Deployment

#### EC2 Setup
```bash
# Launch EC2 instance (t3.large or larger)
# Configure security groups (ports 80, 443, 22)
# Attach Elastic IP

# Install Docker and deploy
ssh -i your-key.pem ubuntu@your-ec2-ip
sudo apt update && sudo apt install docker.io docker-compose
# Follow production setup steps
```

#### RDS Database
```bash
# Create RDS PostgreSQL instance
# Update DATABASE_URL in .env
# Run database migrations
```

#### ElastiCache Redis
```bash
# Create ElastiCache Redis cluster
# Update REDIS_URL in .env
```

### Google Cloud Platform

#### Compute Engine
```bash
# Create VM instance
gcloud compute instances create prototype-game \
    --machine-type=n1-standard-4 \
    --zone=us-central1-a \
    --image-family=ubuntu-2004-lts \
    --image-project=ubuntu-os-cloud

# Deploy application
# Follow production setup steps
```

#### Cloud SQL
```bash
# Create Cloud SQL PostgreSQL instance
gcloud sql instances create game-db \
    --database-version=POSTGRES_15 \
    --tier=db-f1-micro \
    --region=us-central1
```

### Kubernetes Deployment

#### Kubernetes Manifests

**Namespace**:
```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: prototype-game
```

**Deployment**:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: prototype-game
  namespace: prototype-game
spec:
  replicas: 3
  selector:
    matchLabels:
      app: prototype-game
  template:
    metadata:
      labels:
        app: prototype-game
    spec:
      containers:
      - name: prototype-game
        image: prototype-game:latest
        ports:
        - containerPort: 8080
        - containerPort: 8081
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: db-secret
              key: url
        resources:
          requests:
            memory: "1Gi"
            cpu: "500m"
          limits:
            memory: "2Gi"
            cpu: "1000m"
```

## Monitoring & Maintenance

### Health Checks

```bash
#!/bin/bash
# health-check.sh

# Check service health
if ! curl -f http://localhost:8080/healthz > /dev/null 2>&1; then
    echo "Gateway health check failed"
    exit 1
fi

if ! curl -f http://localhost:8081/healthz > /dev/null 2>&1; then
    echo "Simulation health check failed"
    exit 1
fi

echo "All services healthy"
```

### Backup Strategy

```bash
#!/bin/bash
# backup.sh

# Database backup
docker exec postgres_container pg_dump -U gameuser gamedb > "backup_$(date +%Y%m%d_%H%M%S).sql"

# Upload to S3 (optional)
aws s3 cp backup_*.sql s3://your-backup-bucket/
```

### Log Rotation

```bash
# Configure logrotate
sudo tee /etc/logrotate.d/prototype-game << EOF
/var/lib/docker/containers/*/*-json.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    create 0644 root root
}
EOF
```

## Troubleshooting

### Common Issues

#### Service Won't Start
```bash
# Check logs
docker-compose logs prototype-game

# Check resource usage
docker stats

# Verify configuration
docker-compose config
```

#### Database Connection Issues
```bash
# Test database connectivity
docker exec -it postgres_container psql -U gameuser -d gamedb -c "SELECT 1;"

# Check database logs
docker-compose logs postgres
```

#### Performance Issues
```bash
# Monitor CPU/Memory
htop

# Check application metrics
curl http://localhost:8080/metrics | grep -E "(cpu|memory|connection)"

# Profile application
go tool pprof http://localhost:8080/debug/pprof/profile
```

### Emergency Procedures

#### Service Restart
```bash
# Graceful restart
docker-compose restart prototype-game

# Force restart
docker-compose down && docker-compose up -d
```

#### Rollback Deployment
```bash
# Rollback to previous version
git checkout <previous-tag>
docker-compose down
docker-compose build
docker-compose up -d
```

#### Scale Services
```bash
# Scale horizontally
docker-compose up -d --scale prototype-game=3
```

---

For additional support or questions about deployment, please refer to the [Architecture Documentation](architecture.md) or contact the development team.