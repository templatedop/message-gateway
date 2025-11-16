-- DROP SCHEMA msggateway;

CREATE SCHEMA msggateway AUTHORIZATION msggateway_admin;

-- DROP SEQUENCE msggateway.msg_application_application_id_seq;

CREATE SEQUENCE msggateway.msg_application_application_id_seq
	INCREMENT BY 1
	MINVALUE 1
	MAXVALUE 2147483647
	START 1
	CACHE 1
	NO CYCLE;

-- Permissions

ALTER SEQUENCE msggateway.msg_application_application_id_seq OWNER TO msggateway_admin;
GRANT ALL ON SEQUENCE msggateway.msg_application_application_id_seq TO msggateway_admin;
GRANT SELECT ON SEQUENCE msggateway.msg_application_application_id_seq TO msggateway_ro;
GRANT ALL ON SEQUENCE msggateway.msg_application_application_id_seq TO msggateway_rw;

-- DROP SEQUENCE msggateway.msg_bulk_files_file_id_seq;

CREATE SEQUENCE msggateway.msg_bulk_files_file_id_seq
	INCREMENT BY 1
	MINVALUE 1
	MAXVALUE 2147483647
	START 1
	CACHE 1
	NO CYCLE;

-- Permissions

ALTER SEQUENCE msggateway.msg_bulk_files_file_id_seq OWNER TO msggateway_admin;
GRANT ALL ON SEQUENCE msggateway.msg_bulk_files_file_id_seq TO msggateway_admin;
GRANT SELECT ON SEQUENCE msggateway.msg_bulk_files_file_id_seq TO msggateway_ro;
GRANT ALL ON SEQUENCE msggateway.msg_bulk_files_file_id_seq TO msggateway_rw;

-- DROP SEQUENCE msggateway.msg_counter_counterid_seq;

CREATE SEQUENCE msggateway.msg_counter_counterid_seq
	INCREMENT BY 1
	MINVALUE 1
	MAXVALUE 2147483647
	START 1
	CACHE 1
	NO CYCLE;

-- Permissions

ALTER SEQUENCE msggateway.msg_counter_counterid_seq OWNER TO msggateway_admin;
GRANT ALL ON SEQUENCE msggateway.msg_counter_counterid_seq TO msggateway_admin;
GRANT SELECT ON SEQUENCE msggateway.msg_counter_counterid_seq TO msggateway_ro;
GRANT ALL ON SEQUENCE msggateway.msg_counter_counterid_seq TO msggateway_rw;

-- DROP SEQUENCE msggateway.msg_provider_provider_id_seq;

CREATE SEQUENCE msggateway.msg_provider_provider_id_seq
	INCREMENT BY 1
	MINVALUE 1
	MAXVALUE 2147483647
	START 1
	CACHE 1
	NO CYCLE;

-- Permissions

ALTER SEQUENCE msggateway.msg_provider_provider_id_seq OWNER TO msggateway_admin;
GRANT ALL ON SEQUENCE msggateway.msg_provider_provider_id_seq TO msggateway_admin;
GRANT SELECT ON SEQUENCE msggateway.msg_provider_provider_id_seq TO msggateway_ro;
GRANT ALL ON SEQUENCE msggateway.msg_provider_provider_id_seq TO msggateway_rw;

-- DROP SEQUENCE msggateway.msg_request_req_id_seq;

CREATE SEQUENCE msggateway.msg_request_req_id_seq
	INCREMENT BY 1
	MINVALUE 1
	MAXVALUE 2147483647
	START 1
	CACHE 1
	NO CYCLE;

-- Permissions

ALTER SEQUENCE msggateway.msg_request_req_id_seq OWNER TO msggateway_admin;
GRANT ALL ON SEQUENCE msggateway.msg_request_req_id_seq TO msggateway_admin;
GRANT SELECT ON SEQUENCE msggateway.msg_request_req_id_seq TO msggateway_ro;
GRANT ALL ON SEQUENCE msggateway.msg_request_req_id_seq TO msggateway_rw;

-- DROP SEQUENCE msggateway.msg_template_template_local_id_seq;

CREATE SEQUENCE msggateway.msg_template_template_local_id_seq
	INCREMENT BY 1
	MINVALUE 1
	MAXVALUE 2147483647
	START 1
	CACHE 1
	NO CYCLE;

-- Permissions

ALTER SEQUENCE msggateway.msg_template_template_local_id_seq OWNER TO msggateway_admin;
GRANT ALL ON SEQUENCE msggateway.msg_template_template_local_id_seq TO msggateway_admin;
GRANT SELECT ON SEQUENCE msggateway.msg_template_template_local_id_seq TO msggateway_ro;
GRANT ALL ON SEQUENCE msggateway.msg_template_template_local_id_seq TO msggateway_rw;

-- DROP SEQUENCE msggateway.pg_log_log_id_seq;

CREATE SEQUENCE msggateway.pg_log_log_id_seq
	INCREMENT BY 1
	MINVALUE 1
	MAXVALUE 2147483647
	START 1
	CACHE 1
	NO CYCLE;

-- Permissions

ALTER SEQUENCE msggateway.pg_log_log_id_seq OWNER TO msggateway_admin;
GRANT ALL ON SEQUENCE msggateway.pg_log_log_id_seq TO msggateway_admin;
GRANT SELECT ON SEQUENCE msggateway.pg_log_log_id_seq TO msggateway_ro;
GRANT ALL ON SEQUENCE msggateway.pg_log_log_id_seq TO msggateway_rw;
-- msggateway.msg_application definition

-- Drop table

-- DROP TABLE msggateway.msg_application;

CREATE TABLE msggateway.msg_application (
	application_id serial4 NOT NULL,
	application_name varchar NULL,
	request_type varchar NULL,
	secret_key varchar NULL,
	created_date timestamp DEFAULT CURRENT_TIMESTAMP NULL,
	updated_date timestamp DEFAULT CURRENT_TIMESTAMP NULL,
	status_cd int4 NULL,
	CONSTRAINT pg_applications_pkey_new PRIMARY KEY (application_id)
);
CREATE UNIQUE INDEX idx_msg_application_application_id ON msggateway.msg_application USING btree (application_id);
CREATE INDEX idx_msg_application_application_name ON msggateway.msg_application USING btree (application_name);
CREATE UNIQUE INDEX idx_msg_application_application_name1 ON msggateway.msg_application USING btree (application_name);

-- Permissions

ALTER TABLE msggateway.msg_application OWNER TO msggateway_admin;
GRANT ALL ON TABLE msggateway.msg_application TO msggateway_admin;
GRANT SELECT ON TABLE msggateway.msg_application TO msggateway_ro;
GRANT INSERT, UPDATE, DELETE, SELECT ON TABLE msggateway.msg_application TO msggateway_rw;


-- msggateway.msg_bulk_file definition

-- Drop table

-- DROP TABLE msggateway.msg_bulk_file;

CREATE TABLE msggateway.msg_bulk_file (
	file_id int4 DEFAULT nextval('msggateway.msg_bulk_files_file_id_seq'::regclass) NOT NULL,
	reference_id bpchar(32) DEFAULT msggateway.generate_random_string(20) NULL,
	application_id varchar NULL,
	template_name varchar NULL,
	template_id varchar NULL,
	entity_id varchar NULL,
	sender_id varchar NULL,
	test_msg varchar NULL,
	is_verified bool DEFAULT false NULL,
	file_name varchar NULL,
	is_valid_file bool DEFAULT false NULL,
	no_of_sms_uploaded int8 NULL,
	no_of_sms_sent int8 NULL,
	no_of_sms_failed int8 NULL,
	no_of_sms_pending int8 NULL,
	scheduled varchar NULL,
	status_cd int4 DEFAULT 0 NULL,
	uploaded_time timestamp DEFAULT CURRENT_TIMESTAMP NULL,
	updated_time timestamp NULL,
	mobile_number _int8 NULL,
	message_type varchar(2) NULL,
	CONSTRAINT msg_bulk_files_file_id PRIMARY KEY (file_id)
);
CREATE INDEX idx_msg_bulk_file_reference_id ON msggateway.msg_bulk_file USING btree (reference_id);
CREATE INDEX idx_msg_bulk_files_file_id ON msggateway.msg_bulk_file USING btree (file_id);

-- Permissions

ALTER TABLE msggateway.msg_bulk_file OWNER TO msggateway_admin;
GRANT ALL ON TABLE msggateway.msg_bulk_file TO msggateway_admin;
GRANT SELECT ON TABLE msggateway.msg_bulk_file TO msggateway_ro;
GRANT INSERT, UPDATE, DELETE, SELECT ON TABLE msggateway.msg_bulk_file TO msggateway_rw;


-- msggateway.msg_counter definition

-- Drop table

-- DROP TABLE msggateway.msg_counter;

CREATE TABLE msggateway.msg_counter (
	counter_id int4 DEFAULT nextval('msggateway.msg_counter_counterid_seq'::regclass) NOT NULL,
	request_date timestamp NULL,
	gateway varchar NULL,
	count int4 NULL,
	CONSTRAINT msg_counter_key PRIMARY KEY (counter_id)
);

-- Permissions

ALTER TABLE msggateway.msg_counter OWNER TO msggateway_admin;
GRANT ALL ON TABLE msggateway.msg_counter TO msggateway_admin;
GRANT SELECT ON TABLE msggateway.msg_counter TO msggateway_ro;
GRANT INSERT, UPDATE, DELETE, SELECT ON TABLE msggateway.msg_counter TO msggateway_rw;


-- msggateway.msg_log definition

-- Drop table

-- DROP TABLE msggateway.msg_log;

CREATE TABLE msggateway.msg_log (
	log_id int4 DEFAULT nextval('msggateway.pg_log_log_id_seq'::regclass) NOT NULL,
	payload varchar NULL,
	remarks varchar NULL,
	created_date timestamp DEFAULT CURRENT_TIMESTAMP NULL,
	CONSTRAINT pg_log_pkey PRIMARY KEY (log_id)
);

-- Permissions

ALTER TABLE msggateway.msg_log OWNER TO msggateway_admin;
GRANT ALL ON TABLE msggateway.msg_log TO msggateway_admin;
GRANT SELECT ON TABLE msggateway.msg_log TO msggateway_ro;
GRANT INSERT, UPDATE, DELETE, SELECT ON TABLE msggateway.msg_log TO msggateway_rw;


-- msggateway.msg_provider definition

-- Drop table

-- DROP TABLE msggateway.msg_provider;

CREATE TABLE msggateway.msg_provider (
	provider_id serial4 NOT NULL,
	provider_name varchar NULL,
	short_name varchar NULL,
	services varchar NULL,
	configuration_key json NULL,
	status_cd int4 NULL,
	sms_charge numeric(25, 3) NULL,
	transit_charge numeric(25, 3) NULL,
	gst numeric(25, 2) NULL,
	threshold int4 NULL,
	nodal_officer_number int8 NULL,
	nodal_officer_name varchar NULL,
	CONSTRAINT pg_provider_pkey PRIMARY KEY (provider_id)
);
CREATE UNIQUE INDEX idx_msg_provider_provider_id ON msggateway.msg_provider USING btree (provider_id);
CREATE INDEX idx_msg_provider_provider_name ON msggateway.msg_provider USING btree (provider_name);
CREATE UNIQUE INDEX idx_msg_provider_provider_name1 ON msggateway.msg_provider USING btree (provider_name);

-- Permissions

ALTER TABLE msggateway.msg_provider OWNER TO msggateway_admin;
GRANT ALL ON TABLE msggateway.msg_provider TO msggateway_admin;
GRANT SELECT ON TABLE msggateway.msg_provider TO msggateway_ro;
GRANT INSERT, UPDATE, DELETE, SELECT ON TABLE msggateway.msg_provider TO msggateway_rw;


-- msggateway.msg_request definition

-- Drop table

-- DROP TABLE msggateway.msg_request;

CREATE TABLE msggateway.msg_request (
	request_id int4 DEFAULT nextval('msggateway.msg_request_req_id_seq'::regclass) NOT NULL,
	application_id varchar NULL,
	communication_id bpchar(20) DEFAULT msggateway.generate_random_string(20) NULL,
	facility_id varchar(13) NULL,
	priority int4 NULL,
	message_text varchar NULL,
	sender_id varchar NULL,
	entity_id varchar NULL,
	template_id varchar NULL,
	gateway varchar NULL,
	status varchar NULL,
	remarks varchar NULL,
	reference_id varchar NULL,
	response_code varchar NULL,
	response_message varchar NULL,
	complete_response varchar NULL,
	created_date timestamp DEFAULT CURRENT_TIMESTAMP NULL,
	updated_date timestamp NULL,
	mobile_number _int8 NULL,
	CONSTRAINT msg_indent_pkey_new PRIMARY KEY (request_id)
);
CREATE INDEX idx_msg_request_communication_id ON msggateway.msg_request USING btree (communication_id);
CREATE INDEX idx_msg_request_created_date ON msggateway.msg_request USING btree (created_date);
CREATE INDEX idx_msg_request_req_id ON msggateway.msg_request USING btree (request_id);

-- Permissions

ALTER TABLE msggateway.msg_request OWNER TO msggateway_admin;
GRANT ALL ON TABLE msggateway.msg_request TO msggateway_admin;
GRANT SELECT ON TABLE msggateway.msg_request TO msggateway_ro;
GRANT INSERT, UPDATE, DELETE, SELECT ON TABLE msggateway.msg_request TO msggateway_rw;


-- msggateway.msg_request_type definition

-- Drop table

-- DROP TABLE msggateway.msg_request_type;

CREATE TABLE msggateway.msg_request_type (
	request_code int8 NOT NULL,
	request_type bpchar(50) NULL,
	CONSTRAINT msg_request_type_pkey PRIMARY KEY (request_code)
);

-- Permissions

ALTER TABLE msggateway.msg_request_type OWNER TO msggateway_admin;
GRANT ALL ON TABLE msggateway.msg_request_type TO msggateway_admin;
GRANT SELECT ON TABLE msggateway.msg_request_type TO msggateway_ro;
GRANT INSERT, UPDATE, DELETE, SELECT ON TABLE msggateway.msg_request_type TO msggateway_rw;


-- msggateway.msg_template definition

-- Drop table

-- DROP TABLE msggateway.msg_template;

CREATE TABLE msggateway.msg_template (
	template_local_id serial4 NOT NULL,
	application_id varchar NULL,
	template_name varchar NULL,
	template_format varchar NULL,
	sender_id varchar NULL,
	entity_id varchar NULL,
	template_id varchar NULL,
	gateway varchar NULL,
	status_cd int4 NULL,
	created_date timestamp DEFAULT CURRENT_TIMESTAMP NULL,
	message_type varchar(2) NULL,
	CONSTRAINT mg_templates_pkey PRIMARY KEY (template_local_id)
);
CREATE UNIQUE INDEX idx_msg_template_template_id ON msggateway.msg_template USING btree (template_id);

-- Permissions

ALTER TABLE msggateway.msg_template OWNER TO msggateway_admin;
GRANT ALL ON TABLE msggateway.msg_template TO msggateway_admin;
GRANT SELECT ON TABLE msggateway.msg_template TO msggateway_ro;
GRANT INSERT, UPDATE, DELETE, SELECT ON TABLE msggateway.msg_template TO msggateway_rw;



-- DROP FUNCTION msggateway.generate_random_string(int4);

CREATE OR REPLACE FUNCTION msggateway.generate_random_string(length integer)
 RETURNS text
 LANGUAGE plpgsql
AS $function$
DECLARE
    charset text := 'abcdefghijklmnopqrstuvwxyz0123456789';
    result text := '';
    i integer;
BEGIN
    FOR i IN 1..length LOOP
        result := result || substr(charset, floor(random() * length(charset) + 1)::integer, 1);
    END LOOP;
    RETURN result;
END;
$function$
;

-- Permissions

ALTER FUNCTION msggateway.generate_random_string(int4) OWNER TO postgres;
GRANT ALL ON FUNCTION msggateway.generate_random_string(int4) TO postgres;

-- DROP FUNCTION msggateway.uuid_generate_v1();

CREATE OR REPLACE FUNCTION msggateway.uuid_generate_v1()
 RETURNS uuid
 LANGUAGE c
 PARALLEL SAFE STRICT
AS '$libdir/uuid-ossp', $function$uuid_generate_v1$function$
;

-- Permissions

ALTER FUNCTION msggateway.uuid_generate_v1() OWNER TO postgres;
GRANT ALL ON FUNCTION msggateway.uuid_generate_v1() TO postgres;

-- DROP FUNCTION msggateway.uuid_generate_v1mc();

CREATE OR REPLACE FUNCTION msggateway.uuid_generate_v1mc()
 RETURNS uuid
 LANGUAGE c
 PARALLEL SAFE STRICT
AS '$libdir/uuid-ossp', $function$uuid_generate_v1mc$function$
;

-- Permissions

ALTER FUNCTION msggateway.uuid_generate_v1mc() OWNER TO postgres;
GRANT ALL ON FUNCTION msggateway.uuid_generate_v1mc() TO postgres;

-- DROP FUNCTION msggateway.uuid_generate_v3(uuid, text);

CREATE OR REPLACE FUNCTION msggateway.uuid_generate_v3(namespace uuid, name text)
 RETURNS uuid
 LANGUAGE c
 IMMUTABLE PARALLEL SAFE STRICT
AS '$libdir/uuid-ossp', $function$uuid_generate_v3$function$
;

-- Permissions

ALTER FUNCTION msggateway.uuid_generate_v3(uuid, text) OWNER TO postgres;
GRANT ALL ON FUNCTION msggateway.uuid_generate_v3(uuid, text) TO postgres;

-- DROP FUNCTION msggateway.uuid_generate_v4();

CREATE OR REPLACE FUNCTION msggateway.uuid_generate_v4()
 RETURNS uuid
 LANGUAGE c
 PARALLEL SAFE STRICT
AS '$libdir/uuid-ossp', $function$uuid_generate_v4$function$
;

-- Permissions

ALTER FUNCTION msggateway.uuid_generate_v4() OWNER TO postgres;
GRANT ALL ON FUNCTION msggateway.uuid_generate_v4() TO postgres;

-- DROP FUNCTION msggateway.uuid_generate_v5(uuid, text);

CREATE OR REPLACE FUNCTION msggateway.uuid_generate_v5(namespace uuid, name text)
 RETURNS uuid
 LANGUAGE c
 IMMUTABLE PARALLEL SAFE STRICT
AS '$libdir/uuid-ossp', $function$uuid_generate_v5$function$
;

-- Permissions

ALTER FUNCTION msggateway.uuid_generate_v5(uuid, text) OWNER TO postgres;
GRANT ALL ON FUNCTION msggateway.uuid_generate_v5(uuid, text) TO postgres;

-- DROP FUNCTION msggateway.uuid_nil();

CREATE OR REPLACE FUNCTION msggateway.uuid_nil()
 RETURNS uuid
 LANGUAGE c
 IMMUTABLE PARALLEL SAFE STRICT
AS '$libdir/uuid-ossp', $function$uuid_nil$function$
;

-- Permissions

ALTER FUNCTION msggateway.uuid_nil() OWNER TO postgres;
GRANT ALL ON FUNCTION msggateway.uuid_nil() TO postgres;

-- DROP FUNCTION msggateway.uuid_ns_dns();

CREATE OR REPLACE FUNCTION msggateway.uuid_ns_dns()
 RETURNS uuid
 LANGUAGE c
 IMMUTABLE PARALLEL SAFE STRICT
AS '$libdir/uuid-ossp', $function$uuid_ns_dns$function$
;

-- Permissions

ALTER FUNCTION msggateway.uuid_ns_dns() OWNER TO postgres;
GRANT ALL ON FUNCTION msggateway.uuid_ns_dns() TO postgres;

-- DROP FUNCTION msggateway.uuid_ns_oid();

CREATE OR REPLACE FUNCTION msggateway.uuid_ns_oid()
 RETURNS uuid
 LANGUAGE c
 IMMUTABLE PARALLEL SAFE STRICT
AS '$libdir/uuid-ossp', $function$uuid_ns_oid$function$
;

-- Permissions

ALTER FUNCTION msggateway.uuid_ns_oid() OWNER TO postgres;
GRANT ALL ON FUNCTION msggateway.uuid_ns_oid() TO postgres;

-- DROP FUNCTION msggateway.uuid_ns_url();

CREATE OR REPLACE FUNCTION msggateway.uuid_ns_url()
 RETURNS uuid
 LANGUAGE c
 IMMUTABLE PARALLEL SAFE STRICT
AS '$libdir/uuid-ossp', $function$uuid_ns_url$function$
;

-- Permissions

ALTER FUNCTION msggateway.uuid_ns_url() OWNER TO postgres;
GRANT ALL ON FUNCTION msggateway.uuid_ns_url() TO postgres;

-- DROP FUNCTION msggateway.uuid_ns_x500();

CREATE OR REPLACE FUNCTION msggateway.uuid_ns_x500()
 RETURNS uuid
 LANGUAGE c
 IMMUTABLE PARALLEL SAFE STRICT
AS '$libdir/uuid-ossp', $function$uuid_ns_x500$function$
;

-- Permissions

ALTER FUNCTION msggateway.uuid_ns_x500() OWNER TO postgres;
GRANT ALL ON FUNCTION msggateway.uuid_ns_x500() TO postgres;


-- Permissions

GRANT ALL ON SCHEMA msggateway TO msggateway_admin;
GRANT USAGE ON SCHEMA msggateway TO msggateway_ro;
GRANT USAGE ON SCHEMA msggateway TO msggateway_rw;