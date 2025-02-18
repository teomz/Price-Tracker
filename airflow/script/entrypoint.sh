#!/bin/bash

set -e

if [ -e "/opt/airflow/requirements.txt" ]; then
  $(command python) pip install --upgrade pip
  $(command -v pip) install --user -r requirements.txt
fi

if [ ! -f "/opt/airflow/airflow.db" ]; then
  airflow db init && \
  airflow users create \
    --username ${AIRFLOW_WWW_USER} \
    --firstname admin \
    --lastname admin \
    --role Admin \
    --email ${AIRFLOW_WWW_EMAIL} \
    --password ${AIRFLOW_WWW_PASSWORD}
fi

$(command -v airflow) db migrate

exec airflow webserver