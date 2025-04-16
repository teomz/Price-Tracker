{% test non_null_unique(model, column_name) %}

SELECT {{ column_name }}
FROM {{ ref(model) }}
WHERE {{ column_name }} IS NOT NULL
GROUP BY {{ column_name }}
HAVING COUNT(*) > 1

{% endtest %}
