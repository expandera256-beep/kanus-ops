import unittest, tempfile, subprocess
from unittest.mock import patch
from pathlib import Path
from datetime import datetime, timezone, timedelta
from cryptography import x509
from cryptography.hazmat.primitives import hashes, serialization
from cryptography.hazmat.primitives.asymmetric import ec
from cryptography.x509.oid import NameOID, ExtendedKeyUsageOID
from common import metadata, observe
from cert_expiry import evaluate
from chain_check import verify_chain

class Tests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.now=datetime.now(timezone.utc)
        # Ephemeral test keys stay in memory: no private-key files or output.
        rk=ec.generate_private_key(ec.SECP256R1()); ik=ec.generate_private_key(ec.SECP256R1())
        lk=ec.generate_private_key(ec.SECP256R1())
        def cert(cn,key,issuer,signer,ca,san=None,days=90):
            name=x509.Name([x509.NameAttribute(NameOID.COMMON_NAME,cn)])
            b=x509.CertificateBuilder().subject_name(name).issuer_name(issuer or name).public_key(key.public_key()).serial_number(x509.random_serial_number()).not_valid_before(cls.now-timedelta(days=1)).not_valid_after(cls.now+timedelta(days=days)).add_extension(x509.BasicConstraints(ca=ca,path_length=None),critical=True)
            if san is not None:b=b.add_extension(x509.SubjectAlternativeName([x509.DNSName(san)]),critical=False)
            if not ca:b=b.add_extension(x509.ExtendedKeyUsage([ExtendedKeyUsageOID.SERVER_AUTH]),critical=False)
            return b.sign(signer,hashes.SHA256())
        cls.root=cert("test root",rk,None,rk,True)
        cls.inter=cert("test intermediate",ik,cls.root.subject,rk,True)
        cls.leaf=cert("service.test",lk,cls.inter.subject,ik,False,"service.test")
        cls.no_san=cert("service.test",lk,cls.inter.subject,ik,False)
        cls.pem=staticmethod(lambda c:c.public_bytes(serialization.Encoding.PEM))
        cls.bundle=cls.pem(cls.leaf)+cls.pem(cls.inter)
        cls.tmp=tempfile.TemporaryDirectory();cls.ca=Path(cls.tmp.name)/"roots.pem";cls.ca.write_bytes(cls.pem(cls.root))
    @classmethod
    def tearDownClass(cls):cls.tmp.cleanup()
    def expiry(self,days,previous=None):
        m=metadata(self.pem(self.leaf));m["notAfter"]=(self.now+timedelta(days=days)).isoformat()
        return evaluate({"host":"service.test","chain":[m]},self.now,previous)
    def test_expiry_26_ok(self):self.assertEqual(self.expiry(26)["status"],"OK")
    def test_expiry_25_ok(self):self.assertEqual(self.expiry(25)["status"],"OK")
    def test_expiry_below25_warn(self):self.assertEqual(self.expiry(24.9)["status"],"WARN")
    def test_expiry_21_warn(self):self.assertEqual(self.expiry(21)["status"],"WARN")
    def test_expiry_7_critical(self):self.assertEqual(self.expiry(7)["status"],"CRITICAL")
    def test_expired_critical(self):self.assertEqual(self.expiry(-1)["status"],"CRITICAL")
    def test_renewal_evidence(self):self.assertEqual(self.expiry(24,(self.now-timedelta(days=90)).isoformat())["status"],"OK")
    def test_no_renewal_warn(self):self.assertEqual(self.expiry(24,metadata(self.pem(self.leaf))["notBefore"])["status"],"WARN")
    def test_chain_san_match(self):self.assertEqual(verify_chain(self.bundle,"service.test",self.ca)["status"],"OK")
    def test_san_mismatch(self):
        with self.assertRaises(Exception):verify_chain(self.bundle,"wrong.test",self.ca)
    def test_san_required(self):
        with self.assertRaises(ValueError):verify_chain(self.pem(self.no_san)+self.pem(self.inter),"service.test",self.ca)
    def test_missing_intermediate(self):
        with self.assertRaisesRegex(ValueError,"issuer certificate"):verify_chain(self.pem(self.leaf),"service.test",self.ca)
    def test_untrusted_root(self):
        with self.assertRaises(ValueError):verify_chain(self.bundle,"service.test")
    def test_timeout_five_seconds(self):
        with patch("common.subprocess.run",side_effect=subprocess.TimeoutExpired("openssl",5)) as run:
            with self.assertRaises(subprocess.TimeoutExpired):observe("service.test")
            self.assertEqual(run.call_count,1);self.assertEqual(run.call_args.kwargs["timeout"],5)
    def test_no_http(self):
        with patch("common.subprocess.run",side_effect=subprocess.TimeoutExpired("openssl",5)) as run:
            with self.assertRaises(subprocess.TimeoutExpired):observe("service.test")
            self.assertEqual(run.call_args.kwargs["input"],b"")
if __name__=="__main__":unittest.main(verbosity=2)
