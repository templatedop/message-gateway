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