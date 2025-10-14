# SkyClust 배포 가이드

이 문서는 SkyClust 애플리케이션의 배포 방법과 운영 환경 설정을 설명합니다.

## 배포 개요

SkyClust는 다양한 배포 방식을 지원합니다:

1. **Docker Compose** (권장)
2. **Kubernetes**
3. **전통적인 서버 배포**
4. **클라우드 플랫폼 배포**

## Docker Compose 배포

### 개발 환경

#### 1. 개발용 Docker Compose

```yaml
# docker-compose.dev.yml
version: '3.8'

services:
  app:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=skyclust
      - DB_PASSWORD=password
      - DB_NAME=skyclust
      - REDIS_URL=redis://redis:6379
      - JWT_SECRET=dev-jwt-secret
      - ENCRYPTION_KEY=dev-encryption-key-32-bytes
    depends_on:
      - postgres
      - redis
    volumes:
      - ./plugins:/app/plugins
      - ./logs:/app/logs

  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
    ports:
      - "3000:3000"
    environment:
      - NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
    depends_on:
      - app

  postgres:
    image: postgres:15
    environment:
      - POSTGRES_DB=skyclust
      - POSTGRES_USER=skyclust
      - POSTGRES_PASSWORD=password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./scripts/init-db.sql:/docker-entrypoint-initdb.d/init-db.sql

  redis:
    image: redis:7
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

volumes:
  postgres_data:
  redis_data:
```

#### 2. 개발 환경 실행

```bash
# 개발 환경 시작
docker-compose -f docker-compose.dev.yml up -d

# 로그 확인
docker-compose -f docker-compose.dev.yml logs -f

# 개발 환경 중지
docker-compose -f docker-compose.dev.yml down
```

### 프로덕션 환경

#### 1. 프로덕션용 Docker Compose

```yaml
# docker-compose.yml
version: '3.8'

services:
  app:
    image: skyclust:latest
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=${DB_USER}
      - DB_PASSWORD=${DB_PASSWORD}
      - DB_NAME=${DB_NAME}
      - REDIS_URL=redis://redis:6379
      - JWT_SECRET=${JWT_SECRET}
      - ENCRYPTION_KEY=${ENCRYPTION_KEY}
    depends_on:
      - postgres
      - redis
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  frontend:
    image: skyclust-frontend:latest
    ports:
      - "3000:3000"
    environment:
      - NEXT_PUBLIC_API_URL=${API_URL}
    depends_on:
      - app
    restart: unless-stopped

  postgres:
    image: postgres:15
    environment:
      - POSTGRES_DB=${DB_NAME}
      - POSTGRES_USER=${DB_USER}
      - POSTGRES_PASSWORD=${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./scripts/init-db.sql:/docker-entrypoint-initdb.d/init-db.sql
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${DB_USER} -d ${DB_NAME}"]
      interval: 30s
      timeout: 10s
      retries: 3

  redis:
    image: redis:7
    volumes:
      - redis_data:/data
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 30s
      timeout: 10s
      retries: 3

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/nginx/ssl
    depends_on:
      - app
      - frontend
    restart: unless-stopped

volumes:
  postgres_data:
  redis_data:
```

#### 2. 환경 변수 설정

```bash
# .env
DB_USER=skyclust
DB_PASSWORD=secure-password
DB_NAME=skyclust
JWT_SECRET=your-jwt-secret-key
ENCRYPTION_KEY=your-32-byte-encryption-key
API_URL=https://api.skyclust.com/api/v1
```

#### 3. 프로덕션 환경 실행

```bash
# 프로덕션 환경 시작
docker-compose up -d

# 로그 확인
docker-compose logs -f

# 프로덕션 환경 중지
docker-compose down
```

## Kubernetes 배포

### 1. 네임스페이스 생성

```yaml
# namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: skyclust
```

### 2. ConfigMap 생성

```yaml
# configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: skyclust-config
  namespace: skyclust
data:
  config.yaml: |
    server:
      port: 8080
      host: "0.0.0.0"
    database:
      host: "postgres"
      port: 5432
      name: "skyclust"
    redis:
      url: "redis://redis:6379"
```

### 3. Secret 생성

```yaml
# secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: skyclust-secret
  namespace: skyclust
type: Opaque
data:
  db-password: <base64-encoded-password>
  jwt-secret: <base64-encoded-jwt-secret>
  encryption-key: <base64-encoded-encryption-key>
```

### 4. PostgreSQL 배포

```yaml
# postgres.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: postgres
  namespace: skyclust
spec:
  replicas: 1
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      containers:
      - name: postgres
        image: postgres:15
        env:
        - name: POSTGRES_DB
          value: "skyclust"
        - name: POSTGRES_USER
          value: "skyclust"
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: skyclust-secret
              key: db-password
        ports:
        - containerPort: 5432
        volumeMounts:
        - name: postgres-storage
          mountPath: /var/lib/postgresql/data
      volumes:
      - name: postgres-storage
        persistentVolumeClaim:
          claimName: postgres-pvc
---
apiVersion: v1
kind: Service
metadata:
  name: postgres
  namespace: skyclust
spec:
  selector:
    app: postgres
  ports:
  - port: 5432
    targetPort: 5432
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: postgres-pvc
  namespace: skyclust
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
```

### 5. Redis 배포

```yaml
# redis.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis
  namespace: skyclust
spec:
  replicas: 1
  selector:
    matchLabels:
      app: redis
  template:
    metadata:
      labels:
        app: redis
    spec:
      containers:
      - name: redis
        image: redis:7
        ports:
        - containerPort: 6379
        volumeMounts:
        - name: redis-storage
          mountPath: /data
      volumes:
      - name: redis-storage
        persistentVolumeClaim:
          claimName: redis-pvc
---
apiVersion: v1
kind: Service
metadata:
  name: redis
  namespace: skyclust
spec:
  selector:
    app: redis
  ports:
  - port: 6379
    targetPort: 6379
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: redis-pvc
  namespace: skyclust
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 5Gi
```

### 6. 애플리케이션 배포

```yaml
# app.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: skyclust-app
  namespace: skyclust
spec:
  replicas: 3
  selector:
    matchLabels:
      app: skyclust-app
  template:
    metadata:
      labels:
        app: skyclust-app
    spec:
      containers:
      - name: app
        image: skyclust:latest
        ports:
        - containerPort: 8080
        env:
        - name: DB_HOST
          value: "postgres"
        - name: DB_PORT
          value: "5432"
        - name: DB_USER
          value: "skyclust"
        - name: DB_NAME
          value: "skyclust"
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: skyclust-secret
              key: db-password
        - name: REDIS_URL
          value: "redis://redis:6379"
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: skyclust-secret
              key: jwt-secret
        - name: ENCRYPTION_KEY
          valueFrom:
            secretKeyRef:
              name: skyclust-secret
              key: encryption-key
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: skyclust-app
  namespace: skyclust
spec:
  selector:
    app: skyclust-app
  ports:
  - port: 8080
    targetPort: 8080
  type: ClusterIP
```

### 7. Ingress 설정

```yaml
# ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: skyclust-ingress
  namespace: skyclust
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  tls:
  - hosts:
    - skyclust.com
    - api.skyclust.com
    secretName: skyclust-tls
  rules:
  - host: skyclust.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: skyclust-frontend
            port:
              number: 3000
  - host: api.skyclust.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: skyclust-app
            port:
              number: 8080
```

## 클라우드 플랫폼 배포

### AWS 배포

#### 1. ECS 배포

```yaml
# ecs-task-definition.json
{
  "family": "skyclust",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "1024",
  "memory": "2048",
  "executionRoleArn": "arn:aws:iam::account:role/ecsTaskExecutionRole",
  "taskRoleArn": "arn:aws:iam::account:role/ecsTaskRole",
  "containerDefinitions": [
    {
      "name": "skyclust-app",
      "image": "skyclust:latest",
      "portMappings": [
        {
          "containerPort": 8080,
          "protocol": "tcp"
        }
      ],
      "environment": [
        {
          "name": "DB_HOST",
          "value": "skyclust-db.cluster-xyz.us-east-1.rds.amazonaws.com"
        },
        {
          "name": "REDIS_URL",
          "value": "redis://skyclust-redis.xyz.cache.amazonaws.com:6379"
        }
      ],
      "secrets": [
        {
          "name": "DB_PASSWORD",
          "valueFrom": "arn:aws:secretsmanager:us-east-1:account:secret:skyclust/db-password"
        },
        {
          "name": "JWT_SECRET",
          "valueFrom": "arn:aws:secretsmanager:us-east-1:account:secret:skyclust/jwt-secret"
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/skyclust",
          "awslogs-region": "us-east-1",
          "awslogs-stream-prefix": "ecs"
        }
      }
    }
  ]
}
```

#### 2. RDS 설정

```bash
# RDS 인스턴스 생성
aws rds create-db-instance \
  --db-instance-identifier skyclust-db \
  --db-instance-class db.t3.micro \
  --engine postgres \
  --engine-version 15.2 \
  --master-username skyclust \
  --master-user-password secure-password \
  --allocated-storage 20 \
  --vpc-security-group-ids sg-12345678 \
  --db-subnet-group-name skyclust-subnet-group
```

#### 3. ElastiCache 설정

```bash
# Redis 클러스터 생성
aws elasticache create-cache-cluster \
  --cache-cluster-id skyclust-redis \
  --cache-node-type cache.t3.micro \
  --engine redis \
  --num-cache-nodes 1 \
  --vpc-security-group-ids sg-12345678 \
  --cache-subnet-group-name skyclust-cache-subnet-group
```

### GCP 배포

#### 1. Cloud Run 배포

```yaml
# cloud-run.yaml
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: skyclust
  annotations:
    run.googleapis.com/ingress: all
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/maxScale: "10"
        run.googleapis.com/cpu-throttling: "false"
    spec:
      containers:
      - image: gcr.io/PROJECT_ID/skyclust:latest
        ports:
        - containerPort: 8080
        env:
        - name: DB_HOST
          value: "skyclust-db.region.gcp.cloud"
        - name: REDIS_URL
          value: "redis://skyclust-redis.region.cache.gcp.cloud:6379"
        resources:
          limits:
            cpu: "1000m"
            memory: "2Gi"
```

#### 2. Cloud SQL 설정

```bash
# Cloud SQL 인스턴스 생성
gcloud sql instances create skyclust-db \
  --database-version=POSTGRES_15 \
  --tier=db-f1-micro \
  --region=us-central1 \
  --storage-type=SSD \
  --storage-size=20GB
```

### Azure 배포

#### 1. Container Instances 배포

```yaml
# container-instances.yaml
apiVersion: 2018-10-01
location: eastus
name: skyclust
properties:
  containers:
  - name: skyclust-app
    properties:
      image: skyclust.azurecr.io/skyclust:latest
      ports:
      - port: 8080
      environmentVariables:
      - name: DB_HOST
        value: "skyclust-db.postgres.database.azure.com"
      - name: REDIS_URL
        value: "redis://skyclust-redis.redis.cache.windows.net:6380"
      resources:
        requests:
          cpu: 1
          memoryInGb: 2
  osType: Linux
  ipAddress:
    type: Public
    ports:
    - protocol: tcp
      port: 8080
  restartPolicy: Always
```

## 모니터링 및 로깅

### 1. Prometheus 설정

```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
- job_name: 'skyclust'
  static_configs:
  - targets: ['skyclust-app:8080']
  metrics_path: '/metrics'
  scrape_interval: 5s
```

### 2. Grafana 대시보드

```json
{
  "dashboard": {
    "title": "SkyClust Monitoring",
    "panels": [
      {
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total[5m])",
            "legendFormat": "{{method}} {{endpoint}}"
          }
        ]
      },
      {
        "title": "Response Time",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "95th percentile"
          }
        ]
      }
    ]
  }
}
```

### 3. 로그 수집

```yaml
# fluentd.conf
<source>
  @type tail
  path /var/log/skyclust/*.log
  pos_file /var/log/fluentd/skyclust.log.pos
  tag skyclust.*
  format json
</source>

<match skyclust.**>
  @type elasticsearch
  host elasticsearch
  port 9200
  index_name skyclust
  type_name log
</match>
```

## 보안 설정

### 1. SSL/TLS 설정

```nginx
# nginx.conf
server {
    listen 443 ssl http2;
    server_name skyclust.com;
    
    ssl_certificate /etc/nginx/ssl/skyclust.crt;
    ssl_certificate_key /etc/nginx/ssl/skyclust.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512;
    ssl_prefer_server_ciphers off;
    
    location / {
        proxy_pass http://skyclust-app:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### 2. 방화벽 설정

```bash
# UFW 설정
ufw allow 22/tcp
ufw allow 80/tcp
ufw allow 443/tcp
ufw enable
```

### 3. 보안 그룹 설정

```yaml
# security-groups.yaml
apiVersion: v1
kind: NetworkPolicy
metadata:
  name: skyclust-network-policy
  namespace: skyclust
spec:
  podSelector:
    matchLabels:
      app: skyclust-app
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to:
    - podSelector:
        matchLabels:
          app: postgres
    ports:
    - protocol: TCP
      port: 5432
  - to:
    - podSelector:
        matchLabels:
          app: redis
    ports:
    - protocol: TCP
      port: 6379
```

## 백업 및 복구

### 1. 데이터베이스 백업

```bash
# PostgreSQL 백업
pg_dump -h localhost -U skyclust -d skyclust > backup.sql

# 자동 백업 스크립트
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)
pg_dump -h localhost -U skyclust -d skyclust > /backups/skyclust_$DATE.sql
find /backups -name "skyclust_*.sql" -mtime +7 -delete
```

### 2. Redis 백업

```bash
# Redis 백업
redis-cli --rdb /backups/redis_$(date +%Y%m%d_%H%M%S).rdb

# 자동 백업 스크립트
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)
redis-cli --rdb /backups/redis_$DATE.rdb
find /backups -name "redis_*.rdb" -mtime +7 -delete
```

### 3. 복구 절차

```bash
# 데이터베이스 복구
psql -h localhost -U skyclust -d skyclust < backup.sql

# Redis 복구
redis-cli --pipe < redis_backup.rdb
```

## 성능 최적화

### 1. 데이터베이스 최적화

```sql
-- 인덱스 생성
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_workspaces_owner_id ON workspaces(owner_id);
CREATE INDEX idx_vms_workspace_id ON vms(workspace_id);

-- 쿼리 최적화
EXPLAIN ANALYZE SELECT * FROM users WHERE email = 'user@example.com';
```

### 2. Redis 최적화

```bash
# Redis 설정 최적화
redis-cli CONFIG SET maxmemory 1gb
redis-cli CONFIG SET maxmemory-policy allkeys-lru
redis-cli CONFIG SET save 900 1
```

### 3. 애플리케이션 최적화

```yaml
# 리소스 제한 설정
resources:
  limits:
    cpu: "1000m"
    memory: "2Gi"
  requests:
    cpu: "500m"
    memory: "1Gi"
```

## 문제 해결

### 1. 일반적인 문제

#### 데이터베이스 연결 실패
```bash
# 연결 상태 확인
kubectl exec -it skyclust-app-xxx -- curl -f http://localhost:8080/health/db

# 로그 확인
kubectl logs skyclust-app-xxx | grep -i database
```

#### Redis 연결 실패
```bash
# Redis 상태 확인
kubectl exec -it skyclust-app-xxx -- curl -f http://localhost:8080/health/redis

# Redis 연결 테스트
kubectl exec -it skyclust-app-xxx -- redis-cli -h redis ping
```

#### 메모리 부족
```bash
# 메모리 사용량 확인
kubectl top pods -n skyclust

# 리소스 제한 확인
kubectl describe pod skyclust-app-xxx
```

### 2. 로그 분석

```bash
# 애플리케이션 로그
kubectl logs -f skyclust-app-xxx

# 특정 로그 필터링
kubectl logs skyclust-app-xxx | grep ERROR

# 로그 레벨 변경
kubectl exec -it skyclust-app-xxx -- curl -X POST http://localhost:8080/debug/log-level -d "level=debug"
```

### 3. 성능 모니터링

```bash
# CPU 사용량
kubectl top pods -n skyclust

# 메모리 사용량
kubectl top nodes

# 네트워크 트래픽
kubectl exec -it skyclust-app-xxx -- netstat -i
```

## 업데이트 및 배포

### 1. 무중단 배포

```bash
# 새 버전 배포
kubectl set image deployment/skyclust-app app=skyclust:v2.0.0

# 배포 상태 확인
kubectl rollout status deployment/skyclust-app

# 롤백
kubectl rollout undo deployment/skyclust-app
```

### 2. 블루-그린 배포

```bash
# 새 환경 배포
kubectl apply -f skyclust-v2.yaml

# 트래픽 전환
kubectl patch service skyclust-app -p '{"spec":{"selector":{"version":"v2"}}}'

# 이전 환경 제거
kubectl delete -f skyclust-v1.yaml
```

### 3. 카나리 배포

```yaml
# canary-deployment.yaml
apiVersion: argoproj.io/v1alpha1
kind: Rollout
metadata:
  name: skyclust-app
spec:
  replicas: 10
  strategy:
    canary:
      steps:
      - setWeight: 10
      - pause: {duration: 10m}
      - setWeight: 20
      - pause: {duration: 10m}
      - setWeight: 50
      - pause: {duration: 10m}
  selector:
    matchLabels:
      app: skyclust-app
  template:
    metadata:
      labels:
        app: skyclust-app
    spec:
      containers:
      - name: app
        image: skyclust:v2.0.0
        ports:
        - containerPort: 8080
```
