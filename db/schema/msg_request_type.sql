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