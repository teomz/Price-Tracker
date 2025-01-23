# Price Tracker Website

This is a microservices-based price tracking website that collects and displays product prices. The project includes a Go backend API for web scraping, a TypeScript-based frontend, and a set of supporting services for scheduling, storing, and processing the data.

## Tech Stack
- **Go (Golang)**: Backend service and web scraping API
- **TypeScript (React)**: Frontend application
- **Docker**: For containerizing services
- **Airflow**: Scheduling and managing scraping tasks
- **PostgreSQL**: Database for storing scraped price data
- **DBT (optional)**: Data quality and transformation tool (to be implemented later)

## Prerequisites
- Docker and Docker Compose
- Go 1.18+ (for API service development)
- Node.js 16+ (for frontend development)
- Python 3.7+ (for Airflow)


---

## **Directory Structure**
```
price-tracker/
│
├── api-service/               # Go API service for web scraping and backend
│   ├── main.go
│   ├── go.mod          
│   ├── minio/
│   │   ├── main.go
│   │   ├── handlers/
│   │   ├── go.mod
│   │   ├── upload.go
│   │   ├── fetch.go
│   │   └── delete.go
│   └── Dockerfile
│
├── frontend/                  # TypeScript frontend (React or similar)
│   ├── public/
│   ├── src/
│   │   ├── components/
│   │   ├── App.tsx
│   │   └── index.tsx
│   ├── package.json
│   └── tsconfig.json
│
├── airflow/                   # Airflow for scheduling scraping tasks
│   ├── dags/
│   └── Dockerfile
│
├── dbt/                       # DBT for data transformation and quality checks (optional)
│   └── dbt_project.yaml
│
├── docker-compose.yaml        # Docker compose to start all services
├── Taskfile.yaml              # Automate common tasks
└── README.md                  # Project documentation

```

## **API Service**

### **Endpoints**
- **`GET /scrape`**: Triggers the scraping process and inserts data into the database.
- **`GET /health`**: Health check endpoint.



## **Future Enhancements**
- Add authentication to the API.
- Integrate monitoring tools like Prometheus and Grafana.

