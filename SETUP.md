# TACACS+ with Zitadel Setup Guide

This guide provides detailed instructions for setting up TACACS+ authentication with Zitadel identity provider.

## Prerequisites

Before starting, ensure you have:

- Docker (version 20.10 or higher)
- Docker Compose (version 2.0 or higher)
- curl (for testing endpoints)
- Basic understanding of network authentication

## Step-by-Step Setup

### 1. Environment Preparation

```bash
# Clone the repository
git clone https://github.com/yourusername/tacacs4zitadel.git
cd tacacs4zitadel

# Make setup script executable
chmod +x setup.sh

# Copy environment template
cp .env.example .env
```

### 2. Configure Environment Variables

Edit the `.env` file:

```bash
# Zitadel Configuration (will be filled after Zitadel setup)
ZITADEL_CLIENT_ID=
ZITADEL_CLIENT_SECRET=
ZITADEL_PROJECT_ID=

# TACACS+ Configuration
TACACS_SECRET=your_strong_secret_here
TACACS_LISTEN_ADDRESS=0.0.0.0:49

# Database Configuration
POSTGRES_DB=zitadel
POSTGRES_USER=zitadel
POSTGRES_PASSWORD=zitadel

# Security Settings
SESSION_TIMEOUT=1800
TOKEN_CACHE_TIMEOUT=300
```

### 3. Start Core Services

```bash
# Start database and Zitadel
docker-compose up -d zitadel-db zitadel

# Wait for services to be ready
sleep 60

# Check Zitadel status
curl http://localhost:8080/debug/healthz
```

### 4. Configure Zitadel

#### Access Zitadel Console

1. Open browser to `http://localhost:8080/ui/console`
2. Use default admin credentials:
   - **Username**: `zitadel-admin@zitadel.localhost`
   - **Password**: `Password1!`

#### Create TACACS+ Project

1. Click **"Projects"** in the left menu
2. Click **"+ New"** button
3. Enter project details:
   - **Name**: `TACACS+ Network Authentication`
   - **Description**: `TACACS+ authentication for network devices`
4. Click **"Save"** to create the project
5. **Copy the Project ID** from the URL or project overview

#### Create OAuth Application

1. Within your project, go to **"Applications"** tab
2. Click **"+ New"** button
3. Select **"API"** application type
4. Configure application:
   - **Name**: `tacacs-server`
   - **Authentication Method**: `CLIENT_SECRET_POST`
5. Click **"Save"**
6. **Copy the Client ID** from the application overview
7. Go to **"Client Secret"** tab and generate a new secret
8. **Copy the Client Secret** (save it securely)

#### Create User Roles

1. Go to **"Roles"** tab in your project
2. Create the following roles:

**Network Admin Role:**
- Click **"+ New"**
- **Key**: `network-admin`
- **Display Name**: `Network Administrator`
- **Description**: `Full network device access`

**Network User Role:**
- Click **"+ New"**  
- **Key**: `network-user`
- **Display Name**: `Network User`
- **Description**: `Standard network device access`

**Network Readonly Role:**
- Click **"+ New"**
- **Key**: `network-readonly`  
- **Display Name**: `Network Read Only`
- **Description**: `Read-only network device access`

#### Create Test Users

1. Go to **"Users"** in the main menu
2. Click **"+ New"** button
3. Create users with different roles:

**Admin User:**
- **Username**: `netadmin`
- **Email**: `admin@example.com`
- **First Name**: `Network`
- **Last Name**: `Administrator`
- **Password**: Set a strong password
- **Roles**: Assign `network-admin` role

**Standard User:**
- **Username**: `netuser`
- **Email**: `user@example.com`
- **First Name**: `Network`
- **Last Name**: `User`
- **Password**: Set a strong password
- **Roles**: Assign `network-user` role

**Readonly User:**
- **Username**: `netread`
- **Email**: `readonly@example.com`
- **First Name**: `Network`
- **Last Name**: `Reader`
- **Password**: Set a strong password
- **Roles**: Assign `network-readonly` role

### 5. Update Environment Configuration

Update your `.env` file with the Zitadel credentials:

```bash
# Zitadel Configuration
ZITADEL_CLIENT_ID=your_client_id_here
ZITADEL_CLIENT_SECRET=your_client_secret_here
ZITADEL_PROJECT_ID=your_project_id_here

# Rest of configuration remains the same...
```

### 6. Start TACACS+ Server

```bash
# Start the TACACS+ server
docker-compose up -d tacacs-server

# Check server status
curl http://localhost:8090/health

# View server logs
docker-compose logs -f tacacs-server
```

### 7. Network Device Configuration

Configure your network devices to use the TACACS+ server:

#### Cisco IOS Example

```cisco
! Enable AAA
aaa new-model

! Configure TACACS+ server
tacacs-server host 192.168.1.100 key your_strong_secret_here
tacacs-server directed-request

! Configure authentication
aaa authentication login default group tacacs+ local
aaa authentication enable default group tacacs+ enable

! Configure authorization
aaa authorization exec default group tacacs+ local
aaa authorization commands 15 default group tacacs+ local

! Configure accounting
aaa accounting exec default start-stop group tacacs+
aaa accounting commands 15 default start-stop group tacacs+

! Set source interface (optional)
ip tacacs source-interface Loopback0
```

#### Arista EOS Example

```arista
! Configure TACACS+ server
tacacs-server host 192.168.1.100 key 7 your_strong_secret_here

! Configure AAA
aaa authentication login default group tacacs+ local
aaa authorization exec default group tacacs+ local
aaa authorization commands all default group tacacs+ local
aaa accounting exec default start-stop group tacacs+
aaa accounting commands all default start-stop group tacacs+
```

### 8. Testing Authentication

#### Using Test Client

```bash
# Run the included test client
docker-compose --profile testing up test-client

# Or run manual test
docker-compose run --rm test-client ./test-client \
  -server tacacs-server:49 \
  -secret your_strong_secret_here \
  -username netuser \
  -password user_password
```

#### Direct Network Device Testing

1. SSH to your network device
2. Try logging in with the created users:
   - `netadmin` (should have full access)
   - `netuser` (should have user access)
   - `netread` (should have readonly access)

### 9. Monitoring and Maintenance

#### Check Service Health

```bash
# All services status
docker-compose ps

# TACACS+ server health
curl http://localhost:8090/health

# Zitadel health
curl http://localhost:8080/debug/healthz

# View metrics
curl http://localhost:8090/metrics
```

#### Database Monitoring

```bash
# Connect to database
docker-compose exec zitadel-db psql -U zitadel -d zitadel

# View recent sessions
SELECT username, client_ip, start_time, status 
FROM tacacs_sessions 
ORDER BY start_time DESC 
LIMIT 10;

# View command history
SELECT session_id, command, timestamp, allowed 
FROM tacacs_commands 
ORDER BY timestamp DESC 
LIMIT 20;
```

#### Log Analysis

```bash
# Follow all logs
docker-compose logs -f

# TACACS+ server logs only
docker-compose logs -f tacacs-server

# Zitadel logs only
docker-compose logs -f zitadel
```

## Troubleshooting

### Common Issues

1. **Services won't start**: Check Docker daemon and available ports
2. **Zitadel console inaccessible**: Wait longer for initialization (can take 2-3 minutes)
3. **Authentication failures**: Verify user credentials and role assignments
4. **Network device can't connect**: Check firewall rules and TACACS+ secret

### Debug Steps

1. Enable debug logging:
   ```bash
   # Edit docker-compose.yml, set LOG_LEVEL: "debug"
   docker-compose restart tacacs-server
   ```

2. Check network connectivity:
   ```bash
   # From network device
   telnet 192.168.1.100 49
   ```

3. Verify Zitadel token generation:
   ```bash
   curl -X POST http://localhost:8080/oauth/v2/token \
     -H "Content-Type: application/x-www-form-urlencoded" \
     -d "grant_type=client_credentials&client_id=YOUR_CLIENT_ID&client_secret=YOUR_CLIENT_SECRET"
   ```

## Security Considerations

1. **Change default passwords** immediately
2. **Use strong TACACS+ shared secrets**
3. **Implement network segmentation**
4. **Enable TLS for production** (modify docker-compose.yml)
5. **Regular security updates** of all components
6. **Monitor authentication logs** for suspicious activity

## Production Deployment

For production environments:

1. Use external PostgreSQL database
2. Configure TLS/SSL certificates
3. Set up proper backup procedures
4. Implement log aggregation
5. Configure monitoring and alerting
6. Use secrets management system
7. Implement high availability

---

For additional help, refer to the main README.md or create an issue in the GitHub repository.