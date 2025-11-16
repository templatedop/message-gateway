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