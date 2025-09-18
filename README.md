# Home Server Controller

> Transform your local network into a powerful personal cloud ecosystem

## What is Home Server Controller?

Home Server Controller is a sleek, self-hosted solution that turns any computer into your personal cloud server. Built for privacy-conscious users who want complete control over their data, it provides secure file storage, user management, and multimedia streaming capabilities—all within your local network.

## Key Features

### **Secure Multi-User Environment**
- Individual user folders with dedicated permissions
- Admin dashboard for user management
- Planned end-to-end encryption for maximum privacy

### **Dual-Interface Architecture**
- **Web Interface**: Intuitive setup and administration panel (Port 8080)
- **TCP Server**: High-performance backend with dynamic port allocation
- Automatic conflict resolution—no port collisions with other applications

### **Smart Device Integration**
- RESTful API for easy third-party app development
- Built for seamless mobile and desktop connectivity
- Future-ready architecture for media streaming applications

### **Coming Soon**
- **Instant Messaging**: Chat with friends and family on your network
- **Media Center Integration**: Control your TV and media players directly from your phone
- **Storage Quotas**: Customizable data limits per user

## Perfect For

- Families wanting their own private cloud
- Small businesses needing secure file sharing
- Media enthusiasts planning home entertainment systems
- Privacy advocates seeking alternatives to big tech solutions
- Developers building local network applications

## Quick Start

### Prerequisites
- [Go](https://golang.org/dl/) installed on your system

### Installation
```bash
# Clone and navigate to the project
cd home-server-controller

# Build the application
go build ./src

# Run the server
./home-server-controller
```

### First-Time Setup
1. **Automatic Admin Creation**: On first launch, the server creates a default admin user with a randomly generated password
2. **Access Web Interface**: Navigate to `http://localhost:8080` in your browser
3. **Login**: Use the generated credentials (displayed in console or found in `/res/config_files/admin_credentials.txt`)
4. **Secure Your Setup**: Change the default password immediately after login

> ⚠️ **Security Tip**: Delete `/res/config_files/admin_credentials.txt` after setup to prevent unauthorized access

## API Integration

### Get TCP Server Details
```bash
POST /WebServerController/details
```
This endpoint provides current TCP server connection information for client applications.

## Roadmap

- [ ] **Phase 1**: User encryption and storage quotas
- [ ] **Phase 2**: Real-time messaging system  
- [ ] **Phase 3**: Media streaming and TV integration
- [ ] **Phase 4**: Mobile companion app
- [ ] **Phase 5**: Plugin ecosystem for extensibility

## 🤝 Contributing

We welcome contributions! Whether it's bug reports, feature requests, or code contributions, your input helps make Home Server Controller better for everyone.
