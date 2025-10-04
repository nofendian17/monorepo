package redis

import (
	"time"
)

// WithAddrs sets the Redis server addresses
func WithAddrs(addrs []string) Option {
	return func(c *Client) {
		c.opts.Addrs = addrs
	}
}

// WithUsername sets the Redis username
func WithUsername(username string) Option {
	return func(c *Client) {
		c.opts.Username = username
	}
}

// WithPassword sets the Redis password
func WithPassword(password string) Option {
	return func(c *Client) {
		c.opts.Password = password
	}
}

// WithDB sets the Redis database number
func WithDB(db int) Option {
	return func(c *Client) {
		c.opts.DB = db
	}
}

// WithDialTimeout sets the dial timeout
func WithDialTimeout(dialTimeout time.Duration) Option {
	return func(c *Client) {
		c.opts.DialTimeout = dialTimeout
	}
}

// WithReadTimeout sets the read timeout
func WithReadTimeout(readTimeout time.Duration) Option {
	return func(c *Client) {
		c.opts.ReadTimeout = readTimeout
	}
}

// WithWriteTimeout sets the write timeout
func WithWriteTimeout(writeTimeout time.Duration) Option {
	return func(c *Client) {
		c.opts.WriteTimeout = writeTimeout
	}
}

// WithPoolSize sets the connection pool size
func WithPoolSize(poolSize int) Option {
	return func(c *Client) {
		c.opts.PoolSize = poolSize
	}
}
