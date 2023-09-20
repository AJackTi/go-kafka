-- DROP TYPE IF EXISTS s;

-- CREATE TYPE s AS ENUM ('Doing', 'Done', 'Deleted');

-- CREATE TABLE
--     IF NOT EXISTS tasks (
--         id SERIAL,
--         title varchar(255) DEFAULT NULL,
--         name varchar(255) DEFAULT NULL,
--         image varchar(255) DEFAULT NULL,
--         description text,
--         status s default 'Doing',
--         created_at timestamp NULL DEFAULT CURRENT_TIMESTAMP,
--         updated_at timestamp NULL DEFAULT CURRENT_TIMESTAMP,
--         PRIMARY KEY (id)
--     );

CREATE TABLE
    IF NOT EXISTS tasks (
        id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
        event_type VARCHAR(50),
        data VARCHAR(500),
        created_at timestamp NULL DEFAULT CURRENT_TIMESTAMP,
        updated_at timestamp NULL DEFAULT CURRENT_TIMESTAMP
    );