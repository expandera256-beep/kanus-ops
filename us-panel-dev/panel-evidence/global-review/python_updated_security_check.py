import sys,socket,smtplib
sys.path.insert(0,'/tmp/global-review.o9517a1v/python_updated/kanus')
import i18n_mail as m
called=[]
def denied(*a,**k):called.append((a,k));raise AssertionError('network/SMTP attempted')
socket.create_connection=denied;smtplib.SMTP=denied;smtplib.SMTP_SSL=denied
for lang in ['../en','/tmp/en','../../tr','en/../../tr','en.json']:
 assert m.render_license_email(lang)[2]=='tr'
payload='{order.__class__.__mro__} {{7*7}} {% include "/etc/passwd" %} <script>alert(1)</script>'
s,b,lang=m.render_license_email('en',order=payload)
assert payload in b; assert not called
assert all(x not in sys.modules for x in ['delivery','mail_worker','jinja2'])
print('PASS: traversal languages fall back to tr; formatter expressions/Jinja syntax remain literal; no network or SMTP attempted; delivery/mail_worker/Jinja not imported. Text renderer only; not an HTML escaper.')
