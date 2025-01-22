# Scalable Data Scraping Framework with Go, PostgreSQL, and dbt

This repository provides a scalable framework to scrape data from websites, store it in a PostgreSQL database, and transform it using dbt. All services are containerized using Docker for portability and scalability.

---

## **Architecture Overview**

### **Components**
1. **API Service**: A RESTful API built with Go to scrape data and insert it into the database.
2. **Database**: PostgreSQL for storing raw scraped data.
3. **dbt Service**: Handles data transformations and scheduling.
4. **Docker Compose**: Orchestrates all services for easy setup and deployment.

---

## **Getting Started**

### **Prerequisites**
- Docker & Docker Compose
- Go (for local API development)
- dbt Core (for local data transformations)

### **Setup**
1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd <repository-folder>
   ```
2. Start the services:
   ```bash
   docker-compose up --build
   ```
3. Access the services:
   - **API**: `http://localhost:8000`
   - **PostgreSQL**: Hosted internally in Docker on port 5432.

---

## **Directory Structure**
```
project/
├── api/
│   ├── Dockerfile       # Dockerfile for Go API
│   ├── main.go          # Main API logic
├── dbt/
│   ├── Dockerfile       # Dockerfile for dbt service
│   ├── dbt_project.yml  # dbt project configuration
│   ├── models/          # dbt models for transformations
├── docker-compose.yml   # Orchestrates all services
```

---

## **API Service**

### **Endpoints**
- **`GET /scrape`**: Triggers the scraping process and inserts data into the database.
- **`GET /health`**: Health check endpoint.

### **Code Overview**
```go
package main

import (
    "database/sql"
    "fmt"
    "log"
    "net/http"
    "time"

    "github.com/gocolly/colly"
    _ "github.com/lib/pq"
)

const dbConn = "postgresql://user:password@db:5432/scraper_db?sslmode=disable"

func scrapeHandler(w http.ResponseWriter, r *http.Request) {
    db, err := sql.Open("postgres", dbConn)
    if err != nil {
        log.Printf("Database connection error: %v\n", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }
    defer db.Close()

    c := colly.NewCollector()
    var data []string

    c.OnHTML("target-element", func(e *colly.HTMLElement) {
        data = append(data, e.Text)
    })

    err = c.Visit("https://example.com")
    if err != nil {
        log.Printf("Scraping error: %v\n", err)
        http.Error(w, "Failed to scrape data", http.StatusInternalServerError)
        return
    }

    for _, d := range data {
        _, err := db.Exec("INSERT INTO scraped_data (data, scraped_at) VALUES ($1, $2)", d, time.Now())
        if err != nil {
            log.Printf("Database insert error: %v\n", err)
        }
    }

    fmt.Fprintln(w, "Scraping completed")
}

func main() {
    http.HandleFunc("/scrape", scrapeHandler)
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        fmt.Fprintln(w, "OK")
    })

    log.Println("Starting server on :8000")
    log.Fatal(http.ListenAndServe(":8000", nil))
}
```

### **Dockerfile**
```dockerfile
FROM golang:1.20-alpine
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o app .
CMD ["./app"]
EXPOSE 8000
```

---

## **Database**

### **PostgreSQL Schema**
```sql
CREATE TABLE scraped_data (
    id SERIAL PRIMARY KEY,
    data TEXT,
    scraped_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### **Docker Compose Configuration**
```yaml
db:
  image: postgres:15
  environment:
    POSTGRES_USER: user
    POSTGRES_PASSWORD: password
    POSTGRES_DB: scraper_db
  volumes:
    - db_data:/var/lib/postgresql/data
```

---

## **dbt Service**

### **dbt Project Setup**
1. **Define Models**:
   - Staging models for raw scraped data.
   - Transformation models for reporting and aggregation.

2. **Example Model**:
   ```sql
   -- models/clean_data.sql
   SELECT
       id,
       data,
       scraped_at,
       LENGTH(data) AS data_length
   FROM
       {{ ref('scraped_data') }}
   WHERE
       data IS NOT NULL;
   ```

3. **Dockerfile**:
   ```dockerfile
   FROM ghcr.io/dbt-labs/dbt-postgres:1.5.0
   WORKDIR /dbt
   COPY . .
   CMD ["dbt", "run"]
   ```

---

## **Docker Compose**

### **docker-compose.yml**
```yaml
version: '3.8'

services:
  api:
    build: ./api
    ports:
      - "8000:8000"
    environment:
      - DATABASE_URL=postgresql://user:password@db:5432/scraper_db
    depends_on:
      - db

  db:
    image: postgres:15
    environment:
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
      POSTGRES_DB: scraper_db
    volumes:
      - db_data:/var/lib/postgresql/data

  dbt:
    build: ./dbt
    depends_on:
      - db

volumes:
  db_data:
```

---

## **Usage**
1. Start all services:
   ```bash
   docker-compose up --build
   ```
2. Trigger scraping:
   ```bash
   curl http://localhost:8000/scrape
   ```
3. Run dbt transformations:
   ```bash
   docker exec -it <dbt-container-id> dbt run
   ```

---

## **Future Enhancements**
- Add authentication to the API.
- Implement advanced scheduling with dbt Cloud or Airflow.
- Integrate monitoring tools like Prometheus and Grafana.

