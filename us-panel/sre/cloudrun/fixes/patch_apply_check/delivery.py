"""Explicit mail worker; the web request never sends mail directly."""
import smtplib
import ssl
import time
import json
import re
import urllib.error
import urllib.request
from email.message import EmailMessage


class DeliveryUncertain(Exception):
    """A remote send may have succeeded; human reconciliation is required."""


def send_hostinger(token, mailbox_id, recipient, tokens, opener=None):
    if not token or not re.fullmatch(r'[A-Za-z0-9_-]{1,100}', mailbox_id):
        raise ValueError('invalid Hostinger configuration')
    body = json.dumps({'to': [recipient], 'subject': 'US lisans anahtarınız',
                       'text': 'US hesabınızı açmak için lisans anahtarınız:\n\n' + '\n'.join(tokens) + '\n'}, ensure_ascii=False).encode('utf-8')
    url = 'https://api.mail.hostinger.com/api/v1/mailboxes/' + mailbox_id + '/send'
    request = urllib.request.Request(url, data=body, method='POST', headers={
        'Authorization': 'Bearer ' + token, 'Content-Type': 'application/json'})
    class NoRedirect(urllib.request.HTTPRedirectHandler):
        def redirect_request(self, req, fp, code, msg, headers, newurl):
            return None
    client = opener or urllib.request.build_opener(urllib.request.ProxyHandler({}), NoRedirect())
    try:
        with client.open(request, timeout=20) as response:
            if response.status != 204:
                raise DeliveryUncertain('unexpected Hostinger response')
    except urllib.error.HTTPError as exc:
        if 400 <= exc.code < 500:
            raise ValueError('Hostinger rejected the request') from exc
        raise DeliveryUncertain('Hostinger delivery status unknown') from exc
    except (OSError, urllib.error.URLError) as exc:
        raise DeliveryUncertain('Hostinger delivery status unknown') from exc


def send_smtp(host, port, username, password, sender, recipient, tokens):
    message=EmailMessage(); message['From']=sender;message['To']=recipient
    message['Subject']='US lisans anahtarınız'
    message.set_content('US hesabınızı açmak için lisans anahtarınız:\n\n'+'\n'.join(tokens)+'\n')
    with smtplib.SMTP_SSL(host,port,context=ssl.create_default_context(),timeout=20) as smtp:
        smtp.login(username,password);smtp.send_message(message)


def deliver_one(store, send, now=None):
    moment=int(time.time() if now is None else now)
    with store._db() as db:
        db.execute('BEGIN IMMEDIATE')
        try:
            row=db.execute("SELECT order_id,email FROM delivery_jobs WHERE state='pending' OR (state='sending' AND lease_until<=?) ORDER BY order_id LIMIT 1",(moment,)).fetchone()
            if not row: db.execute('COMMIT');return False
            tokens=[r[0] for r in db.execute("SELECT token FROM issued_licenses WHERE order_id=? AND status='active' ORDER BY license_hash",(row[0],)).fetchall()]
            if not tokens:
                db.execute("UPDATE delivery_jobs SET state='cancelled' WHERE order_id=?",(row[0],));db.execute('COMMIT');return False
            db.execute("UPDATE delivery_jobs SET state='sending',lease_until=? WHERE order_id=?",(moment+120,row[0]));db.execute('COMMIT')
        except Exception:
            db.execute('ROLLBACK');raise
    try:
        send(row[1],tokens)
    except DeliveryUncertain:
        with store._db() as db: db.execute("UPDATE delivery_jobs SET state='review',lease_until=0 WHERE order_id=? AND state='sending'",(row[0],))
        raise
    except Exception:
        with store._db() as db: db.execute("UPDATE delivery_jobs SET state='pending',lease_until=0 WHERE order_id=? AND state='sending'",(row[0],))
        raise
    with store._db() as db: db.execute("UPDATE delivery_jobs SET state='sent',lease_until=0 WHERE order_id=? AND state='sending'",(row[0],))
    return True
