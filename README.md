# Robot API Server

A REST API server for controlling and managing robots in a virtual environment. Created as part of the MKSS2 course at HSB Bremen.

## Features

- Robot status management (position, energy, inventory)
- Robot movement and item interaction
- Action history with pagination
- Robot combat system
- HATEOAS navigation links

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

**Default variables:**

- `baseUrl`: `http://localhost:8080` (local development)
- `cloudUrl`: `http://robot-api-Milad9A.westeurope.azurecontainer.io:8080` (cloud deployment)
- `robotId`: `robot1`
- `itemId`: `item1`

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

## Testing

```bash
# Run unit tests
go test -v

# Run with coverage
go test -cover
```

## Cloud Deployment

### GitLab CI/CD Pipeline

The project includes automated deployment to Azure Container Instances via GitLab CI/CD.

**Required CI/CD Variables:**

- `AZURE_SP_ID`: Azure Service Principal ID
- `AZURE_SP_PASSWORD`: Azure Service Principal Password
- `AZURE_TENANT_ID`: Azure Tenant ID

**Create Service Principal:**

```bash
az ad sp create-for-rbac --name "robot-api-sp" --role contributor --scopes /subscriptions/YOUR_SUBSCRIPTION_ID
```

### Manual Deployment

1. **Build and push to GitLab Container Registry:**

   ```bash
   docker build -t robot-api .
   docker tag robot-api registry.gitlab.com/hsbremen/mkss2/sose-2025/labor/YOUR_PROJECT/robot-api:latest
   docker login registry.gitlab.com
   docker push registry.gitlab.com/hsbremen/mkss2/sose-2025/labor/YOUR_PROJECT/robot-api:latest
   ```

2. **Deploy to Azure:**
   ```bash
   az container create \
     --resource-group robotApiGroup \
     --name robot-api-container \
     --image registry.gitlab.com/hsbremen/mkss2/sose-2025/labor/YOUR_PROJECT/robot-api:latest \
     --dns-name-label robot-api-YOUR_USERNAME \
     --ports 8080 \
     --registry-login-server registry.gitlab.com \
     --registry-username YOUR_GITLAB_USERNAME \
     --registry-password YOUR_GITLAB_PAT \
     --environment-variables PORT=8080 GIN_MODE=release
   ```

## Initial Data

The server starts with:

- **robot1**: Position (0,0), Energy 100, Pre-populated action history
- **robot2**: Position (10,10), Energy 100, Basic action history
- **Items**: item1, item2, item3, item4, item5
