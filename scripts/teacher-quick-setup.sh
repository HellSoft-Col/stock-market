#!/bin/bash

# Quick setup script for teachers - creates Azure resources and provides GitHub configuration
# Optimized for 3-week courses with cost savings

set -e

echo "🎓 Teacher Quick Setup for Stock Exchange Course"
echo "💰 Cost-optimized for 3-week duration (~$12-15 total)"
echo ""

# Check if Azure CLI is installed
if ! command -v az &> /dev/null; then
    echo "❌ Azure CLI is not installed. Please install it first:"
    echo "   https://docs.microsoft.com/en-us/cli/azure/install-azure-cli"
    exit 1
fi

# Check if user is logged in
if ! az account show &> /dev/null; then
    echo "🔐 You need to login to Azure first..."
    az login
    if [ $? -ne 0 ]; then
        echo "❌ Azure login failed"
        exit 1
    fi
fi

# Get subscription info
SUBSCRIPTION_ID=$(az account show --query id -o tsv)
SUBSCRIPTION_NAME=$(az account show --query name -o tsv)
TENANT_ID=$(az account show --query tenantId -o tsv)

echo "✅ Logged in to Azure:"
echo "   Subscription: $SUBSCRIPTION_NAME"
echo "   Subscription ID: $SUBSCRIPTION_ID"
echo ""

# Prompt for resource group name
read -p "Enter resource group name (or press Enter for 'trading-course-rg'): " RESOURCE_GROUP
RESOURCE_GROUP=${RESOURCE_GROUP:-"trading-course-rg"}

# Prompt for container name
read -p "Enter container name (or press Enter for 'trading-course'): " CONTAINER_NAME
CONTAINER_NAME=${CONTAINER_NAME:-"trading-course"}

# Prompt for location
echo "Recommended locations for cost optimization:"
echo "1. East US (cheapest)"
echo "2. West Europe" 
echo "3. Southeast Asia"
read -p "Enter location (or press Enter for 'eastus'): " LOCATION
LOCATION=${LOCATION:-"eastus"}

echo ""
echo "📋 Configuration Summary:"
echo "   Resource Group: $RESOURCE_GROUP"
echo "   Container Name: $CONTAINER_NAME"
echo "   Location: $LOCATION"
echo "   Estimated Cost: ~$12-15 for 3 weeks"
echo ""

read -p "Continue with this configuration? (y/n): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Setup cancelled."
    exit 1
fi

echo ""
echo "🏗️ Creating Azure resources..."

# Create resource group
echo "📁 Creating resource group..."
az group create \
    --name "$RESOURCE_GROUP" \
    --location "$LOCATION" \
    --output none

if [ $? -eq 0 ]; then
    echo "✅ Resource group '$RESOURCE_GROUP' created"
else
    echo "❌ Failed to create resource group"
    exit 1
fi

# Create service principal for GitHub Actions
echo "🤖 Creating service principal for GitHub Actions..."
SP_NAME="trading-course-github-$(date +%s)"

SP_OUTPUT=$(az ad sp create-for-rbac \
    --name "$SP_NAME" \
    --role "Contributor" \
    --scopes "/subscriptions/$SUBSCRIPTION_ID/resourceGroups/$RESOURCE_GROUP" \
    --output json)

if [ $? -eq 0 ]; then
    echo "✅ Service principal '$SP_NAME' created"
else
    echo "❌ Failed to create service principal"
    exit 1
fi

# Extract service principal details
CLIENT_ID=$(echo $SP_OUTPUT | jq -r '.appId')
CLIENT_SECRET=$(echo $SP_OUTPUT | jq -r '.password')

# Create Azure credentials JSON for GitHub
AZURE_CREDENTIALS=$(cat <<EOF
{
  "clientId": "$CLIENT_ID",
  "clientSecret": "$CLIENT_SECRET",
  "subscriptionId": "$SUBSCRIPTION_ID",
  "tenantId": "$TENANT_ID"
}
EOF
)

echo ""
echo "🎉 Azure setup complete!"
echo ""
echo "=================================================="
echo "📋 GITHUB REPOSITORY CONFIGURATION"
echo "=================================================="
echo ""
echo "🔐 ADD THESE SECRETS (Settings → Secrets and variables → Actions → Secrets):"
echo ""
echo "Secret 1: AZURE_CREDENTIALS"
echo "Copy this JSON (including the curly braces):"
echo ""
echo "$AZURE_CREDENTIALS"
echo ""
echo "Secret 2: MONGODB_URI"
echo "Add your MongoDB Atlas connection string here"
echo "(Format: mongodb+srv://user:pass@cluster.mongodb.net/database)"
echo ""
echo "=================================================="
echo "📝 ADD THESE VARIABLES (Settings → Secrets and variables → Actions → Variables):"
echo "=================================================="
echo ""
cat << EOF
Variable Name: AZURE_RESOURCE_GROUP
Value: $RESOURCE_GROUP

Variable Name: AZURE_CONTAINER_NAME  
Value: $CONTAINER_NAME

Variable Name: AZURE_LOCATION
Value: $LOCATION

Variable Name: CONTAINER_PORT
Value: 80

Variable Name: CPU_CORES
Value: 0.5

Variable Name: MEMORY_GB
Value: 1

Variable Name: LOG_LEVEL
Value: info

Variable Name: ENVIRONMENT
Value: teacher-course

Variable Name: MAX_CONNECTIONS
Value: 30
EOF

echo ""
echo "=================================================="
echo "🌐 STUDENT ACCESS INFORMATION"
echo "=================================================="
echo ""
echo "After deployment, your students will access:"
echo "http://$CONTAINER_NAME.$LOCATION.azurecontainer.io"
echo ""
echo "=================================================="
echo "💰 COST TRACKING"
echo "=================================================="
echo ""
echo "Daily cost: ~$0.60"
echo "Weekly cost: ~$4-5"
echo "3-week total: ~$12-15"
echo "Per student (30 students): ~$0.40-0.50 each"
echo ""
echo "Monitor costs: Azure Portal → Cost Management + Billing"
echo ""
echo "=================================================="
echo "🚀 NEXT STEPS"
echo "=================================================="
echo ""
echo "1. Add the SECRETS and VARIABLES above to your GitHub repository"
echo "2. Set up MongoDB Atlas (if not done already)"
echo "3. Push code to main branch to trigger deployment"
echo "4. Your trading server will be ready for students!"
echo ""
echo "🎓 Happy teaching!"
echo ""

# Save configuration to file for reference
CONFIG_FILE="teacher-azure-config.txt"
cat > "$CONFIG_FILE" << EOF
# Teacher Azure Configuration
# Generated on $(date)

Resource Group: $RESOURCE_GROUP
Container Name: $CONTAINER_NAME
Location: $LOCATION
Student URL: http://$CONTAINER_NAME.$LOCATION.azurecontainer.io

# Azure Credentials (for GitHub secrets)
$AZURE_CREDENTIALS

# Cost Estimate
Daily: ~$0.60
3-week total: ~$12-15
EOF

echo "💾 Configuration saved to: $CONFIG_FILE"
echo ""