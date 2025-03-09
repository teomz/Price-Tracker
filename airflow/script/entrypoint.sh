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

if [ ! -z "$AIRFLOW__LOGGING__REMOTE_LOGGING" ]; then
    sed -i "s/^remote_logging = .*/remote_logging = $AIRFLOW__LOGGING__REMOTE_LOGGING/" /opt/airflow/airflow.cfg
fi

if [ ! -z "$AIRFLOW__LOGGING__REMOTE_LOG_CONN_ID" ]; then
    sed -i "s/^remote_log_conn_id = .*/remote_log_conn_id = $AIRFLOW__LOGGING__REMOTE_LOG_CONN_ID/" /opt/airflow/airflow.cfg
fi

if [ ! -z "$AIRFLOW__LOGGING__REMOTE_BASE_LOG_FOLDER" ]; then
    sed -i "s|^remote_base_log_folder = .*|remote_base_log_folder = $AIRFLOW__LOGGING__REMOTE_BASE_LOG_FOLDER|" /opt/airflow/airflow.cfg
fi

if [ ! -z "$AIRFLOW__LOGGING__DELETE_LOCAL_LOGS" ]; then
    sed -i "s|^delete_local_logs = .*|delete_local_logs = $AIRFLOW__LOGGING__DELETE_LOCAL_LOGS|" /opt/airflow/airflow.cfg
fi

$(command -v airflow) db migrate

exec airflow webserver