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