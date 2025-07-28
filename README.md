# TACACS+ Authentication Server with Zitadel Integration

A robust TACACS+ authentication server that integrates with Zitadel Identity Provider for enterprise-grade authentication and authorization.

## 🚀 Features

- **TACACS+ Protocol Support**: Full support for authentication, authorization, and accounting
- **Zitadel Integration**: Seamless integration with Zitadel identity provider
- **Role-Based Access Control**: Support for network-admin, network-user, and network-readonly roles  
- **Session Management**: PostgreSQL-backed session tracking and logging
- **Docker Ready**: Fully containerized deployment with Docker Compose
- **Health Monitoring**: Built-in health checks and metrics endpoints
- **Token Caching**: Intelligent caching for improved performance

## 📖 Documentation

- **[README.md](README.md)** - Main project documentation
- **[SETUP.md](SETUP.md)** - Detailed setup instructions
- **[ZITADEL_CONFIGURATION.md](ZITADEL_CONFIGURATION.md)** - Complete Zitadel configuration guide

## 📋 Prerequisites

- Docker (>= 20.10)
- Docker Compose (>= 2.0)
- curl (for health checks)
- jq (for JSON processing, optional)

## 🛠️ Quick Start

### 1. Clone and Setup

```bash
git clone https://github.com/yourusername/tacacs4zitadel.git
cd tacacs4zitadel
chmod +x setup.sh
./setup.sh
```

### 2. Manual Setup

If you prefer manual setup:

```bash
# Copy environment template
cp .env.example .env

# Start services
docker-compose up -d

# Check service status
docker-compose ps
```

## ⚙️ Configuration

### Environment Variables

Edit the `.env` file to configure your deployment:

```bash
# Zitadel Configuration
ZITADEL_CLIENT_ID=your_client_id
ZITADEL_CLIENT_SECRET=your_client_secret  
ZITADEL_PROJECT_ID=your_project_id

# TACACS+ Configuration
TACACS_SECRET=your_shared_secret
TACACS_LISTEN_ADDRESS=0.0.0.0:49

# Database Configuration
POSTGRES_DB=zitadel
POSTGRES_USER=zitadel
POSTGRES_PASSWORD=zitadel

# Security Settings
SESSION_TIMEOUT=1800
TOKEN_CACHE_TIMEOUT=300
```

### Zitadel Setup

For detailed Zitadel configuration instructions, see [**ZITADEL_CONFIGURATION.md**](ZITADEL_CONFIGURATION.md).

**Quick Setup:**
1. Access Zitadel Console at `http://localhost:8080/ui/console`
2. Complete initial setup and create admin user
3. Create a new project for TACACS+
4. Create an API application with client credentials
5. Configure roles: `network-admin`, `network-user`, `network-readonly`
6. Create users and assign appropriate roles
7. Update `.env` file with the generated credentials
8. Restart services: `docker-compose restart tacacs-server`

## 📡 Network Device Configuration

### Cisco IOS/IOS-XE Example

```
aaa new-model
aaa authentication login default group tacacs+ local
aaa authorization exec default group tacacs+ local
aaa authorization commands 15 default group tacacs+ local
aaa accounting exec default start-stop group tacacs+
aaa accounting commands 15 default start-stop group tacacs+

tacacs-server host 192.168.1.100 key testing123
tacacs-server directed-request
ip tacacs source-interface Loopback0
```

### Arista EOS Example

```
aaa authentication login default group tacacs+ local
aaa authorization exec default group tacacs+ local
aaa authorization commands all default group tacacs+ local

tacacs-server host 192.168.1.100 key 7 testing123
```

## 🔧 Service Management

### Docker Compose Commands

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services  
docker-compose down

# Restart services
docker-compose restart

# Scale services (if needed)
docker-compose up -d --scale tacacs-server=2
```

### Health Checks

```bash
# Check TACACS+ server health
curl http://localhost:8090/health

# Check Zitadel health  
curl http://localhost:8080/debug/healthz

# Check PostgreSQL
docker-compose exec zitadel-db pg_isready -U zitadel
```

## 📊 Monitoring

### Service URLs

- **Zitadel Console**: http://localhost:8080/ui/console
- **Zitadel Account**: http://localhost:8080/ui/login  
- **TACACS+ Health**: http://localhost:8090/health
- **TACACS+ Metrics**: http://localhost:8090/metrics

### Database Access

```bash
# Connect to PostgreSQL
docker-compose exec zitadel-db psql -U zitadel -d zitadel

# View TACACS+ sessions
SELECT * FROM tacacs_sessions ORDER BY start_time DESC LIMIT 10;

# View command history
SELECT * FROM tacacs_commands ORDER BY timestamp DESC LIMIT 20;
```

## 🧪 Testing

### Automated Testing

```bash
# Run test client
docker-compose --profile testing up test-client

# Run specific tests
docker-compose run --rm test-client go test ./...
```

### Manual Testing

```bash
# Test authentication with tacquito client
docker-compose exec test-client ./test-client -server tacacs-server:49 -secret testing123 -username testuser -password testpass
```

## 🔍 Troubleshooting

### Common Issues

1. **Zitadel not starting**: Check database connectivity and increase startup timeout
2. **Authentication failures**: Verify Zitadel configuration and user roles
3. **Network device connection issues**: Check firewall rules and TACACS+ secret
4. **Permission denied errors**: Ensure proper role assignments in Zitadel

### Debug Mode

Enable debug logging:

```bash
# Edit docker-compose.yml
LOG_LEVEL: "debug"

# Restart services
docker-compose restart tacacs-server
```

### Log Analysis

```bash
# View TACACS+ server logs
docker-compose logs tacacs-server

# View Zitadel logs
docker-compose logs zitadel

# Follow logs in real-time
docker-compose logs -f
```

## 📚 Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Network       │    │   TACACS+       │    │    Zitadel      │
│   Device        │◄──►│   Server        │◄──►│    Identity     │
│   (Router/      │    │                 │    │    Provider     │
│   Switch)       │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │                        │
                              ▼                        ▼
                       ┌─────────────────┐    ┌─────────────────┐
                       │   PostgreSQL    │    │   PostgreSQL    │
                       │   (Sessions &   │    │   (Identity     │
                       │   Commands)     │    │   Data)         │
                       └─────────────────┘    └─────────────────┘
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Commit your changes: `git commit -m 'Add amazing feature'`
4. Push to branch: `git push origin feature/amazing-feature`
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Zitadel](https://github.com/zitadel/zitadel) - Modern identity provider
- [Tacquito](https://github.com/facebookincubator/tacquito) - TACACS+ protocol library
- [PostgreSQL](https://postgresql.org) - Reliable database system

## 📞 Support

For support and questions:

- Create an issue in the GitHub repository
- Check the troubleshooting section above
- Review Zitadel documentation for identity provider setup

---

**Made with ❤️ for network administrators and DevOps engineers**