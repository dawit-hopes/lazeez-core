# Lazeez Core

**Lazeez Core** is the backend service powering **Lazeez**, a digital ordering platform built for restaurants and businesses in Ethiopia.

The backend is designed to provide a reliable foundation for managing restaurants, menus, customers, orders, and the business logic behind the Lazeez platform.

## 🚀 About Lazeez

Lazeez helps restaurants move from traditional, manual ordering workflows to a simple digital ordering experience.

Customers can discover a restaurant's menu, browse available items, and place orders digitally, while restaurants can manage their menus and incoming orders from a centralized system.

The platform is built with the realities of the Ethiopian market in mind, with an emphasis on simplicity, affordability, and digital payments.

## 🏗️ Backend

Lazeez Core is built using:

* **Go (Golang)** for the backend
* **PostgreSQL** for persistent data storage
* **REST APIs** for communication with client applications
* **Docker** for containerized development and deployment

Go was chosen for its performance, simplicity, strong concurrency model, and suitability for building scalable backend services.

PostgreSQL provides the relational foundation for Lazeez, handling transactional data such as restaurants, menus, products, customers, and orders.

## 📦 Core Capabilities

The backend provides the foundation for:

* 🏪 Restaurant management
* 📋 Digital menu management
* 🍔 Product and menu-item management
* 👤 Customer management
* 🛒 Order management
* 💳 Payment-related workflows
* 🔐 Authentication and authorization
* 📊 Business and transaction data
* 🔄 Order lifecycle management

## 🧱 Architecture

At a high level, the system follows a client-server architecture:

```text
                    Lazeez Clients
                  Web / Mobile / QR
                         │
                         ▼
                 ┌───────────────┐
                 │  Lazeez Core  │
                 │   Go Backend  │
                 └───────┬───────┘
                         │
                         ▼
                 ┌───────────────┐
                 │  PostgreSQL   │
                 │    Database   │
                 └───────────────┘
```

The backend contains the business logic and exposes APIs consumed by Lazeez's client applications.

## 🛠️ Technology Stack

| Technology | Role                |
| ---------- | ------------------- |
| Go         | Backend development |
| PostgreSQL | Database            |
| REST       | API communication   |
| Docker     | Containerization    |

## 🎯 Project Goals

Lazeez is built around a few simple principles:

**Simple for businesses.**
Restaurants should be able to adopt digital ordering without having to understand complicated technology.

**Simple for customers.**
Customers should be able to access a menu and order with as little friction as possible.

**Built for Ethiopia.**
The platform is designed with Ethiopian restaurants, businesses, payment ecosystems, and operating environments in mind.

**Designed to scale.**
The backend is built on technologies capable of supporting the platform as the number of restaurants, customers, and orders grows.

## 🚧 Project Status

Lazeez Core is actively under development as the backend foundation of the Lazeez platform.

Future development includes expanding restaurant management capabilities, payment integrations, notifications, analytics, and additional tools for restaurant operators.

## 📄 License

See the repository's `LICENSE` file for licensing information.
