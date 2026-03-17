package database

import (
    "context"
    "errors"
    "fmt"
    "time"

    "github.com/jpillora/backoff"
    log "github.com/sirupsen/logrus"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/schema"
)

var (
    ErrNilDatabase = errors.New("database connection is nil")
)

type Config struct {
    DSN             string
    MaxIdleConns    int
    MaxOpenConns    int
    ConnMaxLifetime time.Duration
    ConnMaxIdleTime time.Duration
    PingInterval    time.Duration
    RetryAttempts   int
}

type DB struct {
    Conn   *gorm.DB
    ctx    context.Context
    cancel context.CancelFunc
}

func Init(cfg Config) (*DB, error) {
    ctx, cancel := context.WithCancel(context.Background())

    conn, err := openDB(cfg.DSN, cfg)
    if err != nil {
        cancel()
        return nil, fmt.Errorf("failed to connect to database: %w", err)
    }

    d := &DB{
        Conn:   conn,
        ctx:    ctx,
        cancel: cancel,
    }

    go d.checkConnection(time.NewTicker(cfg.PingInterval), cfg)

    log.Info("successfully connected to PostgreSQL database")
    return d, nil
}

func (d *DB) Close() {
    if d.cancel != nil {
        d.cancel()
    }
    if d.Conn != nil {
        db, err := d.Conn.DB()
        if err != nil {
            log.WithError(err).Error("error getting database instance while closing")
            return
        }
        if err := db.Close(); err != nil {
            log.WithError(err).Error("error closing database connection")
        }
    }
}

func (d *DB) checkConnection(ticker *time.Ticker, cfg Config) {
    defer ticker.Stop()
    for {
        select {
        case <-d.ctx.Done():
            log.Info("stopping database health check")
            return
        case <-ticker.C:
            if err := d.ping(); err != nil {
                log.WithError(err).Error("database ping failed")
                if err := d.reconnect(cfg); err != nil {
                    log.WithError(err).Error("failed to reconnect to database")
                }
            }
            d.recordMetrics()
        }
    }
}

func (d *DB) ping() error {
    if d.Conn == nil {
        return ErrNilDatabase
    }
    db, err := d.Conn.DB()
    if err != nil {
        return fmt.Errorf("getting database instance: %w", err)
    }
    start := time.Now()
    if err := db.Ping(); err != nil {
        return fmt.Errorf("pinging database: %w", err)
    }
    log.WithField("latency", time.Since(start)).Debug("database ping successful")
    return nil
}

func (d *DB) reconnect(cfg Config) error {
    b := backoff.Backoff{
        Factor: 2,
        Jitter: true,
        Min:    100 * time.Millisecond,
        Max:    1 * time.Second,
    }
    for attempt := 0; attempt < cfg.RetryAttempts; attempt++ {
        select {
        case <-d.ctx.Done():
            return d.ctx.Err()
        default:
            log.WithField("attempt", attempt+1).Info("attempting database reconnection")
            conn, err := openDB(cfg.DSN, cfg)
            if err == nil && conn != nil {
                d.Conn = conn
                log.Info("successfully reconnected to database")
                return nil
            }
            log.WithError(err).Error("reconnection attempt failed")
            time.Sleep(b.Duration())
            b.Reset()
        }
    }
    return fmt.Errorf("failed to reconnect after %d attempts", cfg.RetryAttempts)
}

func (d *DB) recordMetrics() {
    if d.Conn == nil {
        return
    }
    db, err := d.Conn.DB()
    if err != nil {
        log.WithError(err).Error("failed to get database stats")
        return
    }
    stats := db.Stats()
    log.WithFields(log.Fields{
        "maxOpenConnections": stats.MaxOpenConnections,
        "openConnections":    stats.OpenConnections,
        "inUse":              stats.InUse,
        "idle":               stats.Idle,
        "waitCount":          stats.WaitCount,
        "waitDuration":       stats.WaitDuration,
    }).Debug("database connection pool stats")
}

func openDB(dsn string, cfg Config) (*gorm.DB, error) {
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
        NamingStrategy: schema.NamingStrategy{SingularTable: false},
    })
    if err != nil {
        return nil, fmt.Errorf("opening database: %w", err)
    }
    sqlDB, err := db.DB()
    if err != nil {
        return nil, fmt.Errorf("getting database instance: %w", err)
    }
    sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
    sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
    sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
    sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
    return db, nil
}