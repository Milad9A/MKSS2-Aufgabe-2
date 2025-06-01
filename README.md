# Robot API Server

A simple REST API server for controlling and managing robots in a virtual environment. This project was created as part of the MKSS2 course.

## Features

- Get robot status (position, energy, inventory)
- Move robots in different directions
- Pickup and put down items
- Update robot state
- Track robot actions
- Robot combat system
- API documentation with Swagger

## Setup and Installation

### Installation

1. Clone the repository:

   ```
   git clone <repository-url>
   cd aufgabe-2
   ```

2. Install dependencies:

   ```
   go mod download
   ```

3. Run the server:
   ```
   go run .
   ```

The server will start on port 8080.

### Docker Setup

You can also run the application in a Docker container:

1. Build the Docker image:

   ```
   docker build -t robot-api .
   ```

2. Run the container:

   ```
   docker run -p 8080:8080 --name robot-container robot-api
   ```

3. The API will be accessible at http://localhost:8080

4. To stop the container:

   ```
   docker stop robot-container
   ```

5. If you need to debug inside the container:
   ```
   docker exec -it robot-container /bin/sh
   ```

## Automated Testing

The project includes comprehensive unit tests for all API endpoints. The tests verify both the successful operations and error handling of the API.

### Running the Tests

To run all tests:

```
go test -v
```

### Test Coverage

To run tests with coverage report:

```
go test -cover
```

### What's Tested

The test suite covers:

- Getting robot status with HATEOAS links
- Moving robots in different directions
- Picking up and putting down items
- Updating robot state (energy and position)
- Retrieving paginated action history
- Robot combat system
- Error handling for non-existent robots and items
- Pagination navigation for action history

Each test verifies both the HTTP response status and the actual data modifications to ensure the API functions correctly.

## Testing with Postman

### Import the Collection

1. Download the Postman collection file from the project directory:

   - `Robot_API.postman_collection.json`

2. Open Postman and click on "Import" in the top left corner

3. Drag and drop the collection file or click "Upload Files" to select it

4. The Robot API collection will be imported with all endpoints ready to test

### Using the Collection

The collection includes all available endpoints with pre-configured:

- Request methods (GET, POST, PATCH)
- URL parameters
- Example request bodies for POST/PATCH requests

To test the API:

1. Start the server: `go run .`
2. Open the imported collection in Postman
3. Select an endpoint and click "Send"
4. View the response in the response panel

### Collection Environment Variables

The collection uses the following variables:

- `baseUrl`: Set to `http://localhost:8080` by default
- `robotId`: Set to `robot1` by default
- `itemId`: Set to `item1` by default

You can modify these variables in the Postman environment settings.

## API Endpoints

### Get Robot Status

```
GET /robot/:id/status
```

Returns the current status of a robot including position, energy, and inventory.

### Move Robot

```
POST /robot/:id/move
```

Move a robot in a specific direction.

Request body:

```json
{
	"direction": "up" // "up", "down", "left", "right"
}
```

### Pickup Item

```
POST /robot/:id/pickup/:itemId
```

Makes a robot pick up an item and add it to its inventory.

### Put Down Item

```
POST /robot/:id/putdown/:itemId
```

Makes a robot put down an item from its inventory.

### Update Robot State

```
PATCH /robot/:id/state
```

Updates robot's state (energy and/or position).

Request body:

```json
{
	"energy": 100,
	"position": {
		"x": 10,
		"y": 20
	}
}
```

### Get Robot Actions

```
GET /robot/:id/actions
```

Returns a history of actions performed by the robot.

### Attack Robot

```
POST /robot/:id/attack/:targetId
```

Makes one robot attack another, reducing the target's energy.

## Examples

### Get Robot Status

```
curl -X GET http://localhost:8080/robot/robot1/status
```

### Move Robot

```
curl -X POST http://localhost:8080/robot/robot1/move -H "Content-Type: application/json" -d '{"direction": "up"}'
```

### Pickup Item

```
curl -X POST http://localhost:8080/robot/robot1/pickup/item1
```

## Initial Data

The server starts with two robots:

- robot1:
  - Position: (0,0)
  - Direction: north
  - Energy: 100
  - Action history: 7 pre-populated actions including creation, movement, item interactions, and combat
- robot2:
  - Position: (10,10)
  - Direction: south
  - Energy: 100
  - Action history: 3 pre-populated actions including creation, movement, and being damaged

And three items in the world:

- item1
- item2
- item3

This pre-populated data allows you to immediately test the API's functionality, including the pagination of action history and HATEOAS navigation links.

## Cloud Deployment

### Manual Deployment Steps

#### Step 1: Create GitLab Personal Access Token

1. Go to GitLab → Settings → Access Tokens
2. Create a new token with:
   - Name: "Docker Registry Access"
   - Scopes: `read_registry`, `write_registry`
   - Expiration: Choose a long duration (e.g., 1 year)
3. Save the token securely - you'll need it for Docker login

#### Step 2: Build and Push to GitLab Container Registry

1. Build the Docker image locally:

   ```bash
   docker build -t robot-api .
   ```

2. Tag the image using the required HSB naming schema:

   ```bash
   docker tag robot-api registry.gitlab.com/hsbremen/mkss2/sose-2025/labor/YOUR_PROJECT_NAME/robot-api:latest
   ```

   Replace `YOUR_PROJECT_NAME` with your actual GitLab project name.

3. Login to GitLab Container Registry:

   ```bash
   docker login registry.gitlab.com
   ```

   - Username: Your GitLab username
   - Password: The PAT you created in Step 1

4. Push the image:

   ```bash
   docker push registry.gitlab.com/hsbremen/mkss2/sose-2025/labor/YOUR_PROJECT_NAME/robot-api:latest
   ```

#### Step 3: Deploy to Azure Container Instances

1. Login to Azure CLI:

   ```bash
   az login
   ```

2. Create a resource group:

   ```bash
   az group create --name robotApiGroup --location westeurope
   ```

3. Register the Container Instance provider:

   ```bash
   az provider register --namespace Microsoft.ContainerInstance
   ```

4. Wait for provider registration (this may take a few minutes):

   ```bash
   az provider show --namespace Microsoft.ContainerInstance --query registrationState
   ```

5. Deploy the container:

   ```bash
   az container create \
     --resource-group robotApiGroup \
     --name robot-api-container \
     --image registry.gitlab.com/hsbremen/mkss2/sose-2025/labor/YOUR_PROJECT_NAME/robot-api:latest \
     --dns-name-label robot-api-YOUR_USERNAME \
     --ports 8080 \
     --registry-login-server registry.gitlab.com \
     --registry-username YOUR_GITLAB_USERNAME \
     --registry-password YOUR_GITLAB_PAT \
     --os-type Linux \
     --cpu 1 \
     --memory 1.5 \
     --environment-variables PORT=8080
   ```

   Replace the placeholders with your actual values.

6. Get the public URL:

   ```bash
   az container show \
     --resource-group robotApiGroup \
     --name robot-api-container \
     --query ipAddress.fqdn \
     --output tsv
   ```

   Your API will be accessible at `http://[FQDN]:8080`

### Automated Deployment with GitLab CI/CD

The project includes a GitLab CI/CD pipeline that automatically:

1. Builds the Go application
2. Runs unit tests
3. Builds and pushes the Docker image to GitLab Container Registry
4. Deploys to Azure Container Instances

#### Setting up CI/CD Variables

In your GitLab project, go to Settings → CI/CD → Variables and add:

1. **Azure Service Principal Variables**:

   - `AZURE_SP_ID`: Azure Service Principal Application ID
   - `AZURE_SP_PASSWORD`: Azure Service Principal Password
   - `AZURE_TENANT_ID`: Azure Tenant ID

2. **GitLab Registry Variables** (usually auto-populated):
   - `CI_REGISTRY_USER`: GitLab registry username
   - `CI_REGISTRY_PASSWORD`: GitLab registry password

#### Creating Azure Service Principal

To create a Service Principal for automated deployment:

```bash
az ad sp create-for-rbac --name "robot-api-sp" --role contributor --scopes /subscriptions/YOUR_SUBSCRIPTION_ID
```

This will output the credentials you need for the CI/CD variables.

#### Pipeline Execution

Once variables are configured, the pipeline will automatically run on pushes to `main` or `master` branches. You can monitor the deployment in GitLab's CI/CD → Pipelines section.

### Verifying Deployment

1. Check container status:

   ```bash
   az container show --resource-group robotApiGroup --name robot-api-container --query provisioningState
   ```

2. Get logs if needed:

   ```bash
   az container logs --resource-group robotApiGroup --name robot-api-container
   ```

3. Test the API:

   ```bash
   curl http://[YOUR_FQDN]:8080/robot/robot1/status
   ```

### Cleanup

To avoid charges, clean up resources when done:

```bash
# Delete the container
az container delete --resource-group robotApiGroup --name robot-api-container --yes

# Delete the resource group (removes everything)
az group delete --name robotApiGroup --yes
```

### Troubleshooting

Common issues and solutions:

1. **Registry authentication failed**: Verify your PAT has the correct scopes and hasn't expired
2. **Provider not registered**: Wait longer for Microsoft.ContainerInstance registration
3. **DNS name already taken**: Use a more unique DNS label
4. **Container fails to start**: Check container logs with `az container logs`
5. **Port not accessible**: Ensure port 8080 is specified in the container creation command
