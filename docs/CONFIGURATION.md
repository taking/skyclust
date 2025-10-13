# Configuration Management

This document describes how to configure the Cloud Management Portal (CMP) application.

## Configuration Architecture

The CMP application uses a layered configuration approach:

1. **Environment-specific config files** (`configs/config.dev.yaml`, `configs/config.prod.yaml`)
2. **Environment-specific .env files** (`.env.dev`, `.env.prod`)
3. **Default .env file** (`.env`)
4. **System environment variables**

## Configuration Priority

The application loads configuration in the following priority order (highest to lowest):

1. **Command line flags**
2. **System environment variables**
3. **Environment-specific .env file** (`.env.dev` or `.env.prod`)
4. **Default .env file** (`.env`)
5. **Environment-specific config file** (`configs/config.dev.yaml` or `configs/config.prod.yaml`)
6. **Default values**

## File Structure

```
├── configs/
│   ├── config.dev.yaml      # Development configuration (non-sensitive)
│   └── config.prod.yaml     # Production configuration (non-sensitive)
├── .env                     # Default environment variables (Git tracked)
├── .env.dev                 # Development environment variables (Git ignored)
├── .env.prod                # Production environment variables (Git ignored)
└── .env.example             # Template file (Git tracked)
```

## Environment-Specific Configuration

### Development Environment
```bash
# Set environment
export CMP_ENV=development

# The application will automatically load:
# 1. .env.dev (if exists)
# 2. .env (fallback)
# 3. configs/config.dev.yaml
```

### Production Environment
```bash
# Set environment
export CMP_ENV=production

# The application will automatically load:
# 1. .env.prod (if exists)
# 2. .env (fallback)
# 3. configs/config.prod.yaml
```

## Configuration Files

### configs/config.dev.yaml
Contains non-sensitive development settings:
```yaml
server:
  port: 8080
  host: "localhost"
  debug: true

database:
  host: "localhost"
  port: 5432
  name: "cmp"
  user: "cmp_user"
  password: "${CMP_DB_PASSWORD:cmp_password}"  # Environment variable with default
  ssl_mode: "disable"

providers:
  aws:
    access_key: "${CMP_AWS_ACCESS_KEY:}"  # Environment variable, empty default
    secret_key: "${CMP_AWS_SECRET_KEY:}"
    region: "${CMP_AWS_REGION:ap-northeast-2}"
```

### configs/config.prod.yaml
Contains non-sensitive production settings:
```yaml
server:
  port: 8080
  host: "0.0.0.0"
  debug: false

database:
  host: "${CMP_DB_HOST}"
  port: "${CMP_DB_PORT:5432}"
  name: "${CMP_DB_NAME}"
  user: "${CMP_DB_USER}"
  password: "${CMP_DB_PASSWORD}"  # Environment variable required
  ssl_mode: "require"

providers:
  aws:
    access_key: "${CMP_AWS_ACCESS_KEY}"  # Environment variable required
    secret_key: "${CMP_AWS_SECRET_KEY}"
    region: "${CMP_AWS_REGION:ap-northeast-2}"
```

### .env.dev (Development)
Contains sensitive development information:
```bash
# Database
CMP_DB_PASSWORD=cmp_password

# Security
CMP_JWT_SECRET=dev-jwt-secret-key
CMP_ENCRYPTION_KEY=dev-encryption-key-32-bytes

# Cloud Providers
CMP_AWS_ACCESS_KEY=your-dev-aws-key
CMP_AWS_SECRET_KEY=your-dev-aws-secret
```

### .env.prod (Production)
Contains sensitive production information:
```bash
# Database
CMP_DB_PASSWORD=secure-production-password

# Security
CMP_JWT_SECRET=super-secure-jwt-key
CMP_ENCRYPTION_KEY=production-encryption-key

# Cloud Providers
CMP_AWS_ACCESS_KEY=prod-aws-key
CMP_AWS_SECRET_KEY=prod-aws-secret
```

## Docker Configuration

### Development
```bash
# Use development configuration
docker-compose up
```

### Production
```bash
# Use production configuration
docker-compose -f docker-compose.prod.yml up
```

## Security Best Practices

1. **Never commit sensitive files**:
   - `.env.dev`
   - `.env.prod`
   - Any file containing secrets

2. **Use environment-specific files**:
   - Separate development and production secrets
   - Use strong, unique secrets for production

3. **Enable SSL in production**:
   - Set `ssl_mode: "require"` in production config
   - Use secure database connections

4. **Rotate secrets regularly**:
   - Change JWT secrets periodically
   - Update encryption keys as needed
   - Rotate cloud provider credentials

5. **Use secrets management**:
   - Consider using Docker secrets for production
   - Use cloud provider secret management services
   - Implement proper secret rotation policies

## Environment Variable Reference

### Application
- `CMP_ENV`: Environment (development/production)
- `CMP_PORT`: Server port
- `CMP_HOST`: Server host

### Database
- `CMP_DB_HOST`: Database host
- `CMP_DB_PORT`: Database port
- `CMP_DB_NAME`: Database name
- `CMP_DB_USER`: Database user
- `CMP_DB_PASSWORD`: Database password
- `CMP_DB_SSLMODE`: SSL mode

### Security
- `CMP_JWT_SECRET`: JWT signing secret
- `CMP_ENCRYPTION_KEY`: Data encryption key

### Cloud Providers
- `CMP_AWS_ACCESS_KEY`: AWS access key
- `CMP_AWS_SECRET_KEY`: AWS secret key
- `CMP_AWS_REGION`: AWS region
- `CMP_GCP_PROJECT_ID`: GCP project ID
- `CMP_GCP_CREDENTIALS_FILE`: GCP credentials file path