from airflow.decorators import dag, task
from airflow.utils.dates import days_ago
import logging
import requests
from dotenv import load_dotenv
from typing import List
from datetime import datetime, timedelta
import os
import sys

sys.path.append('/opt/airflow')

from models.model import Omnibus



# Load environment variables
load_dotenv()

# Scraper API
TASKUSER = os.getenv("AIRFLOW_USER")


# Set up logging
logging.basicConfig(level=logging.INFO)

@dag(schedule_interval="@weekly", start_date=days_ago(1), catchup=False)
def weekly_update_with_new_release():
    
    @task()
    def fetch_info_data() -> List[Omnibus]:
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
            data = response.json()['data']
            return data

        except requests.exceptions.RequestException as e:
            logging.error("Error fetching data: %s", response.json())

        logging.info("InfoList Lenght:", len(infoList))

    @task()
    def clean_duplicates(infoList: List[Omnibus]) -> List[Omnibus]:
        logging.info("Removing the duplicates")

        # Format the date in yyyy-mm-dd format
        formatted_date = (datetime.now() - timedelta(weeks=1)).strftime("%Y-%m-%d")

        params = {
            "TaskUser": TASKUSER,
            "Date": formatted_date,
        }

        try:
            response = requests.get("http://api-service:8080/api/v1/postgresql/getInfoByDate", params=params)
            response.raise_for_status()  # Raises HTTPError for 4xx/5xx codes
            # Process the response data
            duplicatesList = response.json()

        except requests.exceptions.RequestException as e:
            logging.error("Error fetching data: %s", duplicatesList)

        
        try:
            infoList = [o for o in infoList if not any(val in o.values() for val in duplicatesList)]
        except Exception as e:
            logging.info("No duplicates")
        
        logging.info("InfoList Lenght:", len(infoList))

        return infoList
    
    
    @task()
    def upload_info(infoList: List[Omnibus]):
            
        params = {
            "TaskUser": TASKUSER,
        }

        try:
            response = requests.post("http://api-service:8080/api/v1/postgresql/uploadInfo", params=params, json=infoList)
            response.raise_for_status()  # Raises HTTPError for 4xx/5xx codes
            # Process the response data
            insertedList = response.json()
            logging.info(insertedList)

        except requests.exceptions.RequestException as e:
            logging.error("Error fetching data: %s", insertedList)

            

    infoList = fetch_info_data()
    infoList = clean_duplicates(infoList)
    upload_info(infoList)



# Initialize the DAG
pipeline = weekly_update_with_new_release()
