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