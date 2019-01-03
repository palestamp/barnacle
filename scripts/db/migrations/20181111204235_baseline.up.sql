
CREATE SCHEMA barnacle;
CREATE SCHEMA queues;

CREATE TABLE barnacle.queue_configs (
    queue_name varchar(63) PRIMARY KEY,
    config JSONB NOT NULL
);
