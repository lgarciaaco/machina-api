INSERT INTO users (user_id, name, roles, password_hash, date_created, date_updated) VALUES
	('5cf37266-3473-4006-984f-9325122678b7', 'Admin Gopher', '{ADMIN,USER}', '$2a$10$1ggfMVZV6Js0ybvJufLRUOWHS5f6KneuP0XwwHpJ8L8ipdry9f2/a', '2019-03-24 00:00:00', '2019-03-24 00:00:00'),
	('45b5fbd3-755f-4379-8f07-a58d4a30fa2f', 'User Gopher', '{USER}', '$2a$10$9/XASPKBbJKVfCAZKDH.UuhsuALDr5vVm6VrYA9VFR8rccK86C1hW', '2019-03-24 00:00:00', '2019-03-24 00:00:00')
	ON CONFLICT DO NOTHING;

INSERT INTO candles (candle_id, symbol, interval, open_time, open_price, close_time, close_price, high, low, volume) VALUES
    ('039eee35-7463-4dbd-ae91-0428f3b89c42', 'ETHUSDT', '4h', '2019-01-01 00:00:03.000001+00', 100.50, '2019-01-01 00:04:03.000001+00', 110.50, 111.00, 98.60, 13456.00),
    ('01d92444-71a7-450f-8a1b-e488cb1a6973', 'ETHUSDT', '4h', '2019-01-01 00:04:03.000001+00', 200.50, '2019-01-01 00:08:03.000001+00', 210.50, 211.00, 198.60, 23456.00),
    ('cd0f4919-2fe7-4808-8ba2-a1ea652cd591', 'ETHUSDT', '4h', '2019-01-01 00:08:03.000001+00', 200.50, '2019-01-01 00:09:03.000001+00', 310.50, 311.00, 398.60, 33456.00)
    ON CONFLICT DO NOTHING;

INSERT INTO symbols (symbol_id, symbol, status, base_asset, base_asset_precision, quote_asset, quote_precision, base_commission_precision, quote_commission_precision, iceberg_allowed, oco_allowed, quote_order_qty_market_allowed, is_spot_trading_allowed, is_margin_trading_allowed) VALUES
    ('125240c0-7f7f-4d0f-b30d-939fd93cf027', 'ETHUSDT', 'TRADING', 'ETH', 1, 'USDT', 1, 1, 1, TRUE, TRUE, TRUE, TRUE, FALSE),
    ('5f25aa33-e294-4353-92b4-246e3bacdfc7', 'BTCUSDT', 'TRADING', 'BTC', 1, 'USDT', 1, 1, 1, TRUE, TRUE, TRUE, TRUE, FALSE),
    ('35aee552-a5bf-42a1-9d40-b6a9d4a5f342', 'XRPUSDT', 'TRADING', 'XRP', 1, 'USDT', 1, 1, 1, TRUE, TRUE, TRUE, TRUE, FALSE)
    ON CONFLICT DO NOTHING;

INSERT INTO positions (position_id, user_id, symbol_id, creation_time, side, status) VALUES
    ('891c178b-3dbf-4f99-a8f0-99a86cb578b7', '45b5fbd3-755f-4379-8f07-a58d4a30fa2f', '125240c0-7f7f-4d0f-b30d-939fd93cf027', '2019-01-01 00:00:01.000001+00', 'SELL', 'closed'),
    ('989efd27-3da5-43ba-abf5-89dabcf4d298', '45b5fbd3-755f-4379-8f07-a58d4a30fa2f', '125240c0-7f7f-4d0f-b30d-939fd93cf027', '2019-02-01 00:00:01.000001+00', 'SELL', 'closed'),
    ('028300d6-6892-44b5-aa1b-17b8a7717ead', '45b5fbd3-755f-4379-8f07-a58d4a30fa2f', '5f25aa33-e294-4353-92b4-246e3bacdfc7', '2019-03-01 00:00:01.000001+00', 'SELL', 'closed'),
    ('75fabb5c-6c22-40c6-9236-0f8017a8e12d', '5cf37266-3473-4006-984f-9325122678b7', '5f25aa33-e294-4353-92b4-246e3bacdfc7', '2019-04-01 00:00:01.000001+00', 'SELL', 'closed')
    ON CONFLICT DO NOTHING;