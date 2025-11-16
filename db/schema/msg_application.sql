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