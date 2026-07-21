<div align="center">
  <img src="https://img.shields.io/badge/Mautrade-1a1a1a?style=for-the-badge&logo=mautrade&logoColor=FF6B00" alt="Mautrade Logo" />
  
  # Mautrade Trading Platform

  <p align="center">
    <img src="https://img.shields.io/badge/Status-Active-success?style=flat-square" alt="Status" />
    <img src="https://img.shields.io/badge/License-MIT-blue?style=flat-square" alt="License" />
    <img src="https://img.shields.io/badge/Version-1.0.0-orange?style=flat-square" alt="Version" />
  </p>
</div>

---

## Introduction

Mautrade is a high-performance, modern trading platform architecture designed for speed, security, and scalability. It provides real-time market execution, secure wallet management, and comprehensive administrative oversight. 

The platform is strictly decoupled into a robust microservices-ready backend built with Go and a highly responsive server-side rendered frontend built with Nuxt.

---

## Technology Stack

The architecture utilizes modern technologies ensuring type-safety, rapid execution, and visual excellence.

### Frontend Technologies

<p>
  <img src="https://img.shields.io/badge/Nuxt-4.4.8-00C58E?style=flat-square&logo=nuxt.js&logoColor=white" alt="Nuxt" />
  <img src="https://img.shields.io/badge/Vue.js-3.3.7-4FC08D?style=flat-square&logo=vuedotjs&logoColor=white" alt="Vue" />
  <img src="https://img.shields.io/badge/Tailwind_CSS-4.3.2-38B2AC?style=flat-square&logo=tailwind-css&logoColor=white" alt="Tailwind" />
  <img src="https://img.shields.io/badge/TypeScript-6.0.3-3178C6?style=flat-square&logo=typescript&logoColor=white" alt="TypeScript" />
  <img src="https://img.shields.io/badge/PNPM-11.13.1-F69220?style=flat-square&logo=pnpm&logoColor=white" alt="PNPM" />
</p>

The frontend is built for extreme performance and developer experience:

*   **Nuxt (v4.4.8)**: Acts as the core framework, providing Server-Side Rendering (SSR) for SEO and fast initial page loads, alongside file-based routing and auto-imports.
*   **TailwindCSS (v4.3.2)**: A utility-first CSS framework used for rapid UI styling, enforcing a strict dark-mode-first aesthetic with custom charcoal and orange accents.
*   **@nuxt/ui (v4.10.0)**: Provides a set of fully styled and customizable UI components built with Tailwind CSS and Headless UI, ensuring design consistency across the dashboard.
*   **Vue-Chartjs (v5.3.4) & Chart.js (v4.5.1)**: Used for rendering high-performance, interactive data visualizations for market execution and trading charts.
*   **TypeScript (v6.0.3)**: Enforces strict static typing across all components and API interactions, preventing runtime errors.
*   **ESLint (v10.7.0)**: Maintains strict code formatting and quality standards.
*   **Package Manager**: Managed strictly via **pnpm (v11.13.1)** for fast, disk-efficient dependency resolution.

### Backend Technologies

<p>
  <img src="https://img.shields.io/badge/Go-1.24.4-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go" />
  <img src="https://img.shields.io/badge/PostgreSQL-pgx_v5.8.0-4169E1?style=flat-square&logo=postgresql&logoColor=white" alt="PostgreSQL" />
  <img src="https://img.shields.io/badge/NATS-1.49.0-27AAE1?style=flat-square&logo=nats&logoColor=white" alt="NATS" />
</p>

The backend infrastructure handles all business logic, data persistence, and real-time events:

*   **Go (v1.24.4)**: The core language used for the backend API, chosen for its unparalleled concurrency model (goroutines) and high-performance execution speeds.
*   **pgx (v5.8.0)**: A pure Go PostgreSQL driver and toolkit providing high-performance database interactions and secure query execution.
*   **nats.go (v1.49.0)**: The NATS messaging client utilized for ultra-low latency, distributed communication between microservices and background workers.
*   **x/crypto (v0.47.0)**: Used for cryptographically secure hashing (bcrypt) and cryptographic functions to secure user credentials and wallet operations.
*   **qdecimal (v1.0.3) & quuid**: Specialized packages for handling high-precision financial arithmetic and generating universally unique identifiers without floating-point inaccuracies.

---

## Deployment & CI/CD

<p>
  <img src="https://img.shields.io/badge/Docker-Compose-2496ED?style=flat-square&logo=docker&logoColor=white" alt="Docker" />
  <img src="https://img.shields.io/badge/GitHub_Actions-CI%2FCD-2088FF?style=flat-square&logo=github-actions&logoColor=white" alt="GitHub Actions" />
  <img src="https://img.shields.io/badge/Traefik-v3.4-24A1C1?style=flat-square&logo=traefik-proxy&logoColor=white" alt="Traefik" />
</p>

The application utilizes a fully automated CI/CD pipeline defined in GitHub Actions. Pushes to the main branch automatically trigger:

1.  **Validation**: Syntax checking and Docker Compose configuration validation.
2.  **Deployment**: Secure SSH execution into the production VPS.
3.  **Smart Reloading**: The custom `deploy-vps.sh` script detects changes and selectively rebuilds the Go API, Nuxt Frontend, or underlying infrastructure.
4.  **Reverse Proxying**: Traefik handles dynamic routing and automatic SSL/TLS certificate generation via Let's Encrypt.
