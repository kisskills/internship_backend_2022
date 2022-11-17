BEGIN;

CREATE SCHEMA avito;

CREATE TABLE avito.balances
(
    user_id    VARCHAR(255) PRIMARY KEY,
    currency   integer   NOT NULL DEFAULT 0,
    reserve    integer   NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT avito_balances_currency_positive CHECK (currency >= 0)
);

CREATE TABLE avito.operations
(
    order_id          VARCHAR(255) PRIMARY KEY,
    user_id           VARCHAR(255) NOT NULL,
    service_id        VARCHAR(255),
    operation_type    integer      NOT NULL,
    operations_status integer      NOT NULL,
    value             integer      NOT NULL,
    created_at        TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMP    NOT NULL DEFAULT NOW(),
    CONSTRAINT avito_balances_currency_positive CHECK (value >= 0)
);

COMMIT;