{{ config(materialized='table') }}

WITH omnibus_cleaned AS (
    SELECT * FROM {{ ref('omnibus_cleaned') }}
),

sale as (
    select * from {{ source('public', 'sale') }}
)
SELECT DISTINCT
    s.date,
    o.name, 
    last_value(s.sale) OVER (
        PARTITION BY s.date, s.upc, s.platform 
        ORDER BY s.time DESC 
        ROWS BETWEEN UNBOUNDED PRECEDING AND UNBOUNDED FOLLOWING
    ) AS latest_sale, 
    (last_value(s.sale) OVER (
        PARTITION BY s.date, s.upc, s.platform 
        ORDER BY s.time DESC 
        ROWS BETWEEN UNBOUNDED PRECEDING AND UNBOUNDED FOLLOWING
    )/ (o.price * 1.39)) AS percent, 
    CASE 
        WHEN s.platform = 'IST' THEN o.isturl 
        ELSE o.amazonurl 
    END AS url
FROM omnibus_cleaned o
LEFT JOIN sale s ON o.upc = s.upc 
where 1=1
and s.date = current_date and s.sale <> -1 
and upper(s.platform) <> 'AMAZON'
and percent <= 0.75
ORDER BY percent asc