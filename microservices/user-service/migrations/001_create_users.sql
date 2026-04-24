IF NOT EXISTS (SELECT * FROM sysobjects WHERE name='users' AND xtype='U')
BEGIN
    CREATE TABLE users (
        id           NVARCHAR(36)  NOT NULL PRIMARY KEY,
        name         NVARCHAR(100) NOT NULL,
        email        NVARCHAR(255) NOT NULL UNIQUE,
        password_hash NVARCHAR(255) NOT NULL,
        role         NVARCHAR(20)  NOT NULL DEFAULT 'customer',
        active       BIT           NOT NULL DEFAULT 1,
        created_at   DATETIME2     NOT NULL DEFAULT GETDATE(),
        updated_at   DATETIME2     NOT NULL DEFAULT GETDATE()
    );

    CREATE INDEX idx_users_email  ON users(email);
    CREATE INDEX idx_users_active ON users(active);
END
