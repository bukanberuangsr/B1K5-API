-- Seed data based on the capstone notebook segmentation output.
-- Default password for every seeded customer: 123456

INSERT INTO customers (id, customer_id, email, password_hash)
VALUES
    (1, 'CUS-000001', 'arna@mail.com', '$2a$10$onHqlNmwgNJ8HLCHO0FX9.NTbVqGW44Nnoz.CquKQEdDt9THdFA3y'),
    (2, 'CUS-000002', 'maya@mail.com', '$2a$10$onHqlNmwgNJ8HLCHO0FX9.NTbVqGW44Nnoz.CquKQEdDt9THdFA3y'),
    (3, 'CUS-000003', 'akhtar@mail.com', '$2a$10$onHqlNmwgNJ8HLCHO0FX9.NTbVqGW44Nnoz.CquKQEdDt9THdFA3y'),
    (4, 'CUS-000004', 'sinta@mail.com', '$2a$10$onHqlNmwgNJ8HLCHO0FX9.NTbVqGW44Nnoz.CquKQEdDt9THdFA3y'),
    (5, 'CUS-000005', 'bima@mail.com', '$2a$10$onHqlNmwgNJ8HLCHO0FX9.NTbVqGW44Nnoz.CquKQEdDt9THdFA3y')
ON CONFLICT (customer_id) DO UPDATE SET
    email = EXCLUDED.email,
    password_hash = EXCLUDED.password_hash;

INSERT INTO customer_profiles (id, customer_id, username, full_name)
VALUES
    (1, 1, 'arna', 'Arna Pratama'),
    (2, 2, 'maya', 'Maya Lestari'),
    (3, 3, 'akhtar', 'Akhtar Ramadhan'),
    (4, 4, 'sinta', 'Sinta Maharani'),
    (5, 5, 'bima', 'Bima Saputra')
ON CONFLICT (customer_id) DO UPDATE SET
    username = EXCLUDED.username,
    full_name = EXCLUDED.full_name;

INSERT INTO accounts (id, customer_id, account_no, type, balance, currency)
VALUES
    (1, 1, '1000000001', 'investment', 57384374.00, 'IDR'),
    (2, 2, '1000000002', 'savings', 5773346.00, 'IDR'),
    (3, 3, '1000000003', 'e_wallet', 46753495.00, 'IDR'),
    (4, 4, '1000000004', 'savings', 2975891.00, 'IDR'),
    (5, 5, '1000000005', 'e_wallet', 58521945.00, 'IDR')
ON CONFLICT (account_no) DO UPDATE SET
    type = EXCLUDED.type,
    balance = EXCLUDED.balance,
    currency = EXCLUDED.currency;

INSERT INTO transactions (id, account_id, trx_id, type, amount, status)
VALUES
    (1, 1, 'TRX-000001', 'investment', 2500000.00, 'success'),
    (2, 2, 'TRX-000002', 'bill_payment', 450000.00, 'success'),
    (3, 3, 'TRX-000003', 'topup', 300000.00, 'success'),
    (4, 4, 'TRX-000004', 'transfer', 150000.00, 'success'),
    (5, 5, 'TRX-000005', 'topup', 500000.00, 'success')
ON CONFLICT (trx_id) DO UPDATE SET
    type = EXCLUDED.type,
    amount = EXCLUDED.amount,
    status = EXCLUDED.status;

INSERT INTO user_activities (id, customer_id, activity_type, feature, metadata)
VALUES
    (
        1,
        1,
        'investment_active',
        'investment',
        '{"age":56,"monthly_income":21167395,"occupation":"entrepreneur","transaction_frequency_monthly":117,"transfer_ratio":0.63,"payment_ratio":0.36,"topup_ratio":0.65,"investment_activity":1,"login_frequency_weekly":37,"avg_session_duration_min":2,"recent_transaction_days":4,"dominant_activity":"investment_active","frequent_product":"e_wallet_link","customer_type":"nasabah reguler","segment":"investor"}'
    ),
    (
        2,
        2,
        'payment_heavy',
        'bill_payment',
        '{"age":46,"monthly_income":44667756,"occupation":"entrepreneur","transaction_frequency_monthly":87,"transfer_ratio":0.37,"payment_ratio":0.72,"topup_ratio":0.28,"investment_activity":0,"login_frequency_weekly":15,"avg_session_duration_min":1,"recent_transaction_days":18,"dominant_activity":"payment_heavy","frequent_product":"credit_card","customer_type":"nasabah reguler","segment":"bill_payer"}'
    ),
    (
        3,
        3,
        'topup_heavy',
        'e_wallet_link',
        '{"age":32,"monthly_income":43682512,"occupation":"employee","transaction_frequency_monthly":90,"transfer_ratio":0.30,"payment_ratio":0.20,"topup_ratio":0.60,"investment_activity":0,"login_frequency_weekly":30,"avg_session_duration_min":8,"recent_transaction_days":1,"dominant_activity":"topup_heavy","frequent_product":"e_wallet_link","customer_type":"nasabah reguler","segment":"digital_spender"}'
    ),
    (
        4,
        4,
        'low_activity',
        'mobile_app',
        '{"age":25,"monthly_income":44868484,"occupation":"student","transaction_frequency_monthly":6,"transfer_ratio":0.50,"payment_ratio":0.20,"topup_ratio":0.15,"investment_activity":0,"login_frequency_weekly":4,"avg_session_duration_min":2,"recent_transaction_days":25,"dominant_activity":"low_activity","frequent_product":"savings","customer_type":"nasabah reguler","segment":"low_activity"}'
    ),
    (
        5,
        5,
        'topup_heavy',
        'e_wallet_link',
        '{"age":38,"monthly_income":36422146,"occupation":"employee","transaction_frequency_monthly":105,"transfer_ratio":0.23,"payment_ratio":0.37,"topup_ratio":0.67,"investment_activity":0,"login_frequency_weekly":36,"avg_session_duration_min":9,"recent_transaction_days":2,"dominant_activity":"topup_heavy","frequent_product":"e_wallet_link","customer_type":"nasabah prioritas","segment":"digital_spender"}'
    )
ON CONFLICT (id) DO UPDATE SET
    customer_id = EXCLUDED.customer_id,
    activity_type = EXCLUDED.activity_type,
    feature = EXCLUDED.feature,
    metadata = EXCLUDED.metadata;

INSERT INTO analytics_events (id, customer_id, event_type, feature, metadata)
VALUES
    (
        1,
        1,
        'recommendation_generated',
        'personalization',
        '{"segment":"investor","recommendation":"promo reksa dana & deposito","why_this_recommendation":"Profil Anda menunjukkan ketertarikan pada produk investasi. Dengan saldo rata-rata yang Anda miliki, produk reksa dana dan deposito kami bisa menjadi pilihan untuk mengoptimalkan pertumbuhan aset Anda.","simulated_ctr_range":[0.30,0.45]}'
    ),
    (
        2,
        2,
        'recommendation_generated',
        'personalization',
        '{"segment":"bill_payer","recommendation":"promo cashback tagihan & auto-debit","why_this_recommendation":"Anda tercatat rutin melakukan pembayaran tagihan setiap bulannya. Dengan cashback otomatis untuk setiap tagihan yang dibayar, Anda bisa mendapatkan manfaat lebih dari transaksi yang sudah biasa Anda lakukan.","simulated_ctr_range":[0.28,0.46]}'
    ),
    (
        3,
        3,
        'recommendation_generated',
        'personalization',
        '{"segment":"digital_spender","recommendation":"promo e-wallet & top-up cashback","why_this_recommendation":"Berdasarkan riwayat transaksi Anda, kami melihat Anda sering melakukan top-up dan transaksi digital. Rekomendasi ini dipilih khusus agar Anda bisa hemat lebih banyak setiap kali bertransaksi via e-wallet.","simulated_ctr_range":[0.32,0.50]}'
    ),
    (
        4,
        4,
        'recommendation_generated',
        'personalization',
        '{"segment":"low_activity","recommendation":"promo transfer gratis & aktivasi fitur dasar","why_this_recommendation":"Kami melihat Anda belum terlalu sering menggunakan fitur-fitur di OCTO Mobile. Mulai dengan fitur transfer gratis tanpa biaya admin untuk transaksi pertama setiap bulannya.","simulated_ctr_range":[0.12,0.25]}'
    ),
    (
        5,
        5,
        'recommendation_generated',
        'personalization',
        '{"segment":"digital_spender","recommendation":"promo e-wallet & top-up cashback","why_this_recommendation":"Berdasarkan riwayat transaksi Anda, kami melihat Anda sering melakukan top-up dan transaksi digital. Rekomendasi ini dipilih khusus agar Anda bisa hemat lebih banyak setiap kali bertransaksi via e-wallet.","simulated_ctr_range":[0.32,0.50]}'
    )
ON CONFLICT (id) DO UPDATE SET
    customer_id = EXCLUDED.customer_id,
    event_type = EXCLUDED.event_type,
    feature = EXCLUDED.feature,
    metadata = EXCLUDED.metadata;

INSERT INTO segments (id, name, description)
VALUES
    (1, 'investor', 'Nasabah dengan sinyal ketertarikan investasi dan saldo rata-rata tinggi.'),
    (2, 'low_activity', 'Nasabah dengan frekuensi transaksi dan penggunaan fitur yang masih rendah.'),
    (3, 'digital_spender', 'Nasabah yang aktif melakukan top-up dan transaksi digital/e-wallet.'),
    (4, 'bill_payer', 'Nasabah yang rutin melakukan pembayaran tagihan dan cocok untuk auto-debit.')
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description;

INSERT INTO user_segments (id, customer_id, segment_id, confidence)
VALUES
    (1, 1, 1, 0.91),
    (2, 2, 4, 0.88),
    (3, 3, 3, 0.90),
    (4, 4, 2, 0.82),
    (5, 5, 3, 0.93)
ON CONFLICT (id) DO UPDATE SET
    customer_id = EXCLUDED.customer_id,
    segment_id = EXCLUDED.segment_id,
    confidence = EXCLUDED.confidence;

INSERT INTO recommendations (id, segment_id, feature, reason, priority)
VALUES
    (1, 1, 'promo reksa dana & deposito', 'Profil nasabah menunjukkan ketertarikan pada produk investasi dan saldo rata-rata yang mendukung pertumbuhan aset.', 1),
    (2, 2, 'promo transfer gratis & aktivasi fitur dasar', 'Nasabah belum terlalu sering menggunakan fitur mobile banking sehingga perlu dorongan aktivasi fitur dasar.', 4),
    (3, 3, 'promo e-wallet & top-up cashback', 'Nasabah sering melakukan top-up dan transaksi digital sehingga cashback e-wallet paling relevan.', 1),
    (4, 4, 'promo cashback tagihan & auto-debit', 'Nasabah rutin melakukan pembayaran tagihan sehingga cashback dan auto-debit sesuai pola transaksi.', 2)
ON CONFLICT (id) DO UPDATE SET
    segment_id = EXCLUDED.segment_id,
    feature = EXCLUDED.feature,
    reason = EXCLUDED.reason,
    priority = EXCLUDED.priority;

SELECT setval('customers_id_seq', COALESCE((SELECT MAX(id) FROM customers), 1));
SELECT setval('customer_profiles_id_seq', COALESCE((SELECT MAX(id) FROM customer_profiles), 1));
SELECT setval('accounts_id_seq', COALESCE((SELECT MAX(id) FROM accounts), 1));
SELECT setval('transactions_id_seq', COALESCE((SELECT MAX(id) FROM transactions), 1));
SELECT setval('user_activities_id_seq', COALESCE((SELECT MAX(id) FROM user_activities), 1));
SELECT setval('analytics_events_id_seq', COALESCE((SELECT MAX(id) FROM analytics_events), 1));
SELECT setval('segments_id_seq', COALESCE((SELECT MAX(id) FROM segments), 1));
SELECT setval('user_segments_id_seq', COALESCE((SELECT MAX(id) FROM user_segments), 1));
SELECT setval('recommendations_id_seq', COALESCE((SELECT MAX(id) FROM recommendations), 1));
