# Robot API Server

A REST API server for controlling and managing robots in a virtual environment. Created as part of the MKSS2 course at HSB Bremen.

## Features

- Robot status management (position, energy, inventory)
- Robot movement and item interaction
- Action history with pagination
- Robot combat system
- HATEOAS navigation links
- **HTTPS support in Azure deployment**

## Quick Start

### Local Development

1. Clone and run:

   ```bash
   git clone <repository-url>
   cd aufgabe-2
   go mod download
   go run .
   ```

2. Test the API:
   ```bash
   curl http://localhost:8080/health
   curl http://localhost:8080/robot/robot1/status
   ```

### Docker

```bash
docker build -t robot-api .
docker run -p 8080:8080 robot-api
```

## API Testing

### Postman Collection

A complete Postman collection is included in the project (`Robot_API.postman_collection.json`)

**Environment variables:**

- `local`: `http://localhost:8080` (local development server)
- `cloudHttp`: `http://robot-api-milad9a.westeurope.cloudapp.azure.com` (cloud HTTP)
- `cloudHttps`: `https://robot-api-milad9a.westeurope.cloudapp.azure.com` (cloud HTTPS)
- `url`: `{{local}}` (active endpoint - change this to switch environments)
- `robotId`: `robot1` (default robot for testing)
- `itemId`: `item1` (default item for testing)

**How to switch environments:**

1. **Local Testing**: Set `url` to `{{local}}`
2. **Cloud HTTP Testing**: Set `url` to `{{cloudHttp}}`
3. **Cloud HTTPS Testing**: Set `url` to `{{cloudHttps}}`

**Note**: Both `cloudHttp` and `cloudHttps` use the same domain (`robot-api-milad9a.westeurope.cloudapp.azure.com`) but different protocols.

## Cloud Deployment

### HTTPS Support

The Azure deployment provides both HTTP and HTTPS on the same domain through Azure Application Gateway:

- **Unified Domain**: `robot-api-milad9a.westeurope.cloudapp.azure.com`
- **HTTP URL**: `http://robot-api-milad9a.westeurope.cloudapp.azure.com`
- **HTTPS URL**: `https://robot-api-milad9a.westeurope.cloudapp.azure.com`
- **Container Direct**: `http://robot-api-milad9a.westeurope.azurecontainer.io:8080`

**Benefits of unified domain:**

- Same URL works for both HTTP and HTTPS
- Easy to switch protocols by just changing `http://` to `https://`
- Standard web practice
- Single SSL certificate covers both

#### SSL Certificate Configuration

To enable full HTTPS support, configure an SSL certificate on the Application Gateway:

1. **Using Azure Key Vault** (Recommended):

   ```bash
   # Create/import certificate in Key Vault
   az keyvault certificate create --vault-name your-vault --name robot-api-cert --policy "$(az keyvault certificate get-default-policy)"

   # Configure Application Gateway to use the certificate
   az network application-gateway ssl-cert create --gateway-name robot-api-gateway --resource-group robotApiGroup --name robot-api-ssl --key-vault-secret-id https://your-vault.vault.azure.net/secrets/robot-api-cert
   ```

2. **Using Let's Encrypt** (Free):

   - Configure automatic certificate renewal
   - Use Azure Key Vault integration

3. **Upload Custom Certificate**:
   ```bash
   az network application-gateway ssl-cert create --gateway-name robot-api-gateway --resource-group robotApiGroup --name robot-api-ssl --cert-file cert.pfx --cert-password yourpassword
   ```

### GitLab CI/CD Pipeline

The project includes automated deployment to Azure Container Instances with HTTPS-ready Application Gateway.

**Required CI/CD Variables:**

- `AZURE_SP_ID`: Azure Service Principal ID
- `AZURE_SP_PASSWORD`: Azure Service Principal Password
- `AZURE_TENANT_ID`: Azure Tenant ID

**The deployment creates:**

- Azure Container Instance (HTTP backend)
- Application Gateway (HTTPS termination)
- Public IP with DNS name
- Virtual Network for Application Gateway

### Security Features

- **HTTPS Redirect**: Application Gateway can redirect HTTP to HTTPS
- **TLS Termination**: SSL/TLS handled at the gateway level
- **Header Detection**: Application detects HTTPS from proxy headers
- **Secure HATEOAS Links**: All links use HTTPS when accessed through secure endpoints

## API Endpoints

| Method | Endpoint                        | Description                    |
| ------ | ------------------------------- | ------------------------------ |
| GET    | `/health`                       | Health check                   |
| GET    | `/robot/{id}/status`            | Get robot status               |
| POST   | `/robot/{id}/move`              | Move robot                     |
| POST   | `/robot/{id}/pickup/{itemId}`   | Pick up item                   |
| POST   | `/robot/{id}/putdown/{itemId}`  | Put down item                  |
| PATCH  | `/robot/{id}/state`             | Update robot state             |
| GET    | `/robot/{id}/actions`           | Get action history (paginated) |
| POST   | `/robot/{id}/attack/{targetId}` | Attack another robot           |
| GET    | `/items`                        | List available items           |

**All endpoints support both HTTP and HTTPS protocols.**

## Testing

```bash
# Run unit tests
go test -v

# Run with coverage
go test -cover
```

## Initial Data

The server starts with:

- **robot1**: Position (0,0), Energy 100, Pre-populated action history
- **robot2**: Position (10,10), Energy 100, Basic action history
- **Items**: item1, item2, item3, item4, item5
