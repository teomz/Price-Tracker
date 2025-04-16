from airflow.decorators import dag, task
from datetime import datetime, timedelta
from airflow.operators.bash_operator import BashOperator
from airflow.sensors.external_task import ExternalTaskSensor
from airflow.utils.dates import days_ago





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
    timeout=1800,       # give up after 1 hour (optional)
    )

    run_dbt = BashOperator(
        task_id="run_dbt_models",
        bash_command="cd /opt/airflow/dbt/price_tracker && dbt run",
    )

    wait_for_sale_upload  >> run_dbt  # this sets the task to be returned from the DAG function

# Instantiate the DAG
dbt_task_dag = dbt_task()