from airflow.decorators import dag, task
from airflow.utils.dates import days_ago
import logging
import requests
import os
from dotenv import load_dotenv

# Load environment variables
load_dotenv()

# Scraper API
TASKUSER = os.getenv("AIRFLOW_USER")


# Set up logging
logging.basicConfig(level=logging.INFO)

@dag(schedule_interval="@weekly", start_date=days_ago(1), catchup=False)
def weekly_update_with_new_release():
    
    @task()
    def fetch_info_data():
        logging.info("Fetching new release info from the scraper API...")

        # Parameters for the GET request
        params = {
            "TaskUser": TASKUSER,
            "URL": "https://www.instocktrades.com/newreleases",
        }

        # Send GET request with parameters
        try:
            response = requests.get("http://api-service:8080/api/v1/scraper/getScrapedInfo", params=params)
            response.raise_for_status()  # Raises HTTPError for 4xx/5xx codes
            # Process the response data
            data = response.json()

        except requests.exceptions.RequestException as e:
            logging.error("Error fetching data: %s", data)

    fetch_info_data()

# Initialize the DAG
pipeline = weekly_update_with_new_release()
