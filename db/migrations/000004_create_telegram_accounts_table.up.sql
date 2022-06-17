CREATE TABLE telegram_accounts
(
    id         bigint PRIMARY KEY,
    auth_token varchar UNIQUE,
    user_id    bigint UNIQUE
);