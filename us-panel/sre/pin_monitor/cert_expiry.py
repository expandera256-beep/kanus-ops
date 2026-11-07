"""Certificate expiration monitor. Exit 0 OK, 1 WARN, 2 CRITICAL, 3 UNKNOWN."""
import argparse
from datetime import datetime, timezone
from common import observe, emit, load_inputs

def evaluate(observation, now=None, previous_not_before=None):
    now = now or datetime.now(timezone.utc)
    leaf = observation["chain"][0]
    days = (datetime.fromisoformat(leaf["notAfter"]) - now).total_seconds()/86400
    reasons=[]
    status="OK"
    if days <= 7:
        status="CRITICAL"; reasons.append("expiry_at_or_below_7_days")
    elif days <= 21:
        status="WARN"; reasons.append("expiry_at_or_below_21_days")
    if days < 25:
        # No renewal can be proved without previous observation: never claim one.
        if not previous_not_before or datetime.fromisoformat(leaf["notBefore"]) <= datetime.fromisoformat(previous_not_before):
            if status == "OK": status="WARN"
            reasons.append("renewal_window_below_25_days_no_renewal_evidence")
    return {"host": observation["host"], "status": status, "days_remaining": round(days, 5),
            "notAfter": leaf["notAfter"], "reasons": reasons,
            "renewal_history": "available" if previous_not_before else "not_available"}
def main():
    a=argparse.ArgumentParser(); a.add_argument("--config", default="cert_hosts.json"); a.add_argument("--snapshot")
    args=a.parse_args(); cfg, snapshots=load_inputs(args); results=[]
    for c in cfg["hosts"]:
        try:
            ob=snapshots[c["host"]] if args.snapshot else observe(c["host"])
            results.append(evaluate(ob, previous_not_before=c.get("previous_not_before")))
        except Exception as e:
            results.append({"host":c["host"], "status":"UNKNOWN", "error":str(e)})
    return emit(results)
if __name__ == "__main__": raise SystemExit(main())
