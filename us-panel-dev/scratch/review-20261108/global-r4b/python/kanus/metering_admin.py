"""Inspect interrupted reservations; explicitly refund one after operator review.

No reservation is automatically replayed or refunded on restart. A crashed
worker may have delivered a completed response before it recorded settlement.
Run against the same private local database; this is not a public admin API.
"""
import argparse
import json
from pathlib import Path
from account_store import AccountStore


def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('database',type=Path)
    parser.add_argument('--refund',metavar='RESERVATION_ID')
    args=parser.parse_args()
    if not args.database.is_absolute() or not args.database.is_file() or args.database.is_symlink():
        parser.error('an existing absolute database path is required')
    store=AccountStore(args.database)
    if args.refund:
        print(json.dumps({'refunded':store.finish_reservation(args.refund,False)}))
    else:
        print(json.dumps({'unsettled':store.unsettled_requests()}))


if __name__=='__main__':main()
