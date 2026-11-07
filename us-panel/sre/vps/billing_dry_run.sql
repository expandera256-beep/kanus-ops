
SELECT
  DATE(usage_start_time, 'Europe/Istanbul') AS day,
  sku.id AS sku_id,
  sku.description AS sku_description,
  SUM(cost) AS cost,
  SUM(IFNULL((SELECT SUM(c.amount) FROM UNNEST(credits) c), 0)) AS credits,
  SUM(usage.amount_in_pricing_units) AS usage_pricing_units,
  ANY_VALUE(usage.pricing_unit) AS pricing_unit,
  ANY_VALUE(currency) AS currency
FROM `project-27082a40-6083-4952-9db.billing_export.gcp_billing_export_resource_v1_017786_B62E9C_E21182`
WHERE _PARTITIONTIME >= TIMESTAMP_SUB(TIMESTAMP(@start_day), INTERVAL 1 DAY)
  AND usage_start_time >= TIMESTAMP(@start_day, 'Europe/Istanbul')
  AND usage_start_time <  TIMESTAMP(@end_day, 'Europe/Istanbul')
  AND service.description = 'Compute Engine'
  AND (resource.name IN UNNEST(@names)
       OR EXISTS(SELECT 1 FROM UNNEST(@names) n WHERE ENDS_WITH(resource.global_name, CONCAT('/', n))))
GROUP BY day, sku_id, sku_description
ORDER BY day, sku_id
