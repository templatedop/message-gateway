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