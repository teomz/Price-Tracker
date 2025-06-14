# Price Tracker Website

This microservices-based price tracking website collects and monitors product prices across various platforms. The project features a Go-based backend API responsible for web scraping, supported by a suite of services for orchestration, data storage, and processing.

It follows an ELT (Extract, Load, Transform) pipeline: data is extracted using the scraper API, loaded into PostgreSQL, and transformed using DBT for analysis and reporting.

## Tech Stack
- **Go (Golang)**: Backend service and web scraping API
- **Docker**: For containerizing services
- **Airflow**: Scheduling and managing scraping tasks
- **PostgreSQL**: Database for storing scraped price data
- **DBT**: Data quality and transformation tool 
- **MinIO**: Open-source object storage server for image/objects (Simulate S3 server)

---

## **Directory Structure**
```
price-tracker/
│
├── api-service/               # Go API service for web scraping and backend
│   ├── main.go
│   ├── initialization.go 
│   ├── go.mod          
│   ├── minio/
│   │   ├── minio.go          # api gateway
│   │   ├── init.go           # initiliaze minio    
│   │   ├── upload.go         # upload file into minio
│   │   ├── get.go            # get file from minio
│   │   └── delete.go         # delete file in minio
│   ├── postgres/
│   │   ├── SQL               # SQL folder
│   │   ├── postgres.go       # api gateway
│   │   ├── init.go           # initiliaze postgres    
│   │   └── upload.go         # post query into postgres
│   ├── scraper/
│   │   ├── scraper.go        # api gateway
│   │   ├── IST.go            # platform specific scraping
│   │   └── get.go            # get info in json format
│   └── Dockerfile   
│  
│  // Work In Progress
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
- **`GET /minio/getImage`**: retrieve image from minio.
- **`POST /minio/uploadImage`**: upload image into minio.
- **`POST /postgres/uploadInfo`**: upload data in postgres.
- **`GET /scraper/getScrapedInfo`**: get product info in json fomat.


## **Future Enhancements**
- Add authentication to the API.
- Integrate monitoring tools like Prometheus and Grafana.

