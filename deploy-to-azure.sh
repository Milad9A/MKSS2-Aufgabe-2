#!/bin/bash
set -e  # Exit on error

# Load environment variables from .env file
if [ -f .env ]; then
  echo "📄 Loading environment variables from .env file..."
  export $(grep -v '^#' .env | xargs)
else
  echo "❌ .env file not found. Please create one based on .env.template"
  echo "   cp .env.template .env"
  echo "   Then edit .env with your actual credentials"
  exit 1
fi

# Verify required environment variables are set
required_vars=(
  "AZURE_SP_ID" 
  "AZURE_SP_PASSWORD" 
  "AZURE_SUBSCRIPTION_ID" 
  "AZURE_TENANT_ID"
  "CI_REGISTRY_USER"
  "CI_REGISTRY_PASSWORD"
  "CI_PROJECT_PATH"
  "CI_REGISTRY"
)

for var in "${required_vars[@]}"; do
  if [ -z "${!var}" ]; then
    echo "❌ Required environment variable $var is not set in .env file"
    exit 1
  fi
done

# Default Azure configuration - these can stay in the script as they're not sensitive
AZURE_RESOURCE_GROUP="${AZURE_RESOURCE_GROUP:-robotApiGroup}"
AZURE_LOCATION="${AZURE_LOCATION:-westeurope}"
AZURE_CONTAINER_NAME="${AZURE_CONTAINER_NAME:-robot-api-container}"
AZURE_DNS_LABEL="${AZURE_DNS_LABEL:-robot-api-milad9a}"
DOCKER_IMAGE_NAME="${DOCKER_IMAGE_NAME:-robot-api}"
DOCKER_IMAGE_TAG="${DOCKER_IMAGE_TAG:-latest}"

# Set full image name
FULL_IMAGE_NAME="$CI_REGISTRY/$CI_PROJECT_PATH/$DOCKER_IMAGE_NAME:$DOCKER_IMAGE_TAG"

echo "🚀 Starting manual deployment to Azure..."

# Check for required tools
check_required_tools() {
  echo "🔍 Checking for required tools..."
  
  if ! command -v docker &> /dev/null; then
    echo "❌ Docker not found. Please install Docker."
    exit 1
  fi
  
  if ! command -v az &> /dev/null; then
    echo "❌ Azure CLI not found. Please install Azure CLI."
    echo "   Visit: https://docs.microsoft.com/en-us/cli/azure/install-azure-cli"
    exit 1
  fi
  
  if ! command -v openssl &> /dev/null; then
    echo "❌ OpenSSL not found. Please install OpenSSL."
    exit 1
  fi
  
  echo "✅ All required tools are installed."
}

# Login to Docker Registry and Azure
login() {
  echo "🔐 Logging into GitLab Container Registry..."
  echo "$CI_REGISTRY_PASSWORD" | docker login -u "$CI_REGISTRY_USER" --password-stdin "$CI_REGISTRY"
  
  echo "🔐 Logging into Azure..."
  az login --service-principal -u "$AZURE_SP_ID" -p "$AZURE_SP_PASSWORD" --tenant "$AZURE_TENANT_ID"
  az account set --subscription "$AZURE_SUBSCRIPTION_ID"
  
  echo "✅ Authentication completed."
}

# Build and push Docker image
build_and_push() {
  echo "🐳 Building Docker image..."
  docker build -t "$FULL_IMAGE_NAME" .
  
  echo "📤 Pushing image to GitLab Container Registry..."
  docker push "$FULL_IMAGE_NAME"
  
  echo "✅ Image built and pushed successfully."
}

# Prepare Azure infrastructure
prepare_azure() {
  echo "☁️ Preparing Azure infrastructure..."
  
  # Create resource group if it doesn't exist
  az group create --name "$AZURE_RESOURCE_GROUP" --location "$AZURE_LOCATION" || true
  
  # Ensure necessary providers are registered
  az provider register --namespace Microsoft.ContainerInstance || true
  az provider register --namespace Microsoft.Network || true
  
  echo "⏳ Waiting for Azure Container Instance provider registration..."
  for i in {1..30}; do
    STATUS=$(az provider show --namespace Microsoft.ContainerInstance --query registrationState -o tsv)
    echo "Registration status: $STATUS (attempt $i/30)"
    if [ "$STATUS" = "Registered" ]; then
      echo "✅ Provider registered successfully"
      break
    fi
    sleep 10
  done
  
  # Clean up existing container if present
  echo "🧹 Cleaning up existing container..."
  az container delete --resource-group "$AZURE_RESOURCE_GROUP" --name "$AZURE_CONTAINER_NAME" --yes || true
  sleep 15
  
  echo "✅ Azure infrastructure prepared."
}

# Deploy container to Azure
deploy_container() {
  echo "🚀 Deploying to Azure Container Instances..."
  
  az container create \
    --resource-group "$AZURE_RESOURCE_GROUP" \
    --name "$AZURE_CONTAINER_NAME" \
    --image "$FULL_IMAGE_NAME" \
    --dns-name-label "$AZURE_DNS_LABEL" \
    --ports 80 443 8080 \
    --protocol TCP \
    --registry-login-server "$CI_REGISTRY" \
    --registry-username "$CI_REGISTRY_USER" \
    --registry-password "$CI_REGISTRY_PASSWORD" \
    --os-type Linux \
    --cpu 1 \
    --memory 1.5 \
    --environment-variables PORT=8080 GIN_MODE=release ENABLE_HTTPS=true \
    --restart-policy OnFailure
    
  echo "✅ Container deployed successfully."
}

# Set up Application Gateway for HTTPS
setup_app_gateway() {
  echo "🔒 Setting up Application Gateway for HTTPS termination..."
  
  # Create virtual network for Application Gateway
  echo "📡 Creating virtual network..."
  az network vnet create \
    --resource-group "$AZURE_RESOURCE_GROUP" \
    --name robot-api-vnet \
    --address-prefix 10.0.0.0/16 \
    --subnet-name robot-api-subnet \
    --subnet-prefix 10.0.0.0/24 || true
    
  # Create public IP for Application Gateway with unified DNS name
  echo "🌐 Creating public IP..."
  az network public-ip create \
    --resource-group "$AZURE_RESOURCE_GROUP" \
    --name robot-api-gateway-ip \
    --allocation-method Static \
    --sku Standard \
    --dns-name robot-api-milad9a || true
    
  # Get container IP
  CONTAINER_IP=$(az container show \
    --resource-group "$AZURE_RESOURCE_GROUP" \
    --name "$AZURE_CONTAINER_NAME" \
    --query ipAddress.ip -o tsv)
    
  echo "📍 Container IP: $CONTAINER_IP"
  
  # Generate self-signed certificate for testing
  echo "🔐 Creating self-signed certificate for HTTPS testing..."
  openssl req -x509 -newkey rsa:4096 -keyout robot-api-key.pem -out robot-api-cert.pem -days 365 -nodes \
    -subj "/C=DE/ST=Bremen/L=Bremen/O=HSB/OU=MKSS2/CN=robot-api-milad9a.westeurope.cloudapp.azure.com"
    
  # Convert to PFX format (required by Application Gateway)
  openssl pkcs12 -export -out robot-api.pfx -inkey robot-api-key.pem -in robot-api-cert.pem -passout pass:TempPassword123!
  
  # Check if Application Gateway exists
  GATEWAY_EXISTS=$(az network application-gateway list \
    --resource-group "$AZURE_RESOURCE_GROUP" \
    --query "[?name=='robot-api-gateway'].name" -o tsv)
    
  if [ -z "$GATEWAY_EXISTS" ]; then
    echo "🏗️ Creating new Application Gateway..."
    # Create Application Gateway with basic HTTP configuration first
    az network application-gateway create \
      --resource-group "$AZURE_RESOURCE_GROUP" \
      --name robot-api-gateway \
      --location "$AZURE_LOCATION" \
      --vnet-name robot-api-vnet \
      --subnet robot-api-subnet \
      --capacity 1 \
      --sku Standard_v2 \
      --http-settings-cookie-based-affinity Disabled \
      --frontend-port 80 \
      --http-settings-port 8080 \
      --http-settings-protocol Http \
      --public-ip-address robot-api-gateway-ip \
      --servers "$CONTAINER_IP" \
      --priority 1000
      
    echo "✅ Application Gateway created successfully"
  else
    echo "🔄 Application Gateway already exists, updating backend pool..."
    # Update backend pool with new container IP
    az network.application-gateway.address-pool update \
      --gateway-name robot-api-gateway \
      --resource-group "$AZURE_RESOURCE_GROUP" \
      --name appGatewayBackendPool \
      --servers "$CONTAINER_IP" || true
  fi
  
  # Wait for gateway to be ready
  echo "⏳ Waiting for Application Gateway to be ready..."
  sleep 30
  
  # Configure HTTPS
  echo "🔒 Configuring HTTPS..."
  
  # Upload SSL certificate
  az network application-gateway ssl-cert create \
    --gateway-name robot-api-gateway \
    --resource-group "$AZURE_RESOURCE_GROUP" \
    --name robot-api-ssl-cert \
    --cert-file robot-api.pfx \
    --cert-password TempPassword123! 2>/dev/null || echo "Certificate already exists"
    
  # Add HTTPS frontend port
  az network application-gateway frontend-port create \
    --gateway-name robot-api-gateway \
    --resource-group "$AZURE_RESOURCE_GROUP" \
    --name httpsPort \
    --port 443 2>/dev/null || echo "HTTPS port already exists"
    
  # Add HTTPS listener
  echo "Creating HTTPS listener..."
  if ! az network application-gateway http-listener show \
    --gateway-name robot-api-gateway \
    --resource-group "$AZURE_RESOURCE_GROUP" \
    --name httpsListener >/dev/null 2>&1; then
    
    az network.application-gateway.http-listener create \
      --gateway-name robot-api-gateway \
      --resource-group "$AZURE_RESOURCE_GROUP" \
      --name httpsListener \
      --frontend-ip appGatewayFrontendIP \
      --frontend-port httpsPort \
      --ssl-cert robot-api-ssl-cert
    echo "✅ HTTPS listener created"
  else
    echo "HTTPS listener already exists"
  fi
  
  # Add HTTPS routing rule
  echo "Creating HTTPS routing rule..."
  if ! az network.application-gateway.rule show \
    --gateway-name robot-api-gateway \
    --resource-group "$AZURE_RESOURCE_GROUP" \
    --name httpsRule >/dev/null 2>&1; then
    
    az network.application-gateway.rule create \
      --gateway-name robot-api-gateway \
      --resource-group "$AZURE_RESOURCE_GROUP" \
      --name httpsRule \
      --http-listener httpsListener \
      --rule-type Basic \
      --address-pool appGatewayBackendPool \
      --http-settings appGatewayBackendHttpSettings \
      --priority 2000
    echo "✅ HTTPS rule created"
  else
    echo "HTTPS rule already exists"
  fi
  
  # Wait longer for Application Gateway to be fully configured
  echo "⏳ Waiting for Application Gateway HTTPS configuration to propagate..."
  sleep 60
  
  # Get the Application Gateway public IP FQDN
  GATEWAY_FQDN=$(az network public-ip show \
    --resource-group "$AZURE_RESOURCE_GROUP" \
    --name robot-api-gateway-ip \
    --query dnsSettings.fqdn -o tsv) || GATEWAY_FQDN=""
    
  echo "🌐 Unified Domain: $GATEWAY_FQDN"
  
  # Clean up certificate files
  rm -f robot-api-key.pem robot-api-cert.pem robot-api.pfx
  
  echo "✅ HTTPS configuration completed"
}

# Test deployment
test_deployment() {
  echo "🧪 Testing deployment..."
  
  # Get container FQDN
  CONTAINER_FQDN=$(az container show \
    --resource-group "$AZURE_RESOURCE_GROUP" \
    --name "$AZURE_CONTAINER_NAME" \
    --query ipAddress.fqdn -o tsv)
    
  echo "🌐 Container FQDN: $CONTAINER_FQDN"
  
  # Get Gateway FQDN
  GATEWAY_FQDN=$(az network public-ip show \
    --resource-group "$AZURE_RESOURCE_GROUP" \
    --name robot-api-gateway-ip \
    --query dnsSettings.fqdn -o tsv) || GATEWAY_FQDN=""
    
  if [ -n "$GATEWAY_FQDN" ]; then
    echo "🌐 Unified Domain: $GATEWAY_FQDN"
    echo "🌐 HTTP endpoint: http://$GATEWAY_FQDN"
    echo "🌐 HTTPS endpoint: https://$GATEWAY_FQDN"
    
    # Test HTTP through Application Gateway
    echo "Testing HTTP through Application Gateway..."
    curl -f --connect-timeout 10 --max-time 30 "http://$GATEWAY_FQDN/health" || echo "⚠️ Gateway HTTP health check failed"
    
    # Test HTTPS through Application Gateway
    echo "Testing HTTPS through Application Gateway..."
    curl -k -f --connect-timeout 15 --max-time 30 "https://$GATEWAY_FQDN/health" || echo "⚠️ HTTPS endpoint failed"
  fi
  
  # Test direct container access
  echo "Testing direct container access..."
  curl -f --connect-timeout 10 --max-time 30 "http://$CONTAINER_FQDN:8080/health" || echo "⚠️ Direct container health check failed"
  
  echo "✅ Deployment testing completed"
}

# Display deployment summary
deployment_summary() {
  # Get container FQDN
  CONTAINER_FQDN=$(az container show \
    --resource-group "$AZURE_RESOURCE_GROUP" \
    --name "$AZURE_CONTAINER_NAME" \
    --query ipAddress.fqdn -o tsv)
    
  # Get Gateway FQDN
  GATEWAY_FQDN=$(az network public-ip show \
    --resource-group "$AZURE_RESOURCE_GROUP" \
    --name robot-api-gateway-ip \
    --query dnsSettings.fqdn -o tsv) || GATEWAY_FQDN=""
  
  echo ""
  echo "🎉 Deployment Summary"
  echo "═══════════════════════════════════════════"
  
  if [ -n "$GATEWAY_FQDN" ]; then
    echo "🌐 Unified Domain: $GATEWAY_FQDN"
    echo "🔓 HTTP URL: http://$GATEWAY_FQDN"
    echo "🔒 HTTPS URL: https://$GATEWAY_FQDN"
    echo "⚠️  Using self-signed certificate (browser will show warning)"
  fi
  
  echo "🐳 Container FQDN: $CONTAINER_FQDN"
  echo "📡 Direct Container URL: http://$CONTAINER_FQDN:8080"
  echo "🏃 Test API: http://$CONTAINER_FQDN:8080/health"
  echo "═══════════════════════════════════════════"
}

# Main deployment process
main() {
  check_required_tools
  login
  build_and_push
  prepare_azure
  deploy_container
  setup_app_gateway
  test_deployment
  deployment_summary
}

# Run deployment
main
