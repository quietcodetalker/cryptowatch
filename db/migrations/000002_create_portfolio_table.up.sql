CREATE TABLE portfolios
(
    id      bigserial PRIMARY KEY,
    user_id bigint REFERENCES users (id) NOT NULL,
    name    varchar                      NOT NULL,

    UNIQUE (user_id, name)
);

CREATE TABLE tokens
(
    ticker varchar PRIMARY KEY,
    price  decimal(32, 16) NOT NULL DEFAULT 0.0
);

CREATE TABLE transactions
(
    id           bigserial PRIMARY KEY,
    portfolio_id bigint REFERENCES portfolios (id)  NOT NULL,
    token_ticker varchar REFERENCES tokens (ticker) NOT NULL,
    quantity     decimal(32, 16)                    NOT NULL,
    price        decimal(32, 16)                    NOT NULL,
    fee          decimal(32, 16)                    NOT NULL,
    timestamp    timestamptz                        NOT NULL DEFAULT current_timestamp
);
