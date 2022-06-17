CREATE TABLE triggers
(
    user_id      bigint REFERENCES users (id)       NOT NULL,
    token_ticker varchar REFERENCES tokens (ticker) NOT NULL,

    PRIMARY KEY (user_id, token_ticker)
);