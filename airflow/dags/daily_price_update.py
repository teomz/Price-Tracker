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



@dag(schedule_interval="18,43 * * * *", start_date=days_ago(1), catchup=False, max_active_runs=1, tags=['price'])
def daily_price_update():

    chunk_size = 1000000 

    @task()
    def get_today_fx_rate() -> str:
        logging.info("Running task: get_today_fx_rate")

        params = {
            'TaskUser': TASKUSER,
        }

        response = requests.get("http://api-service:8080/api/v1/scraper/getCurrency", params=params)
        if response.status_code != 200:
            raise AirflowFailException(f"Request failed with status code {response.status_code}")
        logging.info("get_today_fx_rate: status ",response.raise_for_status())
        fxrate = response.json()['inserted']
        logging.info("Exchange rate received",fxrate)
        return fxrate
    
    @task()
    def getSaleforPlatform(platform: str) -> List[SaleList]:
        logging.info("Running task: getSalefor",platform)

        params = {
            'TaskUser': TASKUSER,
            'Publisher': platform
        }

        response =  requests.get("http://api-service:8080/api/v1/postgresql/getInfoByPublisher", params=params)
        if response.status_code == 200:
            logging.info(f"Successfully get info from PostgreSQL.")
        else:
            logging.error(f"Failed get info from PostgreSQL: {response.json()['error']}")
        if response.status_code != 200:
            raise AirflowFailException(f"Request failed with status code {response.status_code}")
        return response.json()['values']

    @task()
    def cleanUnvalidLink(saleList: List[SaleList]) -> List[SaleList]:
        pattern = "gp/help/customer/display"
        for sale in saleList:
            if re.search(pattern, sale['amazonurl']):
                sale['amazonurl'] = ""  
        return saleList

    @task()
    def getScrapedSale(saleList: List[SaleList]) -> List[Sale]:

        logging.info("Running task: getScrapedSale")

        params = {
            'TaskUser': TASKUSER,
        }

        response =  requests.post("http://api-service:8080/api/v1/scraper/getScrapedSale",json=saleList, params=params)
        if response.status_code == 200:
            logging.info(f"Successfully get info from PostgreSQL.")
        else:
            logging.error(f"Failed get info from PostgreSQL: {response.json()['error']}")
        if response.status_code != 200:
            raise AirflowFailException(f"Request failed with status code {response.status_code}")
        return response.json()['data']
    
    @task()
    def transformSale(saleList: List[Sale], fxrate: str) -> List[Sale]:
        for sale in saleList:
            if sale['platform'] == 'IST':
                if sale['sale'] != -1:
                    sale['sale'] = sale['sale'] * (float(fxrate[0].strip()) + 0.0325)
        return saleList

    
    @task()
    def upload_sale(saleList: List[Sale]):
        logging.info("Running task: upload_sale")

        logging.info("Preparing chunks")
        chunked_sale = split_list(saleList, chunk_size)

        logging.info("Preparing chunks")

        # Use ThreadPoolExecutor to upload chunks asynchronously
        with ThreadPoolExecutor(max_workers=4) as executor:
            futures = []
            for chunk in chunked_sale:
                futures.append(executor.submit(upload_to_postgresql, chunk, 'sale'))
                
            # Wait for all futures to complete
            for future in as_completed(futures):
                future.result() 

    # @task.bash
    # def run_dbt()-> str:
    #     return 'cd /opt/dbt/price_tracker && dbt run'


    for i in ["DC", "Marvel", "Image", "IDW", "Titan", "Boom!", "Dark Horse"]:
        saleList = getSaleforPlatform(i)
        saleList = cleanUnvalidLink(saleList)
        saleList = getScrapedSale(saleList)
        fxrate = get_today_fx_rate()
        saleList = transformSale(saleList, fxrate)
        upload_sale(saleList)

sale_pipeline = daily_price_update()