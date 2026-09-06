\restrict dbmate

-- Dumped from database version 18.4 (Debian 18.4-1.pgdg13+1)
-- Dumped by pg_dump version 18.6

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

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: identities; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.identities (
    id character varying(50) NOT NULL,
    email character varying(255) DEFAULT ''::character varying NOT NULL,
    first_name character varying(255) DEFAULT ''::character varying NOT NULL,
    last_name character varying(255) DEFAULT ''::character varying NOT NULL,
    state character varying(50) DEFAULT 'active'::character varying NOT NULL,
    email_verified_at timestamp with time zone,
    observations text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT check_identity_state CHECK (((state)::text = ANY ((ARRAY['active'::character varying, 'suspended'::character varying, 'banned'::character varying, 'deleted'::character varying])::text[])))
);


--
-- Name: identity_accounts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.identity_accounts (
    id character varying(50) NOT NULL,
    identity_id character varying(50) NOT NULL,
    kind character varying(50) DEFAULT ''::character varying NOT NULL,
    provider_id character varying(255) DEFAULT ''::character varying NOT NULL,
    password_hash character varying(255) DEFAULT ''::character varying NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: identity_sessions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.identity_sessions (
    id character varying(100) NOT NULL,
    identity_id character varying(50) DEFAULT ''::character varying NOT NULL,
    organization_id character varying(50) DEFAULT ''::character varying NOT NULL,
    ip_address character varying(50) DEFAULT ''::character varying NOT NULL,
    user_agent text DEFAULT ''::text NOT NULL,
    expires_at timestamp with time zone DEFAULT (now() + '01:00:00'::interval) NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: invitations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.invitations (
    id character varying(100) NOT NULL,
    kind character varying(50) DEFAULT ''::character varying NOT NULL,
    email character varying(255) DEFAULT ''::character varying NOT NULL,
    first_name character varying(255) DEFAULT ''::character varying NOT NULL,
    last_name character varying(255) DEFAULT ''::character varying NOT NULL,
    target_id character varying(50) NOT NULL,
    inviter_id character varying(50) DEFAULT ''::character varying NOT NULL,
    member_role character varying(50) DEFAULT ''::character varying NOT NULL,
    state character varying(50) DEFAULT 'pending'::character varying NOT NULL,
    expires_at timestamp with time zone DEFAULT (now() + '1 day'::interval) NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT check_invitation_state CHECK (((state)::text = ANY ((ARRAY['pending'::character varying, 'accepted'::character varying, 'declined'::character varying, 'revoked'::character varying])::text[])))
);


--
-- Name: organization_memberships; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.organization_memberships (
    organization_id character varying(50) NOT NULL,
    identity_id character varying(50) NOT NULL,
    member_role character varying(50) DEFAULT ''::character varying NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: organizations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.organizations (
    id character varying(50) NOT NULL,
    name character varying(255) NOT NULL,
    slug character varying(50) DEFAULT ''::character varying NOT NULL,
    kind character varying(50) DEFAULT ''::character varying NOT NULL,
    logo character varying(500) DEFAULT ''::character varying NOT NULL,
    state character varying(50) DEFAULT 'active'::character varying NOT NULL,
    observations text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT check_org_state CHECK (((state)::text = ANY ((ARRAY['active'::character varying, 'suspended'::character varying, 'deleted'::character varying])::text[])))
);


--
-- Name: organization_members_view; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.organization_members_view AS
 SELECT i.id AS identity_id,
    i.email AS identity_email,
    (i.email_verified_at IS NOT NULL) AS identity_email_verified,
    i.first_name AS identity_first_name,
    i.last_name AS identity_last_name,
    i.state AS identity_state,
    i.created_at AS identity_created_at,
    i.updated_at AS identity_updated_at,
    COALESCE(o.id, ''::character varying) AS organization_id,
    COALESCE(o.name, ''::character varying) AS organization_name,
    COALESCE(o.slug, ''::character varying) AS organization_slug,
    COALESCE(o.logo, ''::character varying) AS organization_logo,
    COALESCE(o.kind, ''::character varying) AS organization_kind,
    COALESCE(o.state, ''::character varying) AS organization_state,
    COALESCE(om.member_role, ''::character varying) AS membership_role,
    COALESCE(om.created_at, '0001-01-01 00:00:00+00'::timestamp with time zone) AS membership_created_at,
    COALESCE(om.updated_at, '0001-01-01 00:00:00+00'::timestamp with time zone) AS membership_updated_at
   FROM ((public.identities i
     LEFT JOIN public.organization_memberships om ON (((om.identity_id)::text = (i.id)::text)))
     LEFT JOIN public.organizations o ON (((o.id)::text = (om.organization_id)::text)));


--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.schema_migrations (
    version character varying NOT NULL
);


--
-- Name: session_introspection_view; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.session_introspection_view AS
 SELECT s.id AS session_id,
    s.expires_at AS session_expires_at,
    s.identity_id,
    i.email AS identity_email,
    i.first_name AS identity_first_name,
    i.last_name AS identity_last_name,
    i.state AS identity_state,
    COALESCE(s.organization_id, ''::character varying) AS organization_id,
    COALESCE(o.name, ''::character varying) AS organization_name,
    COALESCE(o.slug, ''::character varying) AS organization_slug,
    COALESCE(o.logo, ''::character varying) AS organization_logo,
    COALESCE(om.member_role, ''::character varying) AS organization_role
   FROM (((public.identity_sessions s
     JOIN public.identities i ON (((i.id)::text = (s.identity_id)::text)))
     LEFT JOIN public.organizations o ON (((o.id)::text = (s.organization_id)::text)))
     LEFT JOIN public.organization_memberships om ON ((((om.organization_id)::text = (s.organization_id)::text) AND ((om.identity_id)::text = (s.identity_id)::text))));


--
-- Name: verifications; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.verifications (
    id character varying(100) NOT NULL,
    identity_id character varying(50) NOT NULL,
    kind character varying(50) DEFAULT ''::character varying NOT NULL,
    hashed_code character varying(255) DEFAULT ''::character varying NOT NULL,
    attempts integer DEFAULT 0 NOT NULL,
    expires_at timestamp with time zone DEFAULT (now() + '00:10:00'::interval) NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: identities identities_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.identities
    ADD CONSTRAINT identities_pkey PRIMARY KEY (id);


--
-- Name: identity_accounts identity_accounts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.identity_accounts
    ADD CONSTRAINT identity_accounts_pkey PRIMARY KEY (id);


--
-- Name: identity_sessions identity_sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.identity_sessions
    ADD CONSTRAINT identity_sessions_pkey PRIMARY KEY (id);


--
-- Name: invitations invitations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invitations
    ADD CONSTRAINT invitations_pkey PRIMARY KEY (id);


--
-- Name: organization_memberships organization_memberships_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.organization_memberships
    ADD CONSTRAINT organization_memberships_pkey PRIMARY KEY (organization_id, identity_id);


--
-- Name: organizations organizations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.organizations
    ADD CONSTRAINT organizations_pkey PRIMARY KEY (id);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: verifications verifications_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.verifications
    ADD CONSTRAINT verifications_pkey PRIMARY KEY (id);


--
-- Name: idx_identity_accounts_provider; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_identity_accounts_provider ON public.identity_accounts USING btree (lower((kind)::text), lower((provider_id)::text));


--
-- Name: idx_identity_email; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_identity_email ON public.identities USING btree (lower((email)::text));


--
-- Name: idx_identity_sessions_expires_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_identity_sessions_expires_at ON public.identity_sessions USING btree (expires_at);


--
-- Name: idx_identity_sessions_identity_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_identity_sessions_identity_id ON public.identity_sessions USING btree (identity_id);


--
-- Name: idx_invitations_email_state; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_invitations_email_state ON public.invitations USING btree (email, state);


--
-- Name: idx_orgs_slug; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_orgs_slug ON public.organizations USING btree (lower((slug)::text));


--
-- Name: identity_accounts identity_accounts_identity_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.identity_accounts
    ADD CONSTRAINT identity_accounts_identity_id_fkey FOREIGN KEY (identity_id) REFERENCES public.identities(id) ON DELETE CASCADE;


--
-- Name: identity_sessions identity_sessions_identity_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.identity_sessions
    ADD CONSTRAINT identity_sessions_identity_id_fkey FOREIGN KEY (identity_id) REFERENCES public.identities(id) ON DELETE CASCADE;


--
-- Name: identity_sessions identity_sessions_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.identity_sessions
    ADD CONSTRAINT identity_sessions_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: invitations invitations_inviter_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invitations
    ADD CONSTRAINT invitations_inviter_id_fkey FOREIGN KEY (inviter_id) REFERENCES public.identities(id) ON DELETE CASCADE;


--
-- Name: organization_memberships organization_memberships_identity_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.organization_memberships
    ADD CONSTRAINT organization_memberships_identity_id_fkey FOREIGN KEY (identity_id) REFERENCES public.identities(id) ON DELETE CASCADE;


--
-- Name: organization_memberships organization_memberships_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.organization_memberships
    ADD CONSTRAINT organization_memberships_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: verifications verifications_identity_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.verifications
    ADD CONSTRAINT verifications_identity_id_fkey FOREIGN KEY (identity_id) REFERENCES public.identities(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

\unrestrict dbmate


--
-- Dbmate schema migrations
--

INSERT INTO public.schema_migrations (version) VALUES
    ('20260902033940'),
    ('20260902034549'),
    ('20260902034841'),
    ('20260902035257'),
    ('20260902035611');
