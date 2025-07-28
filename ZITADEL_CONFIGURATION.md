# Zitadel Configuration for TACACS+ Integration

This guide provides step-by-step instructions for configuring Zitadel Identity Provider to work with the TACACS+ authentication server.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Initial Setup](#initial-setup)
- [Project Configuration](#project-configuration)
- [Application Setup](#application-setup)
- [User Management](#user-management)
- [Role Configuration](#role-configuration)
- [Testing Configuration](#testing-configuration)
- [Advanced Configuration](#advanced-configuration)
- [Troubleshooting](#troubleshooting)

## Prerequisites

Before starting, ensure you have:

- Zitadel instance running (via Docker Compose)
- Access to Zitadel Console at `http://localhost:8080/ui/console`
- Admin credentials for Zitadel
- Basic understanding of OAuth2/OIDC concepts

## Initial Setup

### 1. Access Zitadel Console

1. Open your web browser and navigate to: `http://localhost:8080/ui/console`
2. You will be redirected to the initial setup page if this is a fresh installation

### 2. Complete Initial Setup

If this is a new Zitadel installation:

1. **Create Admin User**:
   - **Email**: `admin@tacacs.local`
   - **First Name**: `TACACS`
   - **Last Name**: `Administrator`
   - **Password**: Choose a strong password (minimum 8 characters)

2. **Organization Setup**:
   - **Organization Name**: `TACACS Network Auth`
   - **Domain**: `tacacs.local`

3. **Complete Setup**: Click "Save" to finish the initial configuration

### 3. Login to Console

After initial setup, login with your admin credentials to access the management console.

## Project Configuration

### 1. Create New Project

1. In the Zitadel Console, navigate to **"Projects"** in the left sidebar
2. Click the **"+ New"** button
3. Fill in the project details:
   - **Name**: `TACACS+ Network Authentication`
   - **Description**: `Authentication and authorization for network devices using TACACS+ protocol`
4. Click **"Save"** to create the project

### 2. Note Project ID

After creating the project:
1. Navigate to the project overview page
2. Copy the **Project ID** from the URL or project details
3. Save this ID - you'll need it for the `.env` configuration

Example Project ID: `220394023840743234`

## Application Setup

### 1. Create API Application

1. Within your TACACS+ project, go to the **"Applications"** tab
2. Click **"+ New"** to create a new application
3. Select **"API"** as the application type
4. Configure the application:
   - **Name**: `tacacs-server`
   - **Description**: `TACACS+ Server Authentication Client`

### 2. Configure Authentication Method

1. In the application settings, go to **"General"** tab
2. Set **Authentication Method** to `CLIENT_SECRET_POST`
3. Click **"Save"** to apply changes

### 3. Generate Client Credentials

1. Go to the **"Client Secret"** section
2. Click **"Generate New Secret"**
3. Copy both the **Client ID** and **Client Secret**
4. Store these securely - you'll need them for the `.env` file

Example:
- **Client ID**: `220394023840743235@tacacs`
- **Client Secret**: `Xy9kL2mN8pQ4rT6uV1wX3yZ5aB7cD9eF`

### 4. Configure Scopes

1. Go to **"Scopes"** tab in the application
2. Ensure the following scopes are available:
   - `openid` (should be enabled by default)
   - `profile`
   - `email`
   - `urn:zitadel:iam:org:project:id:zitadel:aud`

## Role Configuration

### 1. Create TACACS+ Roles

Navigate to the **"Roles"** tab in your project and create the following roles:

#### Network Administrator Role
1. Click **"+ New"**
2. Configure:
   - **Key**: `network-admin`
   - **Display Name**: `Network Administrator`
   - **Description**: `Full administrative access to all network devices and commands`
3. Click **"Save"**

#### Network User Role
1. Click **"+ New"**
2. Configure:
   - **Key**: `network-user`
   - **Display Name**: `Network User`
   - **Description**: `Standard user access with ability to execute most network commands`
3. Click **"Save"**

#### Network Readonly Role
1. Click **"+ New"**
2. Configure:
   - **Key**: `network-readonly`
   - **Display Name**: `Network Read-Only`
   - **Description**: `Read-only access to network devices, limited to show commands`
3. Click **"Save"**

### 2. Role Privilege Mapping

The TACACS+ server maps Zitadel roles to privilege levels:

| Zitadel Role | TACACS+ Privilege Level | Access Description |
|--------------|------------------------|-------------------|
| `network-admin` | 15 | Full administrative access |
| `network-user` | 1 | Standard user commands |
| `network-readonly` | 0 | Read-only commands only |

## User Management

### 1. Create Network Users

Navigate to **"Users"** in the main Zitadel menu:

#### Create Administrator User
1. Click **"+ New"**
2. Fill user details:
   - **Username**: `netadmin`
   - **Email**: `netadmin@tacacs.local`
   - **First Name**: `Network`
   - **Last Name**: `Administrator`
   - **Preferred Language**: `English`
3. Set **Initial Password**: Create a strong password
4. Click **"Save"**

#### Create Standard User
1. Click **"+ New"**
2. Fill user details:
   - **Username**: `netuser`
   - **Email**: `netuser@tacacs.local`
   - **First Name**: `Network`
   - **Last Name**: `User`
   - **Preferred Language**: `English`
3. Set **Initial Password**: Create a strong password
4. Click **"Save"**

#### Create Read-Only User
1. Click **"+ New"**
2. Fill user details:
   - **Username**: `netread`
   - **Email**: `netread@tacacs.local`
   - **First Name**: `Network`
   - **Last Name**: `Reader`
   - **Preferred Language**: `English`
3. Set **Initial Password**: Create a strong password
4. Click **"Save"**

### 2. Assign User Roles

For each created user:

1. Go to **"Users"** and select a user
2. Navigate to **"Authorizations"** tab
3. Click **"+ New"** to add project authorization
4. Select your TACACS+ project
5. Assign appropriate roles:
   - **netadmin**: `network-admin`
   - **netuser**: `network-user`
   - **netread**: `network-readonly`
6. Click **"Save"**

## Advanced Configuration

### 1. Token Configuration

#### Token Lifetime Settings
1. Go to **"Settings"** → **"Login Policy"**
2. Configure token lifetimes:
   - **Access Token Lifetime**: `1h` (3600 seconds)
   - **Refresh Token Lifetime**: `24h` (86400 seconds)
   - **ID Token Lifetime**: `1h` (3600 seconds)

#### Password Policy
1. Go to **"Settings"** → **"Password Policy"**
2. Configure password requirements:
   - **Minimum Length**: `8`
   - **Must have lowercase**: ✅
   - **Must have uppercase**: ✅
   - **Must have number**: ✅
   - **Must have symbol**: ✅ (recommended)

### 2. Branding Configuration

1. Go to **"Settings"** → **"Branding"**
2. Customize appearance:
   - **Primary Color**: `#1f2937` (dark blue)
   - **Background Color**: `#f9fafb` (light gray)
   - **Logo**: Upload your organization logo
   - **Favicon**: Upload custom favicon

### 3. Lockout Policy

1. Go to **"Settings"** → **"Lockout Policy"**
2. Configure security settings:
   - **Max Password Attempts**: `5`
   - **Max OTP Attempts**: `5`

## Testing Configuration

### 1. Test OAuth2 Token Flow

Use curl to test the client credentials flow:

```bash
curl -X POST http://localhost:8080/oauth/v2/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials" \
  -d "client_id=YOUR_CLIENT_ID" \
  -d "client_secret=YOUR_CLIENT_SECRET" \
  -d "scope=openid profile email urn:zitadel:iam:org:project:id:zitadel:aud"
```

Expected response:
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "scope": "openid profile email urn:zitadel:iam:org:project:id:zitadel:aud"
}
```

### 2. Test User Authentication

Test password grant flow:

```bash
curl -X POST http://localhost:8080/oauth/v2/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=password" \
  -d "client_id=YOUR_CLIENT_ID" \
  -d "client_secret=YOUR_CLIENT_SECRET" \
  -d "username=netuser" \
  -d "password=USER_PASSWORD" \
  -d "scope=openid profile email"
```

### 3. Verify User Info

Use the access token to get user information:

```bash
curl -H "Authorization: Bearer ACCESS_TOKEN" \
  http://localhost:8080/oidc/v1/userinfo
```

Expected response:
```json
{
  "sub": "220394023840743236",
  "preferred_username": "netuser",
  "name": "Network User",
  "email": "netuser@tacacs.local",
  "email_verified": true,
  "urn:zitadel:iam:org:project:roles": {
    "network-user": {}
  }
}
```

## Environment Configuration

Update your `.env` file with the Zitadel configuration:

```bash
# Zitadel Configuration
ZITADEL_CLIENT_ID=220394023840743235@tacacs
ZITADEL_CLIENT_SECRET=Xy9kL2mN8pQ4rT6uV1wX3yZ5aB7cD9eF
ZITADEL_PROJECT_ID=220394023840743234

# TACACS+ Configuration
TACACS_SECRET=your_strong_tacacs_secret
TACACS_LISTEN_ADDRESS=0.0.0.0:49

# Database Configuration (shared with Zitadel)
POSTGRES_DB=zitadel
POSTGRES_USER=zitadel
POSTGRES_PASSWORD=zitadel

# Security Settings
SESSION_TIMEOUT=1800
TOKEN_CACHE_TIMEOUT=300
```

## Troubleshooting

### Common Issues

#### 1. Invalid Client Error
**Symptoms**: `invalid_client` error during token requests

**Solutions**:
- Verify Client ID and Client Secret are correct
- Ensure the application is created as "API" type
- Check that authentication method is `CLIENT_SECRET_POST`

#### 2. Scope Not Found Error
**Symptoms**: `invalid_scope` error

**Solutions**:
- Verify all required scopes are enabled in the application
- Check project configuration includes `urn:zitadel:iam:org:project:id:zitadel:aud`

#### 3. User Authentication Fails
**Symptoms**: Users cannot authenticate via TACACS+

**Solutions**:
- Verify user exists and is active in Zitadel
- Check user has correct role assignments
- Ensure password is correct and not expired
- Verify project authorization is granted to user

#### 4. Role Claims Missing
**Symptoms**: User authenticates but has no privileges

**Solutions**:
- Check user is assigned to correct roles in the project
- Verify role keys match exactly (`network-admin`, `network-user`, `network-readonly`)
- Ensure project authorization includes role assignments

### Debug Commands

#### Check Zitadel Health
```bash
curl http://localhost:8080/debug/healthz
```

#### Verify Token Contents
Use [jwt.io](https://jwt.io) to decode access tokens and verify role claims.

#### Check TACACS+ Server Logs
```bash
docker-compose logs -f tacacs-server
```

### Support Resources

- [Zitadel Documentation](https://zitadel.com/docs)
- [OAuth2 RFC 6749](https://tools.ietf.org/html/rfc6749)
- [TACACS+ RFC 8907](https://tools.ietf.org/html/rfc8907)

---

For additional help with Zitadel configuration, refer to the [official Zitadel documentation](https://zitadel.com/docs) or create an issue in the project repository.