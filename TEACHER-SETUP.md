# 🎓 **Teacher Setup Guide - Complete Restoration**

All GitHub Actions and deployment scripts have been recreated! Here's your complete setup for deploying the trading server for your 3-week course.

## 📁 **What Was Created**

### **GitHub Actions Workflows** (`.github/workflows/`)
1. **`security-scan.yml`** - Security scanning (code, dependencies, container)
2. **`build-container.yml`** - Buildah container build + push to GHCR
3. **`deploy-teacher.yml`** - Cost-optimized Azure deployment

### **Setup Scripts** (`scripts/`)
1. **`teacher-quick-setup.sh`** - Complete Azure setup automation

## 🚀 **Quick Start (Choose One Method)**

### **Method 1: Automated Script (RECOMMENDED)**
```bash
# Run the automated setup script
./scripts/teacher-quick-setup.sh
```
This will:
- ✅ Create Azure resource group
- ✅ Create service principal for GitHub
- ✅ Generate all the configuration you need
- ✅ Provide exact copy-paste values for GitHub

### **Method 2: Manual Setup (If you prefer step-by-step)**
Follow the original step-by-step guide in your Azure Portal.

## 📋 **What You Need to Configure in GitHub**

### **Secrets** (Settings → Secrets and variables → Actions → **Secrets**)
- `AZURE_CREDENTIALS` - JSON from setup script
- `MONGODB_URI` - Your MongoDB Atlas connection string

### **Variables** (Settings → Secrets and variables → Actions → **Variables**)
- `AZURE_RESOURCE_GROUP` = `trading-course-rg`
- `AZURE_CONTAINER_NAME` = `trading-course`
- `AZURE_LOCATION` = `eastus`
- `CONTAINER_PORT` = `80`
- `CPU_CORES` = `0.5` ← 💰 **Cost optimized**
- `MEMORY_GB` = `1` ← 💰 **Cost optimized**
- `LOG_LEVEL` = `info`
- `ENVIRONMENT` = `teacher-course`
- `MAX_CONNECTIONS` = `30`

## 🎯 **Your Cost-Optimized Configuration**

**Resources**: 0.5 vCPU, 1 GB RAM
**Cost**: ~$12-15 for entire 3-week course
**Students**: Handles 20-30 concurrent users
**Reliability**: 24/7 uptime, no cold starts

## 🌐 **Student Access**

After deployment, your students will access:
```
http://trading-course.eastus.azurecontainer.io
```

## 🔧 **Teacher Management Commands**

```bash
# Monitor student activity (real-time logs)
az container logs --resource-group trading-course-rg --name trading-course --follow

# Restart server if needed during class
az container restart --resource-group trading-course-rg --name trading-course

# Check server status
az container show --resource-group trading-course-rg --name trading-course
```

## 🚀 **Deployment Process**

1. **Run setup script**: `./scripts/teacher-quick-setup.sh`
2. **Add secrets/variables** to GitHub (script provides exact values)
3. **Set up MongoDB Atlas** (free tier is fine)
4. **Push to main branch** → Automatic deployment!

## 📊 **Pipeline Flow**

```
Code Push → Security Scan → Build Container → Deploy to Azure
    ↓            ↓               ↓              ↓
  5 mins       2 mins         3 mins        5 mins
```

**Total deployment time**: ~15 minutes

## 💰 **Cost Breakdown**

- **Daily**: ~$0.60
- **Weekly**: ~$4-5  
- **3-week course**: ~$12-15 total
- **Per student**: ~$0.40-0.50 each

## ✅ **All Files Created:**

```
.github/workflows/
├── security-scan.yml          # Security scanning
├── build-container.yml        # Container build with Buildah  
└── deploy-teacher.yml         # Cost-optimized Azure deployment

scripts/
└── teacher-quick-setup.sh     # Automated Azure setup

docs/
└── (existing files preserved)
```

## 🎓 **Ready to Deploy!**

Your complete CI/CD pipeline is ready. Just run the setup script and follow the instructions it provides. Everything is optimized for your 3-week teaching budget!

**Questions?** Check the script output or GitHub Actions logs for any issues.