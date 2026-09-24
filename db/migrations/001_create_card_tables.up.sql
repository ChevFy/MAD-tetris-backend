CREATE TABLE guests (
    card_id VARCHAR(255) PRIMARY KEY,
    balance BIGINT NOT NULL DEFAULT 0 CHECK (balance >= 0)
);

CREATE TABLE members (
    card_id VARCHAR(255) PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE,
    balance BIGINT NOT NULL DEFAULT 0 CHECK (balance >= 0),
    points BIGINT NOT NULL DEFAULT 0 CHECK (points >= 0)
);