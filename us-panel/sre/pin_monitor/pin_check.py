"""Draft strict-chain pin monitor, not a panel enforcement configuration."""
import argparse, json
from pathlib import Path
from common import observe, emit, load_inputs

def evaluate(observation, config):
    chain = observation["chain"]; leaf=chain[0]["spki"]; allowed=set(config.get("pins", []))
    unexpected=[c["spki"] for c in chain if c["spki"] not in allowed]
    backup_count=len(set(config.get("backup_pins", [])) & (allowed-{leaf}))
    reasons=[]; status="OK"
    if unexpected or leaf not in allowed:
        status="CRITICAL"; reasons.append("unlisted_chain_spki_or_leaf")
    if backup_count < 2:
        if status == "OK": status="WARN"
        reasons.append("fewer_than_two_distinct_backup_pins")
    rotated=leaf != config.get("previous_leaf", leaf) and leaf in allowed
    if rotated:
        if status == "OK": status="INFO"
        reasons.append("planned_leaf_rotation")
    return {"host":observation["host"], "status":status, "leaf_spki":leaf,
            "unexpected_spki":unexpected, "backup_count":backup_count, "planned_rotation":rotated,
            "reasons":reasons, "configuration_status": config.get("note", "unconfirmed")}
def main():
    a=argparse.ArgumentParser(); a.add_argument("--config",default="pins.json"); a.add_argument("--snapshot")
    a.add_argument("--save-snapshot"); args=a.parse_args(); cfg,snapshots=load_inputs(args); results=[]; captured={}
    for c in cfg["hosts"]:
        try:
            ob=snapshots[c["host"]] if args.snapshot else observe(c["host"])
            captured[c["host"]]=ob; results.append(evaluate(ob,c))
        except Exception as e:
            results.append({"host":c["host"],"status":"CRITICAL","error":str(e)})
    if args.save_snapshot: Path(args.save_snapshot).write_text(json.dumps(captured,indent=2))
    return emit(results)
if __name__=="__main__":
    print("{\"status\":\"UNKNOWN\",\"reason\":\"CANCELLED: pinning will not be used\"}")
    raise SystemExit(3)
