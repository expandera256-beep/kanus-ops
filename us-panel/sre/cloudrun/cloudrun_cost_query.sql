SELECT resource.name AS service, sku.description AS sku, SUM(cost) AS cost,
  SUM(IFNULL((SELECT SUM(c.amount) FROM UNNEST(credits) c),0)) AS credits
FROM `project-27082a40-6083-4952-9db.billing_export.gcp_billing_export_resource_v1_017786_B62E9C_E21182`
WHERE _PARTITIONTIME >= TIMESTAMP('2026-10-31') AND usage_start_time >= TIMESTAMP('2026-11-01')
  AND service.description IN ('Cloud Run','Cloud Scheduler')
GROUP BY service, sku
