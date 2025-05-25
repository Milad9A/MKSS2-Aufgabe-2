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

### Pushing to GitLab Container Registry

1. Build the Docker image locally:

   ```
   docker build -t robot-api .
   ```

2. Tag the image according to GitLab's naming scheme:

   ```
   docker tag robot-api registry.gitlab.com/hsbremen/mkss2/sose-2025/labor/PROJECT_NAME/robot-api:latest
   ```

   Replace `PROJECT_NAME` with your GitLab project name.

3. Create a Personal Access Token (PAT) in GitLab:

   - Go to GitLab → Settings → Access Tokens
   - Create a new token with `read_registry` and `write_registry` scopes
   - Save the token securely

4. Login to the GitLab Container Registry:

   ```
   docker login registry.gitlab.com
   ```

   Use your GitLab username and the PAT as password.

5. Push the image to the GitLab Container Registry:
   ```
   docker push registry.gitlab.com/hsbremen/mkss2/sose-2025/labor/PROJECT_NAME/robot-api:latest
   ```

### Deploying to Azure Container Instances

1. Login to Azure CLI:

   ```
   az login
   ```

2. Create a resource group if you don't have one:

   ```
   az group create --name robotApiGroup --location westeurope
   ```

3. Register the Microsoft.ContainerInstance provider:

   ```
   az provider register --namespace Microsoft.ContainerInstance
   ```

   Wait a few moments for the registration to complete.

4. Deploy the container from GitLab Container Registry:

   ```
   az container create \
     --resource-group robotApiGroup \
     --name robot-api-container \
     --image registry.gitlab.com/hsbremen/mkss2/sose-2025/labor/PROJECT_NAME/robot-api:latest \
     --dns-name-label robot-api-YOUR_USERNAME \
     --ports 8080 \
     --registry-username YOUR_GITLAB_USERNAME \
     --registry-password YOUR_GITLAB_PAT \
     --os-type Linux \
     --cpu 1 \
     --memory 1.5
   ```

   Replace:

   - `PROJECT_NAME` with your GitLab project name
   - `YOUR_USERNAME` with a unique name (this will be part of your DNS name)
   - `YOUR_GITLAB_USERNAME` with your GitLab username
   - `YOUR_GITLAB_PAT` with your GitLab Personal Access Token

5. Check deployment status:

   ```
   az container show \
     --resource-group robotApiGroup \
     --name robot-api-container \
     --query "{FQDN:ipAddress.fqdn,ProvisioningState:provisioningState}" \
     --out table
   ```

6. Your API is now accessible at:
   ```
   http://robot-api-YOUR_USERNAME.westeurope.azurecontainer.io:8080
   ```

### Cleanup

To remove your deployed container and avoid unnecessary costs:

```
az container delete --resource-group robotApiGroup --name robot-api-container --yes
```

To delete the entire resource group:

```
az group delete --name robotApiGroup --yes
```
