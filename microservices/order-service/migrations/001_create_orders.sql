IF NOT EXISTS (SELECT * FROM sysobjects WHERE name='orders' AND xtype='U')
BEGIN
    CREATE TABLE orders (
        id           NVARCHAR(36)   NOT NULL PRIMARY KEY,
        user_id      NVARCHAR(36)   NOT NULL,
        items        NVARCHAR(MAX)  NOT NULL,
        total_amount DECIMAL(18,2)  NOT NULL,
        status       NVARCHAR(20)   NOT NULL DEFAULT 'pending',
        notes        NVARCHAR(500)  NULL,
        created_at   DATETIME2      NOT NULL DEFAULT GETDATE(),
        updated_at   DATETIME2      NOT NULL DEFAULT GETDATE()
    );

    CREATE INDEX idx_orders_user_id   ON orders(user_id);
    CREATE INDEX idx_orders_status    ON orders(status);
    CREATE INDEX idx_orders_created   ON orders(created_at DESC);
END
