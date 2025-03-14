# 📘 CHANGELOG

This document monitors the development of the Authentication Service API. The Auth service API will also support other processes within the Robot Lab which require authentication for future projects.

All notable changes to this project will be documented in this file.  
This project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html) and follows the [Keep a Changelog](https://keepachangelog.com/en/1.0.0/) format.


## [v1.0.2] - 14-03-2025

### ✨ Added
- [x] Added user role into the JWT token to prevent extra API calls to find the user's role type
- [x] Added `.env.example` file as a guide template
- [x] Added `DEPLOYMENT.md` for guidance of deploying on custom server

### Changed
- [x] Contents of the `settings.yaml` file and removed the sensative data 
- [x] The way API from the `./cmd/api/main.go` read those variables from the `.env` file
- [x] Changed the `run.sh` script to start the app without running the `generate-env.sh` script

### Removed
- [x] Removed the `generate-env.sh` script

---

## [v1.0.1] - 11-02-2025

### ✨ Added
- [x] A new field in the User model for the organization inside `/internal/models/User.go`

---

## [v1.0.0] - 11-02-2025

### ✨ Added
- [x] The Golang based Auth service including basic user management endpoints.

---