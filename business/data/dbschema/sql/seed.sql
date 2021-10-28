INSERT INTO users (user_id, name, roles, password_hash, date_created, date_updated) VALUES
	('5cf37266-3473-4006-984f-9325122678b7', 'Admin Gopher', '{ADMIN,USER}', '$2a$10$1ggfMVZV6Js0ybvJufLRUOWHS5f6KneuP0XwwHpJ8L8ipdry9f2/a', '2019-03-24 00:00:00', '2019-03-24 00:00:00'),
	('45b5fbd3-755f-4379-8f07-a58d4a30fa2f', 'User Gopher', '{USER}', '$2a$10$9/XASPKBbJKVfCAZKDH.UuhsuALDr5vVm6VrYA9VFR8rccK86C1hW', '2019-03-24 00:00:00', '2019-03-24 00:00:00'),
	('d29b4b23-c003-4519-b3af-051b9c9b3c5a', 'Strategy Moving Average', '{USER}', '$2a$10$9/XASPKBbJKVfCAZKDH.UuhsuALDr5vVm6VrYA9VFR8rccK86C1hW', '2019-03-24 00:00:00', '2019-03-24 00:00:00')
	ON CONFLICT DO NOTHING;

INSERT INTO candles (candle_id, symbol_id, interval, open_time, open_price, close_time, close_price, high, low, volume) VALUES
    ('039eee35-7463-4dbd-ae91-0428f3b89c42', '125240c0-7f7f-4d0f-b30d-939fd93cf027', '4h', '2019-01-01 00:00:03.000001+00', 100.50, '2019-01-01 00:04:03.000001+00', 110.50, 111.00, 98.60, 13456.00),
    ('01d92444-71a7-450f-8a1b-e488cb1a6973', '125240c0-7f7f-4d0f-b30d-939fd93cf027', '4h', '2019-01-01 00:04:03.000001+00', 200.50, '2019-01-01 00:08:03.000001+00', 210.50, 211.00, 198.60, 23456.00),
    ('cd0f4919-2fe7-4808-8ba2-a1ea652cd591', '125240c0-7f7f-4d0f-b30d-939fd93cf027', '4h', '2019-01-01 00:08:03.000001+00', 200.50, '2019-01-01 00:09:03.000001+00', 310.50, 311.00, 398.60, 33456.00)
    ON CONFLICT DO NOTHING;

INSERT INTO symbols (symbol_id, symbol, status, base_asset, base_asset_precision, quote_asset, quote_precision, base_commission_precision, quote_commission_precision, iceberg_allowed, oco_allowed, quote_order_qty_market_allowed, is_spot_trading_allowed, is_margin_trading_allowed) VALUES
    ('125240c0-7f7f-4d0f-b30d-939fd93cf027', 'ETHUSDT', 'TRADING', 'ETH', 1, 'USDT', 1, 1, 1, TRUE, TRUE, TRUE, TRUE, FALSE),
    ('5f25aa33-e294-4353-92b4-246e3bacdfc7', 'BTCUSDT', 'TRADING', 'BTC', 1, 'USDT', 1, 1, 1, TRUE, TRUE, TRUE, TRUE, FALSE),
    ('97514fb4-4ff5-4561-91d1-c8da711d8f32', 'BNBUSDT', 'TRADING', 'BNB', 1, 'USDT', 1, 1, 1, TRUE, TRUE, TRUE, TRUE, FALSE)
    ON CONFLICT DO NOTHING;

INSERT INTO positions (position_id, user_id, symbol_id, creation_time, side, status) VALUES
    ('891c178b-3dbf-4f99-a8f0-99a86cb578b7', '45b5fbd3-755f-4379-8f07-a58d4a30fa2f', '125240c0-7f7f-4d0f-b30d-939fd93cf027', '2019-01-01 00:00:01.000001+00', 'SELL', 'closed'),
    ('989efd27-3da5-43ba-abf5-89dabcf4d298', '45b5fbd3-755f-4379-8f07-a58d4a30fa2f', '125240c0-7f7f-4d0f-b30d-939fd93cf027', '2019-02-01 00:00:01.000001+00', 'SELL', 'closed'),
    ('028300d6-6892-44b5-aa1b-17b8a7717ead', '45b5fbd3-755f-4379-8f07-a58d4a30fa2f', '5f25aa33-e294-4353-92b4-246e3bacdfc7', '2019-03-01 00:00:01.000001+00', 'SELL', 'open'),
    ('75fabb5c-6c22-40c6-9236-0f8017a8e12d', '5cf37266-3473-4006-984f-9325122678b7', '97514fb4-4ff5-4561-91d1-c8da711d8f32', '2019-04-01 00:00:01.000001+00', 'SELL', 'open')
    ON CONFLICT DO NOTHING;

INSERT INTO orders (order_id, symbol_id, position_id, price, quantity, status, type, side, creation_time) VALUES
    ('ef984be8-da66-4d52-b659-591b95d92591', '125240c0-7f7f-4d0f-b30d-939fd93cf027', '891c178b-3dbf-4f99-a8f0-99a86cb578b7', 1500, 2, 'FILLED', 'MARKET', 'SELL', '2019-04-01 00:00:01.000001+00'),
    ('813e5b67-d408-4271-84d3-d0587f17dae7', '125240c0-7f7f-4d0f-b30d-939fd93cf027', '891c178b-3dbf-4f99-a8f0-99a86cb578b7', 1600, 2, 'FILLED', 'MARKET', 'BUY', '2019-05-01 00:00:01.000001+00'),
    ('0e5c467e-c953-4638-a4b1-eead302d7b47', '125240c0-7f7f-4d0f-b30d-939fd93cf027', '989efd27-3da5-43ba-abf5-89dabcf4d298', 1300, 3, 'FILLED', 'MARKET', 'SELL', '2019-04-02 00:00:01.000001+00'),
    ('55d147fe-c39c-431f-9bca-3c42dd6619cd', '125240c0-7f7f-4d0f-b30d-939fd93cf027', '989efd27-3da5-43ba-abf5-89dabcf4d298', 1350, 3, 'FILLED', 'MARKET', 'BUY', '2019-04-03 00:00:01.000001+00'),
    ('9ec42f42-6413-48e8-ac65-80f6d83b9b1c', '125240c0-7f7f-4d0f-b30d-939fd93cf027', '028300d6-6892-44b5-aa1b-17b8a7717ead', 1250, 1, 'FILLED', 'MARKET', 'BUY', '2019-05-03 00:00:01.000001+00'),
    ('8a89e4ec-4b51-44ac-be9f-f15910d93682', '125240c0-7f7f-4d0f-b30d-939fd93cf027', '75fabb5c-6c22-40c6-9236-0f8017a8e12d', 1510, 4, 'FILLED', 'MARKET', 'BUY', '2019-06-03 00:00:01.000001+00')
    ON CONFLICT DO NOTHING;