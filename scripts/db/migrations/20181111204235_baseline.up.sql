
CREATE SCHEMA barnacle;
CREATE SCHEMA queues;


CREATE TABLE barnacle.resource_configs (
    resource_id varchar(32) PRIMARY KEY,
    config JSONB NOT NULL
);


CREATE TABLE barnacle.queue_configs (
    queue_id varchar(63) PRIMARY KEY,
    resource_id varchar(32) REFERENCES barnacle.resource_configs(resource_id) NOT NULL,
    backend_type varchar(64) NOT NULL,
    queue_type varchar(64) NOT NULL,
    queue_state varchar(32) NOT NULL,
    config JSONB NOT NULL
);
