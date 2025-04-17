from airflow.decorators import dag, task
from datetime import datetime, timedelta
from airflow.operators.bash_operator import BashOperator
from airflow.sensors.external_task import ExternalTaskSensor
from airflow.utils.dates import days_ago
from airflow.exceptions import AirflowFailException
import logging
from airflow import DAG
from datetime import datetime
import pandas as pd
from google.oauth2 import service_account
from googleapiclient.discovery import build
from googleapiclient.http import MediaIoBaseUpload
import io
import requests
import os
from dotenv import load_dotenv

load_dotenv()

TASKUSER = os.getenv("AIRFLOW_USER")
FILEID = os.getenv("GOOGLE_FILE")
AMAZONID = os.getenv("GOOGLE_AMAZON")
ISTID = os.getenv("GOOGLE_IST")




@dag(
    dag_id="dbt_task_dag",
    start_date=days_ago(1),
    schedule_interval="25 */4 * * *",  # or "@daily" if you want it scheduled
    catchup=False,
    max_active_runs=1,
    tags=["dbt"]
)
def dbt_task():

    wait_for_sale_upload  = ExternalTaskSensor(
    task_id='wait_for_sale_upload',
    external_dag_id='daily_price_update',
    external_task_id=None,
    allowed_states=['success'],
    execution_delta=timedelta(minutes=25),
    mode='poke',  # or 'reschedule' if you want to free up resources
    poke_interval=60,  # check every 5 minutes
    timeout=3600,       # give up after 1 hour (optional)
    )

    run_dbt = BashOperator(
        task_id="run_dbt_models",
        bash_command="cd /opt/airflow/dbt/price_tracker && dbt run",
    )

    @task
    def query_and_upload(table,fileid):

        logging.info("Running task: getScrapedSale")

        params = {
            'TaskUser': TASKUSER,
            'Table': table,
        }
        
        response =  requests.get("http://api-service:8080/api/v1/postgresql/getAnalytics", params=params)
        if response.status_code == 200:
            logging.info(f"Successfully get info from PostgreSQL.")
        else:
            logging.error(f"Failed get info from PostgreSQL: {response.json()['error']}")
        if response.status_code != 200:
            raise AirflowFailException(f"Request failed with status code {response.status_code}")
        data = response.json()['values']

        df = pd.DataFrame(data)
        logging.info("DataFrame created with shape: %s", df.shape)
        logging.debug("DataFrame content:\n%s", df.head())

        # Convert to CSV in memory
        csv_buffer = io.StringIO()
        df.to_csv(csv_buffer, index=False)
        csv_buffer.seek(0)
        logging.info("CSV buffer prepared.")

        # Service account credentials
        SERVICE_ACCOUNT_FILE = '/opt/airflow/creds/key.json'
        # FOLDER_ID = FILEID  # Replace with actual folder ID if required

        try:
            creds = service_account.Credentials.from_service_account_file(
                SERVICE_ACCOUNT_FILE,
                scopes=['https://www.googleapis.com/auth/drive']
            )
            logging.info("Google service account credentials loaded.")
        except Exception as e:
            logging.error("Failed to load service account credentials: %s", str(e))
            raise

        # Build Google Drive API service
        try:
            service = build('drive', 'v3', credentials=creds)
            logging.info("Google Drive service built successfully.")
        except Exception as e:
            logging.error("Failed to build Google Drive service: %s", str(e))
            raise


        # Upload the file
        media = MediaIoBaseUpload(csv_buffer, mimetype='text/csv')

        try:
            file = service.files().update(
                fileId=fileid,
                media_body=media,
                fields='id'
            ).execute()
            logging.info("File uploaded successfully. File ID: %s", file.get('id'))
        except Exception as e:
            logging.error("Failed to upload file to Google Drive: %s", str(e))
            raise

    wait_for_sale_upload  >> run_dbt >> query_and_upload('analytics_amazon',AMAZONID) >> query_and_upload('analytics_ist',ISTID) 


dbt_task_dag = dbt_task()