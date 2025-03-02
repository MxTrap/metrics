DROP TABLE IF EXISTS schema_migration;
DROP TABLE IF EXISTS metric;
DROP TABLE IF EXISTS metric_type;

CREATE TABLE metric_type
(
    id            SERIAL       PRIMARY KEY,
    metric_type   VARCHAR(20)  NOT NULL UNIQUE
);

INSERT INTO metric_type(id, metric_type) VALUES
    (1, 'gauge'),
    (2, 'counter');

CREATE TABLE IF NOT EXISTS metric
(
    id              SERIAL PRIMARY KEY,
    metric_type_id  INT,
    metric_name     VARCHAR,
    value           DOUBLE PRECISION,
    delta           INT,
    CONSTRAINT fk_metric_metric_type
        FOREIGN KEY (metric_type_id)
        REFERENCES metric_type (id)
)