#!/usr/bin/env bash
# HAZIRLIK — CALISTIRILMADI. Demir onayi olmadan calistirmayin.
# Karar: aylik 200 USD (Demir karari 06.11.2026 09:36 TSI, KAN.US araciligiyla).
# Kapsam: yalniz project-27082a40-6083-4952-9db; esikler 50/80/100% (current spend).
# Bildirim: varsayilan IAM e-postasi (billing admin/user rolleri). Pub/Sub YOK, otomatik kapatma YOK.
# On kosul (06.11 09:4x TSI tespiti): billingbudgets.googleapis.com projede KAPALI.
#   Etkinlestirme ayri bir yazma islemidir ve ayrica onay gerektirir:
#   gcloud services enable billingbudgets.googleapis.com --project=project-27082a40-6083-4952-9db
set -euo pipefail
BA=017786-B62E9C-E21182
PROJECT=project-27082a40-6083-4952-9db
NAME="kanus-${PROJECT}-monthly-200usd"
GCLOUD=${GCLOUD:-gcloud}

if [[ "${1:-}" == "--rollback" ]]; then
  # GERI ALMA: yalniz bu isimdeki butceyi siler
  for B in $($GCLOUD billing budgets list --billing-account="$BA" --billing-project="$PROJECT" --filter="displayName=$NAME" --format="value(name)"); do
    $GCLOUD billing budgets delete "$B" --billing-account="$BA" --billing-project="$PROJECT" --quiet
  done
  exit 0
fi
if [[ "${1:-}" != "--i-have-approval" ]]; then
  echo "Onay bayragi yok; hicbir sey yapilmadi. Kullanim: $0 --i-have-approval | --rollback" >&2; exit 2
fi

$GCLOUD billing budgets create \
  --billing-account="$BA" \
  --display-name="$NAME" \
  --budget-amount=200USD \
  --calendar-period=month \
  --filter-projects="projects/247516932580" \
  --threshold-rule=percent=0.5,basis=current-spend \
  --threshold-rule=percent=0.8,basis=current-spend \
  --threshold-rule=percent=1.0,basis=current-spend \
  --billing-project="$PROJECT"
# not: --notifications-rule-pubsub-topic ve --disable-default-iam-recipients bilincli olarak VERILMEDI.

$GCLOUD billing budgets list --billing-account="$BA" --billing-project="$PROJECT" \
  --filter="displayName=$NAME" --format="value(name)"
