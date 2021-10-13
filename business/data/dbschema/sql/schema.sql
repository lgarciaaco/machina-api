-- Version: 1.1
-- Description: Create table users
CREATE TABLE users
(
    user_id       UUID,
    name          TEXT,
    roles         TEXT [],
    password_hash TEXT,
    date_created  TIMESTAMP,
    date_updated  TIMESTAMP,

    PRIMARY KEY (user_id)
);

-- Version: 1.2
-- Description: Create table products
CREATE TABLE products
(
    product_id   UUID,
    name         TEXT,
    cost         INT,
    quantity     INT,
    user_id      UUID,
    date_created TIMESTAMP,
    date_updated TIMESTAMP,

    PRIMARY KEY (product_id),
    FOREIGN KEY (user_id) REFERENCES users (user_id) ON DELETE CASCADE
);

-- Version: 1.3
-- Description: Create table sales
CREATE TABLE sales
(
    sale_id      UUID,
    user_id      UUID,
    product_id   UUID,
    quantity     INT,
    paid         INT,
    date_created TIMESTAMP,

    PRIMARY KEY (sale_id),
    FOREIGN KEY (user_id) REFERENCES users (user_id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES products (product_id) ON DELETE CASCADE
);

-- Version: 1.4
-- Description: Create table candles
CREATE TABLE candles
(
    candle_id   UUID,
    symbol      TEXT,
    interval    TEXT,
    open_time   TIMESTAMP,
    open_price  FLOAT,
    close_time  TIMESTAMP,
    close_price FLOAT,
    high        FLOAT,
    low         FLOAT,
    volume      FLOAT,

    PRIMARY KEY (candle_id)
);

CREATE TABLE symbols
(
    symbol_id                      UUID,
    symbol                         TEXT NOT NULL UNIQUE,
    status                         TEXT,
    base_asset                     TEXT NOT NULL,
    base_asset_precision           INT,
    quote_asset                    TEXT NOT NULL,
    quote_precision                INT,
    base_commission_precision      INT,
    quote_commission_precision     INT,
    iceberg_allowed                BOOLEAN,
    oco_allowed                    BOOLEAN,
    quote_order_qty_market_allowed BOOLEAN,
    is_spot_trading_allowed        BOOLEAN,
    is_margin_trading_allowed      BOOLEAN,

    PRIMARY KEY (symbol_id)
);

CREATE TABLE positions
(
    position_id   UUID,
    symbol_id     UUID,
    user_id       UUID,
    creation_time TIMESTAMP,
    side          TEXT,
    status        TEXT,

    PRIMARY KEY (position_id),
    FOREIGN KEY (symbol_id) REFERENCES symbols (symbol_id),
    FOREIGN KEY (user_id) REFERENCES users (user_id)
);

CREATE TABLE orders
(
    order_id      UUID,
    symbol_id     UUID,
    position_id   UUID,
    price         FLOAT,
    quantity      FLOAT,
    status        TEXT,
    type          TEXT,
    side          TEXT,
    creation_time TIMESTAMP,

    PRIMARY KEY (order_id),
    FOREIGN KEY (symbol_id) REFERENCES symbols (symbol_id),
    FOREIGN KEY (position_id) REFERENCES positions (position_id)
);