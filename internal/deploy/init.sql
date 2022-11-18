BEGIN;

CREATE SCHEMA avito;

CREATE TABLE avito.balances
(
    user_id    VARCHAR(255) PRIMARY KEY,
    value      integer   NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT avito_balances_value_positive CHECK (balances.value >= 0)
);

CREATE TABLE avito.operations
(
    user_id        VARCHAR(255) NOT NULL,
    service_id     VARCHAR(255),
    order_id       VARCHAR(255) NOT NULL,
    operation_type integer      NOT NULL,
    value          integer      NOT NULL,
    reserve        integer      NOT NULL DEFAULT 0,
    created_at     TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMP    NOT NULL DEFAULT NOW(),
    CONSTRAINT avito_operations_value_positive CHECK (operations.value >= 0),
    CONSTRAINT avito_operations_reserve_positive CHECK (operations.reserve >= 0),
    PRIMARY KEY (user_id, service_id, order_id)
);

COMMIT;