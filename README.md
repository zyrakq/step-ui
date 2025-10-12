# Step-CA Web UI

A modern web interface for managing Step-CA certificates with a clean, intuitive design.

**Repository**: https://github.com/marcin-kruszynski/step-ui.git

## Features

- **Issue Certificates**: Create new certificates with server-generated keypairs
- **Sign CSRs**: Sign certificate signing requests from external systems
- **Certificate Inventory**: View, search, and manage all certificates
- **Renewal & Revocation**: Renew or revoke certificates as needed
- **Download Bundles**: Get PEM or PFX certificate bundles with installation instructions
- **Trust Management**: Easy instructions for installing CA root certificates
- **No Authentication**: Simple setup for trusted networks (MVP)

## Architecture

- **Backend**: Go with Gin framework, calling step CLI as subprocess
- **Frontend**: Next.js 14 with TypeScript and Tailwind CSS
- **Database**: SQLite for certificate inventory and audit logs
- **Containerization**: Docker Compose for easy deployment

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Running Step-CA instance
- JWK provisioner configured for the UI

### 1. Clone and Setup

```bash
git clone https://github.com/marcin-kruszynski/step-ui.git
cd step-ui
```

### 2. Configure Environment

```bash
cp example.env .env
```

Edit `.env` with your Step-CA configuration:

```env
CA_URL=https://ca.home:9000
PROVISIONER_NAME=ui-admin
PROVISIONER_PASSWORD=your-provisioner-password-here
```

### 3. Configure Environment

The root CA certificate will be automatically downloaded from your Step-CA instance at `${CA_URL}/roots.pem`. No manual setup required!

### 4. Start the Application

```bash
docker compose up -d
```

The application will be available at:
- Frontend: http://localhost:3000
- Backend API: http://localhost:8080

## Step-CA Setup

### 1. Create JWK Provisioner

Add a JWK provisioner to your Step-CA configuration:

```json
{
  "name": "ui-admin",
  "type": "JWK",
  "key": {
    "use": "sig",
    "kty": "EC",
    "kid": "your-key-id",
    "crv": "P-256",
    "x": "your-x-coordinate",
    "y": "your-y-coordinate"
  },
  "claims": {
    "maxTLSCertDuration": "2160h",
    "defaultTLSCertDuration": "720h"
  }
}
```

### 2. Generate Provisioner Password

```bash
step ca provisioner add ui-admin --type JWK --create
```

## Usage

### Issue New Certificate

1. Navigate to the "Issue New Certificate" page
2. Enter the Common Name (CN) and Subject Alternative Names (SANs)
3. Select validity period and output format (PEM or PFX)
4. Click "Issue Certificate" to generate and download

### Sign CSR

1. Go to the "Sign CSR" page
2. Upload a CSR file or paste the PEM content
3. Select validity period
4. Click "Sign CSR" to get the signed certificate

### Manage Certificates

1. View all certificates in the "Inventory" page
2. Search and filter by status
3. Renew or revoke certificates as needed
4. Download certificate bundles

### Install CA Trust

1. Go to the "Settings" page
2. Follow platform-specific instructions to install the root CA
3. Test certificate trust with the provided commands

## API Endpoints

The backend provides a REST API:

- `POST /api/certs/issue` - Issue new certificate
- `POST /api/certs/sign-csr` - Sign a CSR
- `GET /api/certs` - List certificates
- `GET /api/certs/:id` - Get specific certificate
- `POST /api/certs/:id/renew` - Renew certificate
- `POST /api/certs/:id/revoke` - Revoke certificate
- `GET /api/settings/ca` - Get CA settings

## Development

### Backend Development

```bash
cd backend
go mod download
go run cmd/main.go
```

### Frontend Development

```bash
cd frontend
npm install
npm run dev
```

## Security Considerations

- **No Authentication**: This MVP has no authentication. Only use on trusted networks
- **Private Keys**: Server-generated keys are not persisted after download
- **CSR Flow**: For maximum security, use the CSR signing flow to keep private keys client-side
- **Audit Logging**: All certificate operations are logged for audit purposes

## Troubleshooting

### Common Issues

1. **Step CLI not found**: Ensure step CLI is installed in the backend container
2. **CA connection failed**: Verify CA_URL and network connectivity
3. **Provisioner authentication**: Check provisioner name and password file
4. **Certificate download**: Ensure browser allows file downloads

### Logs

```bash
# View backend logs
docker-compose logs backend

# View frontend logs
docker-compose logs frontend
```

## License

MIT License - see LICENSE file for details.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## Roadmap

### Phase 2 Features

- [ ] Authentication (OIDC/SSO)
- [ ] Background expiry notifications
- [ ] Auto-renewal
- [ ] ACME wizard for wildcard certificates
- [ ] Push to targets (NPM, nginx, etc.)
- [ ] Multi-user roles and permissions
- [ ] Certificate templates
- [ ] Bulk operations
