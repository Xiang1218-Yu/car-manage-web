-- 停车场管理系统数据库表结构

CREATE TABLE IF NOT EXISTS parking_lots (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT    NOT NULL,
    address     TEXT    NOT NULL DEFAULT '',
    total_spots INTEGER NOT NULL DEFAULT 0,
    created_at  TEXT    NOT NULL
);

CREATE TABLE IF NOT EXISTS parking_spots (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    lot_id     INTEGER NOT NULL,
    code       TEXT    NOT NULL,                       -- 车位编号，如 A01
    status     TEXT    NOT NULL DEFAULT 'available',   -- available / occupied / maintenance
    created_at TEXT    NOT NULL,
    FOREIGN KEY (lot_id) REFERENCES parking_lots(id) ON DELETE CASCADE,
    UNIQUE (lot_id, code)
);

CREATE INDEX IF NOT EXISTS idx_spots_lot ON parking_spots(lot_id);

CREATE TABLE IF NOT EXISTS vehicles (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    plate        TEXT    NOT NULL UNIQUE,              -- 车牌号
    vehicle_type TEXT    NOT NULL DEFAULT 'car',       -- car / suv / truck / motorcycle
    color        TEXT    NOT NULL DEFAULT '',
    owner_name   TEXT    NOT NULL DEFAULT '',
    owner_phone  TEXT    NOT NULL DEFAULT '',
    created_at   TEXT    NOT NULL
);

CREATE TABLE IF NOT EXISTS fee_rules (
    id                   INTEGER PRIMARY KEY AUTOINCREMENT,
    lot_id               INTEGER,                       -- NULL 表示适用于所有停车场
    name                 TEXT    NOT NULL,
    free_minutes         INTEGER NOT NULL DEFAULT 0,    -- 免费时长（分钟）
    first_block_minutes  INTEGER NOT NULL DEFAULT 60,  -- 首段时间（分钟）
    first_block_price    REAL    NOT NULL DEFAULT 0,   -- 首段价格
    unit_minutes         INTEGER NOT NULL DEFAULT 60,  -- 后续计费单位时间（分钟）
    unit_price           REAL    NOT NULL DEFAULT 0,   -- 后续单位价格
    daily_cap            REAL    NOT NULL DEFAULT 0,   -- 每日封顶金额，0 表示不封顶
    active               INTEGER NOT NULL DEFAULT 1,
    created_at           TEXT    NOT NULL,
    FOREIGN KEY (lot_id) REFERENCES parking_lots(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS parking_records (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    spot_id        INTEGER NOT NULL,
    vehicle_id     INTEGER NOT NULL,
    rule_id        INTEGER,                              -- 入场时绑定的费用规则
    check_in_time  TEXT    NOT NULL,
    check_out_time TEXT,
    fee            REAL    NOT NULL DEFAULT 0,
    status         TEXT    NOT NULL DEFAULT 'active',    -- active / completed
    created_at     TEXT    NOT NULL,
    FOREIGN KEY (spot_id)    REFERENCES parking_spots(id),
    FOREIGN KEY (vehicle_id) REFERENCES vehicles(id),
    FOREIGN KEY (rule_id)    REFERENCES fee_rules(id)
);

CREATE INDEX IF NOT EXISTS idx_records_status ON parking_records(status);
CREATE INDEX IF NOT EXISTS idx_records_spot   ON parking_records(spot_id);

-- 同一车辆任意时刻只能保留一条在场记录，避免跨入口重复入场。
CREATE UNIQUE INDEX IF NOT EXISTS idx_records_active_vehicle
    ON parking_records(vehicle_id)
    WHERE status = 'active';
