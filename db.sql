--
-- PostgreSQL database dump
--

-- Dumped from database version 13.2 (Ubuntu 13.2-1.pgdg20.04+1)
-- Dumped by pg_dump version 13.2 (Ubuntu 13.2-1.pgdg20.04+1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;



--
-- Name: auth; Type: TABLE; Schema: public; Owner: xanadu
--

CREATE TABLE public.auth (
    username character varying(20) PRIMARY KEY,
    hash character varying(200)
);


ALTER TABLE public.auth OWNER TO xanadu;

--
-- Name: domains; Type: TABLE; Schema: public; Owner: xanadu
--

CREATE TABLE public.domains (
    origin character varying(50) NOT NULL,
    url character varying(200),
);
ALTER TABLE public.domains OWNER TO xanadu;

--
-- Name: history;Type:Table;Schema : public;Owner: xanadu
--

CREATE TABLE history (
    username character varying(20),
    time timestamp  with time,
    title character varying(100),
    url character varying(200),
)
ALTER TABLE public.history OWNER TO xanadu;

--
-- Name: sessions; Type: TABLE; Schema: public; Owner: xanadu
--

CREATE TABLE public.sessions (
    username character varying(20),
    sessionid character varying(100)
);


ALTER TABLE public.sessions OWNER TO xanadu;

--
-- Name: domains domains_pkey; Type: CONSTRAINT; Schema: public; Owner: xanadu
--

ALTER TABLE ONLY public.domains
    ADD CONSTRAINT domains_pkey PRIMARY KEY (origin);


--
-- PostgreSQL database dump complete
--

