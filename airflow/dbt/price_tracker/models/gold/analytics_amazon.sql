WITH omnibus_cleaned AS (
    SELECT * FROM {{ ref('omnibus_cleaned') }}
),
sale AS (
    SELECT * FROM {{ source('public', 'sale') }}
),
base AS (
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
        ) / (o.price * 1.39)) AS percent,
        s.platform,
        o.isturl,
        o.amazonurl
    FROM omnibus_cleaned o
    LEFT JOIN sale s ON o.upc = s.upc 
    WHERE 1=1
      AND s.date = current_date 
      AND s.sale <> -1 
      AND upper(s.platform) <> 'IST'
)
SELECT
    date,
    name,
    latest_sale,
    percent,
    latest_sale * 0.95 AS "95%_Price",
    percent * 0.95 AS "95%_Percent",
    latest_sale * 0.90 AS "90%_Price",
    percent * 0.90 AS "90%_Percent",
    CASE 
        WHEN platform = 'IST' THEN isturl
        ELSE amazonurl
    END AS url
FROM base
WHERE percent <= 0.75
ORDER BY percent ASC
