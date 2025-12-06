# 🔥 The Forge

**A High-Performance Remote Code Execution (RCE) Engine.**

Designed to safely compile and execute untrusted user code in a sandboxed environment. Built with **Golang**, **Docker**, and **Systems Engineering** principles.

---

## 🗺️ The Roadmap

### 👶 Phase 1: The Foundation (✅ Completed)
* **Level 1: The Core Engine**
    * Implemented `os/exec` pipeline to compile/run C++.
    * Handled `stdout` and `stderr` stream merging.
    * Built local file I/O management.

### 🚧 Phase 2: The Engineering (✅ Completed)
* **Level 2: The API Gateway**
    * Built RESTful API using `net/http`.
    * Implemented JSON Marshaling/Unmarshaling.
    * Added strict Method Guards (POST only).

### 🐳 Phase 3: The Sandbox (✅ Completed)
* **Level 3: Docker Isolation**
    * Integrated Docker SDK Client.
    * Implemented Container Lifecycle (Create -> Start -> Wait -> Destroy).
    * Built "Teleporter" to stream in-memory C++ code using TAR archives.
    * Connected HTTP API to Docker Engine (Dependency Injection).

### ⚡ Phase 4: High-Performance (Next)
* **Level 4: Concurrency Engine**
    * Implement Worker Pools using Goroutines.
    * Handle buffered channels for job queuing.
    * Prevent server crashes under load.

### 🚀 Phase 5: Production Optimization
* **Level 5: The Memory (Redis)**
    * Implement Caching layer.
    * Hash source code to return cached results instantly.

### 📊 Phase 6: Observability
* **Level 6: The Watchtower**
    * Integrate Prometheus metrics.
    * Build Grafana dashboards for "Compilations/Sec".

### 👑 Phase 7: The Endgame
* **Level 7: Security Hardening**
    * Implement gVisor / Firecracker for kernel-level isolation.
* **Level 8: The Cluster (Kubernetes)**
    * Deploy to K8s for auto-scaling capabilities.

---

## 🛠 Tech Stack
* **Language:** Go (Golang)
* **Architecture:** REST Microservice
* **Infrastructure:** Docker, Linux, WSL2