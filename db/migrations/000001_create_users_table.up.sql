CREATE TABLE users
(
    id            bigserial,
    username      varchar     NOT NULL UNIQUE,
    password_hash varchar     NOT NULL,
    first_name    varchar     NOT NULL,
    last_name     varchar     NOT NULL,
    create_time   timestamp NOT NULL DEFAULT current_timestamp,

    CONSTRAINT users_pkey PRIMARY KEY (id),
    CONSTRAINT users_username_key UNIQUE (username),
    CONSTRAINT users_username_valid CHECK (username ~ '^[a-zA-Z0-9]{4,16}$')
);