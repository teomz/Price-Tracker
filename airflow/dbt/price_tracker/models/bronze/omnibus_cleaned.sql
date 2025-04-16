{{ config(materialized='table') }}
{% set current_date = modules.datetime.date.today().strftime('%Y-%m-%d') %}

SELECT 
  {{ dbt_utils.star(from=ref('omnibus'), except=["status"]) }},
  
  CASE
    WHEN datecreated >= {{ dateadd(datepart="year", interval=-3, from_date_or_timestamp="'" ~ current_date ~ "'::DATE") }}
    THEN 'Hot'
    ELSE 'Cold'
  END AS status

FROM {{ ref('omnibus') }}
