from airflow.decorators import dag, task
from airflow.utils.dates import days_ago
from airflow.exceptions import AirflowFailException
import logging
import requests
from dotenv import load_dotenv
from typing import List
from datetime import datetime, timedelta
import os
import sys
import pandas as pd
import io
import time 
from concurrent.futures import ThreadPoolExecutor, as_completed
import pendulum
import re
from airflow.operators.bash_operator import BashOperator





sys.path.append('/opt/airflow')

from models.model import Omnibus,Sale,SaleList

load_dotenv()

TASKUSER = os.getenv("AIRFLOW_USER")
SGT = pendulum.timezone("Asia/Singapore")

URL = "https://www.instocktrades.com/newreleases"

logging.basicConfig(level=logging.INFO)

def upload_to_postgresql(chunk: List[any], use: str):
    """Upload a chunk of data to PostgreSQL."""
    headers = {'Content-Type': 'application/json'}
    params = {
        'TaskUser': TASKUSER,  
        'Use': use,
    }
    response = requests.post("http://api-service:8080/api/v1/postgresql/uploadInfo", json=chunk, headers=headers, params=params)

    if response.status_code == 200:
        logging.info(f"Successfully uploaded chunk to PostgreSQL.")
    else:
        logging.error(f"Failed to upload chunk to PostgreSQL: {response.text}")
    if response.status_code != 200:
        raise AirflowFailException(f"Request failed with status code {response.status_code}")


def split_list(data: List[any], chunk_size: int) -> List[List[any]]:
    """Breaks the list into chunks of specified size."""
    return [data[i:i + chunk_size] for i in range(0, len(data), chunk_size)]    

@dag(schedule_interval="0 2 * * 2", start_date=days_ago(1), catchup=True, max_active_runs=1)
def weekly_update_with_new_release():

    vers = 1
    
    @task()
    def fetch_info_data(source) -> List[Omnibus]:
        logging.info("Fetching new release info from the scraper API...")

        params = {
            "TaskUser": TASKUSER,
            "URL": source,
        }

        try:
            response = requests.get("http://api-service:8080/api/v1/scraper/getScrapedInfo", params=params)
            if response.status_code != 200:
                raise AirflowFailException(f"Request failed with status code {response.status_code}")            
            data = response.json()['data']
            logging.info(f"InfoList Length: {len(data)}") 
            return data

        except requests.exceptions.RequestException as e:
            logging.error("Error fetching data: %s", response.json())


    @task()
    def clean_duplicates(infoList: List[Omnibus]) -> List[Omnibus]:
        logging.info("Removing the duplicates")

        # formatted_date = (datetime.now() - timedelta(weeks=1)).strftime("%Y-%m-%d")
        # current_date = datetime.now().strftime("%Y-%m-%d")


        params = {
            "TaskUser": TASKUSER,
            # "Date": formatted_date,
        }

        try:
            response = requests.get("http://api-service:8080/api/v1/postgresql/getInfoByDate", params=params)
            if response.status_code != 200:
                raise AirflowFailException(f"Request failed with status code {response.status_code}")
            duplicatesList = response.json()['values']
            if duplicatesList is not None:
                logging.info(f"duplicatesList Length: {len(duplicatesList)}") 

        except requests.exceptions.RequestException as e:
            logging.error("Error fetching data: %s", e)

        
        try:
            infoList = [o for o in infoList if not any(val in o.values() for val in duplicatesList)]
        except Exception as e:
            logging.info("No duplicates")
        
        logging.info(f"InfoList Length: {len(infoList)}") 

        return infoList

    def split_list(data: List[Omnibus], chunk_size: int) -> List[List[Omnibus]]:
        """Breaks the list into chunks of specified size."""
        return [data[i:i + chunk_size] for i in range(0, len(data), chunk_size)]    
    
    # def split_dataframe(df, chunk_size):
    #     """
    #     Splits a DataFrame into smaller chunks.

    #     :param df: The DataFrame to split.
    #     :param chunk_size: The size of each chunk (in rows).
    #     :return: A list of smaller DataFrames.
    #     """
    #     chunks = []
    #     for start in range(0, len(df), chunk_size):
    #         chunks.append(df.iloc[start:start+chunk_size])
    #     return chunks

    def upload_chunk(chunk_number, parquet_buffer, today_date):
        """Function to upload a chunk to MinIO, retrying until successful."""
        file_name = f'{today_date}_{chunk_number}.parquet'
        files = {
            'file': (file_name, parquet_buffer, 'application/octet-stream'),
        }
        params = {
            'extension' : 'parquet' ,
            'BucketNameKey': 'rawjson',  
            'TaskUser': TASKUSER,
        }

        success = False
        retries = 5  
        attempt = 0

        while not success and attempt < retries:
            try:
                response = requests.post("http://api-service:8080/api/v1/minio/uploadImage", files=files, params=params)

                if response.status_code == 200:
                    logging.info(f"Chunk {chunk_number} uploaded successfully.")
                    success = True  
                else:
                    logging.info(f"Failed to upload chunk {chunk_number}. Response: {response.text}")
                    attempt += 1
                    time.sleep(2)  
            except requests.RequestException as e:
                logging.error(f"Error occurred while uploading chunk {chunk_number}: {e}")
                attempt += 1
                time.sleep(2)  

    @task()
    def upload_minio_postgres(infoList: List[Omnibus]):
        """Convert JSON to Parquet format and upload in chunks asynchronously with retries."""

        logging.info("Running task: json_to_parquet")
  
        chunk_size = 1000000

        today_date = datetime.today().strftime('%Y-%m-%d')  

        chunked_omnibus = split_list(infoList, chunk_size)

        logging.info("Preparing chunks")


        # Use ThreadPoolExecutor to upload chunks asynchronously
        with ThreadPoolExecutor(max_workers=4) as executor:
            futures = []
            for chunk_number, chunk in enumerate(chunked_omnibus, start=vers):
                df = pd.DataFrame(chunk) 
                parquet_buffer = io.BytesIO()
                df.to_parquet(parquet_buffer, engine='pyarrow')
                parquet_buffer.seek(0)

                futures.append(executor.submit(upload_chunk, chunk_number, parquet_buffer, today_date))
                futures.append(executor.submit(upload_to_postgresql, chunk, 'omnibus'))

            # Wait for all futures to complete
            for future in as_completed(futures):
                future.result() 
        


    infoList = fetch_info_data(URL) 
    infoList = clean_duplicates(infoList) 
    upload_minio_postgres(infoList)


# Initialize the DAG
pipeline = weekly_update_with_new_release()
