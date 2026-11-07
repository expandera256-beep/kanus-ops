--
-- PostgreSQL database dump
--

\restrict ytZsgEDn44t2BL9U3vDOxZeAPvyp9HfAnb6FXUxe2B6FCfAduY7GawZOSVKGAx8

-- Dumped from database version 17.11
-- Dumped by pg_dump version 17.11 (Debian 17.11-0+deb13u1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: us_app; Type: SCHEMA; Schema: -; Owner: us_runtime
--

CREATE SCHEMA us_app;


ALTER SCHEMA us_app OWNER TO us_runtime;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: accounts; Type: TABLE; Schema: us_app; Owner: us_runtime
--

CREATE TABLE us_app.accounts (
    id bigint NOT NULL,
    email text NOT NULL,
    salt bytea NOT NULL,
    password bytea NOT NULL,
    license_hash text NOT NULL,
    expires bigint NOT NULL,
    quota bigint NOT NULL,
    used bigint DEFAULT 0 NOT NULL
);


ALTER TABLE us_app.accounts OWNER TO us_runtime;

--
-- Name: accounts_id_seq; Type: SEQUENCE; Schema: us_app; Owner: us_runtime
--

CREATE SEQUENCE us_app.accounts_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE us_app.accounts_id_seq OWNER TO us_runtime;

--
-- Name: accounts_id_seq; Type: SEQUENCE OWNED BY; Schema: us_app; Owner: us_runtime
--

ALTER SEQUENCE us_app.accounts_id_seq OWNED BY us_app.accounts.id;


--
-- Name: auth_attempts; Type: TABLE; Schema: us_app; Owner: us_runtime
--

CREATE TABLE us_app.auth_attempts (
    scope text NOT NULL,
    window_start bigint NOT NULL,
    attempts bigint NOT NULL
);


ALTER TABLE us_app.auth_attempts OWNER TO us_runtime;

--
-- Name: credit_packs; Type: TABLE; Schema: us_app; Owner: us_runtime
--

CREATE TABLE us_app.credit_packs (
    license_hash text NOT NULL,
    account_id bigint NOT NULL,
    expires bigint NOT NULL,
    quota bigint NOT NULL,
    used bigint DEFAULT 0 NOT NULL
);


ALTER TABLE us_app.credit_packs OWNER TO us_runtime;

--
-- Name: delivery_jobs; Type: TABLE; Schema: us_app; Owner: us_runtime
--

CREATE TABLE us_app.delivery_jobs (
    order_id text NOT NULL,
    email text NOT NULL,
    state text DEFAULT 'pending'::text NOT NULL,
    lease_until bigint DEFAULT 0 NOT NULL
);


ALTER TABLE us_app.delivery_jobs OWNER TO us_runtime;

--
-- Name: inference_events; Type: TABLE; Schema: us_app; Owner: us_runtime
--

CREATE TABLE us_app.inference_events (
    event_id text NOT NULL,
    worker text NOT NULL,
    created bigint NOT NULL,
    kind text NOT NULL,
    detail text NOT NULL
);


ALTER TABLE us_app.inference_events OWNER TO us_runtime;

--
-- Name: inference_watchdog; Type: TABLE; Schema: us_app; Owner: us_runtime
--

CREATE TABLE us_app.inference_watchdog (
    name text NOT NULL,
    heartbeat bigint NOT NULL
);


ALTER TABLE us_app.inference_watchdog OWNER TO us_runtime;

--
-- Name: inference_workers; Type: TABLE; Schema: us_app; Owner: us_runtime
--

CREATE TABLE us_app.inference_workers (
    worker text NOT NULL,
    state text NOT NULL
);


ALTER TABLE us_app.inference_workers OWNER TO us_runtime;

--
-- Name: issued_licenses; Type: TABLE; Schema: us_app; Owner: us_runtime
--

CREATE TABLE us_app.issued_licenses (
    license_hash text NOT NULL,
    order_id text NOT NULL,
    token text NOT NULL,
    email text NOT NULL,
    status text NOT NULL
);


ALTER TABLE us_app.issued_licenses OWNER TO us_runtime;

--
-- Name: metered_requests; Type: TABLE; Schema: us_app; Owner: us_runtime
--

CREATE TABLE us_app.metered_requests (
    reservation text NOT NULL,
    account_id bigint NOT NULL,
    request_id text NOT NULL,
    request_hash text NOT NULL,
    source text NOT NULL,
    source_id text NOT NULL,
    state text NOT NULL,
    created bigint NOT NULL,
    updated bigint NOT NULL
);


ALTER TABLE us_app.metered_requests OWNER TO us_runtime;

--
-- Name: sessions; Type: TABLE; Schema: us_app; Owner: us_runtime
--

CREATE TABLE us_app.sessions (
    token_hash text NOT NULL,
    account_id bigint NOT NULL,
    expires bigint NOT NULL
);


ALTER TABLE us_app.sessions OWNER TO us_runtime;

--
-- Name: usage; Type: TABLE; Schema: us_app; Owner: us_runtime
--

CREATE TABLE us_app.usage (
    account_id bigint NOT NULL,
    request_id text NOT NULL
);


ALTER TABLE us_app.usage OWNER TO us_runtime;

--
-- Name: webhook_orders; Type: TABLE; Schema: us_app; Owner: us_runtime
--

CREATE TABLE us_app.webhook_orders (
    order_id text NOT NULL,
    body_hash text NOT NULL
);


ALTER TABLE us_app.webhook_orders OWNER TO us_runtime;

--
-- Name: accounts id; Type: DEFAULT; Schema: us_app; Owner: us_runtime
--

ALTER TABLE ONLY us_app.accounts ALTER COLUMN id SET DEFAULT nextval('us_app.accounts_id_seq'::regclass);


--
-- Data for Name: accounts; Type: TABLE DATA; Schema: us_app; Owner: us_runtime
--

COPY us_app.accounts (id, email, salt, password, license_hash, expires, quota, used) FROM stdin;
\.


--
-- Data for Name: auth_attempts; Type: TABLE DATA; Schema: us_app; Owner: us_runtime
--

COPY us_app.auth_attempts (scope, window_start, attempts) FROM stdin;
\.


--
-- Data for Name: credit_packs; Type: TABLE DATA; Schema: us_app; Owner: us_runtime
--

COPY us_app.credit_packs (license_hash, account_id, expires, quota, used) FROM stdin;
\.


--
-- Data for Name: delivery_jobs; Type: TABLE DATA; Schema: us_app; Owner: us_runtime
--

COPY us_app.delivery_jobs (order_id, email, state, lease_until) FROM stdin;
\.


--
-- Data for Name: inference_events; Type: TABLE DATA; Schema: us_app; Owner: us_runtime
--

COPY us_app.inference_events (event_id, worker, created, kind, detail) FROM stdin;
\.


--
-- Data for Name: inference_watchdog; Type: TABLE DATA; Schema: us_app; Owner: us_runtime
--

COPY us_app.inference_watchdog (name, heartbeat) FROM stdin;
external	1793938862
\.


--
-- Data for Name: inference_workers; Type: TABLE DATA; Schema: us_app; Owner: us_runtime
--

COPY us_app.inference_workers (worker, state) FROM stdin;
ds-a100	{"billed_at":1793938862.9293423,"day":20763,"day_seconds":8461.137307882309,"error":"","last_activity":0,"leases":{},"op":null,"phase":"running","provider_status":"RUNNING","run_since":1793936393.903,"starts":0}
\.


--
-- Data for Name: issued_licenses; Type: TABLE DATA; Schema: us_app; Owner: us_runtime
--

COPY us_app.issued_licenses (license_hash, order_id, token, email, status) FROM stdin;
\.


--
-- Data for Name: metered_requests; Type: TABLE DATA; Schema: us_app; Owner: us_runtime
--

COPY us_app.metered_requests (reservation, account_id, request_id, request_hash, source, source_id, state, created, updated) FROM stdin;
\.


--
-- Data for Name: sessions; Type: TABLE DATA; Schema: us_app; Owner: us_runtime
--

COPY us_app.sessions (token_hash, account_id, expires) FROM stdin;
\.


--
-- Data for Name: usage; Type: TABLE DATA; Schema: us_app; Owner: us_runtime
--

COPY us_app.usage (account_id, request_id) FROM stdin;
\.


--
-- Data for Name: webhook_orders; Type: TABLE DATA; Schema: us_app; Owner: us_runtime
--

COPY us_app.webhook_orders (order_id, body_hash) FROM stdin;
\.


--
-- Name: accounts_id_seq; Type: SEQUENCE SET; Schema: us_app; Owner: us_runtime
--

SELECT pg_catalog.setval('us_app.accounts_id_seq', 1, false);


--
-- Name: accounts accounts_email_key; Type: CONSTRAINT; Schema: us_app; Owner: us_runtime
--

ALTER TABLE ONLY us_app.accounts
    ADD CONSTRAINT accounts_email_key UNIQUE (email);


--
-- Name: accounts accounts_license_hash_key; Type: CONSTRAINT; Schema: us_app; Owner: us_runtime
--

ALTER TABLE ONLY us_app.accounts
    ADD CONSTRAINT accounts_license_hash_key UNIQUE (license_hash);


--
-- Name: accounts accounts_pkey; Type: CONSTRAINT; Schema: us_app; Owner: us_runtime
--

ALTER TABLE ONLY us_app.accounts
    ADD CONSTRAINT accounts_pkey PRIMARY KEY (id);


--
-- Name: auth_attempts auth_attempts_pkey; Type: CONSTRAINT; Schema: us_app; Owner: us_runtime
--

ALTER TABLE ONLY us_app.auth_attempts
    ADD CONSTRAINT auth_attempts_pkey PRIMARY KEY (scope);


--
-- Name: credit_packs credit_packs_pkey; Type: CONSTRAINT; Schema: us_app; Owner: us_runtime
--

ALTER TABLE ONLY us_app.credit_packs
    ADD CONSTRAINT credit_packs_pkey PRIMARY KEY (license_hash);


--
-- Name: delivery_jobs delivery_jobs_pkey; Type: CONSTRAINT; Schema: us_app; Owner: us_runtime
--

ALTER TABLE ONLY us_app.delivery_jobs
    ADD CONSTRAINT delivery_jobs_pkey PRIMARY KEY (order_id);


--
-- Name: inference_events inference_events_pkey; Type: CONSTRAINT; Schema: us_app; Owner: us_runtime
--

ALTER TABLE ONLY us_app.inference_events
    ADD CONSTRAINT inference_events_pkey PRIMARY KEY (event_id);


--
-- Name: inference_watchdog inference_watchdog_pkey; Type: CONSTRAINT; Schema: us_app; Owner: us_runtime
--

ALTER TABLE ONLY us_app.inference_watchdog
    ADD CONSTRAINT inference_watchdog_pkey PRIMARY KEY (name);


--
-- Name: inference_workers inference_workers_pkey; Type: CONSTRAINT; Schema: us_app; Owner: us_runtime
--

ALTER TABLE ONLY us_app.inference_workers
    ADD CONSTRAINT inference_workers_pkey PRIMARY KEY (worker);


--
-- Name: issued_licenses issued_licenses_order_id_license_hash_key; Type: CONSTRAINT; Schema: us_app; Owner: us_runtime
--

ALTER TABLE ONLY us_app.issued_licenses
    ADD CONSTRAINT issued_licenses_order_id_license_hash_key UNIQUE (order_id, license_hash);


--
-- Name: issued_licenses issued_licenses_pkey; Type: CONSTRAINT; Schema: us_app; Owner: us_runtime
--

ALTER TABLE ONLY us_app.issued_licenses
    ADD CONSTRAINT issued_licenses_pkey PRIMARY KEY (license_hash);


--
-- Name: metered_requests metered_requests_account_id_request_id_key; Type: CONSTRAINT; Schema: us_app; Owner: us_runtime
--

ALTER TABLE ONLY us_app.metered_requests
    ADD CONSTRAINT metered_requests_account_id_request_id_key UNIQUE (account_id, request_id);


--
-- Name: metered_requests metered_requests_pkey; Type: CONSTRAINT; Schema: us_app; Owner: us_runtime
--

ALTER TABLE ONLY us_app.metered_requests
    ADD CONSTRAINT metered_requests_pkey PRIMARY KEY (reservation);


--
-- Name: sessions sessions_pkey; Type: CONSTRAINT; Schema: us_app; Owner: us_runtime
--

ALTER TABLE ONLY us_app.sessions
    ADD CONSTRAINT sessions_pkey PRIMARY KEY (token_hash);


--
-- Name: usage usage_pkey; Type: CONSTRAINT; Schema: us_app; Owner: us_runtime
--

ALTER TABLE ONLY us_app.usage
    ADD CONSTRAINT usage_pkey PRIMARY KEY (account_id, request_id);


--
-- Name: webhook_orders webhook_orders_pkey; Type: CONSTRAINT; Schema: us_app; Owner: us_runtime
--

ALTER TABLE ONLY us_app.webhook_orders
    ADD CONSTRAINT webhook_orders_pkey PRIMARY KEY (order_id);


--
-- Name: credit_packs credit_packs_account_id_fkey; Type: FK CONSTRAINT; Schema: us_app; Owner: us_runtime
--

ALTER TABLE ONLY us_app.credit_packs
    ADD CONSTRAINT credit_packs_account_id_fkey FOREIGN KEY (account_id) REFERENCES us_app.accounts(id);


--
-- Name: credit_packs credit_packs_license_hash_fkey; Type: FK CONSTRAINT; Schema: us_app; Owner: us_runtime
--

ALTER TABLE ONLY us_app.credit_packs
    ADD CONSTRAINT credit_packs_license_hash_fkey FOREIGN KEY (license_hash) REFERENCES us_app.issued_licenses(license_hash);


--
-- Name: metered_requests metered_requests_account_id_fkey; Type: FK CONSTRAINT; Schema: us_app; Owner: us_runtime
--

ALTER TABLE ONLY us_app.metered_requests
    ADD CONSTRAINT metered_requests_account_id_fkey FOREIGN KEY (account_id) REFERENCES us_app.accounts(id);


--
-- Name: sessions sessions_account_id_fkey; Type: FK CONSTRAINT; Schema: us_app; Owner: us_runtime
--

ALTER TABLE ONLY us_app.sessions
    ADD CONSTRAINT sessions_account_id_fkey FOREIGN KEY (account_id) REFERENCES us_app.accounts(id);


--
-- Name: usage usage_account_id_fkey; Type: FK CONSTRAINT; Schema: us_app; Owner: us_runtime
--

ALTER TABLE ONLY us_app.usage
    ADD CONSTRAINT usage_account_id_fkey FOREIGN KEY (account_id) REFERENCES us_app.accounts(id);


--
-- Name: SCHEMA public; Type: ACL; Schema: -; Owner: pg_database_owner
--

GRANT ALL ON SCHEMA public TO cloudsqlsuperuser;


--
-- PostgreSQL database dump complete
--

\unrestrict ytZsgEDn44t2BL9U3vDOxZeAPvyp9HfAnb6FXUxe2B6FCfAduY7GawZOSVKGAx8

