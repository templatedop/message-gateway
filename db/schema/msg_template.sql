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