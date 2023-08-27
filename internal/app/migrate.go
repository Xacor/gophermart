package app

import (
	"context"
	"errors"
	"time"

	"github.com/Xacor/gophermart/pkg/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
)

const (
	defaultAttempts = 20
	defaultTimeout  = time.Second
)

func migrate(dbURI string, l *zap.Logger) {
	if len(dbURI) == 0 {
		l.Fatal("migrate failed", zap.Error(errors.New("environment variable or flag not declared: DATABASE_URI")))
	}
	const sql = `
		BEGIN;


CREATE TABLE IF NOT EXISTS public.users
(
    id serial,
    login character varying(256)[] NOT NULL,
    password character varying(256)[] NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS public.orders
(
    id text,
    user_id serial,
    status order_status NOT NULL,
    accrual bigint NOT NULL,
    uploaded_at timestamp with time zone NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS public.withdrawals
(
    id serial,
    order_id text,
    user_id serial NOT NULL,
    sum bigint NOT NULL DEFAULT 0,
    processed_at timestamp with time zone NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS public.balances
(
    id serial,
    current bigint,
    withdrawn bigint,
    user_id serial,
    PRIMARY KEY (id),
    UNIQUE (user_id)
);

ALTER TABLE IF EXISTS public.orders
    ADD CONSTRAINT "FK_orders_users" FOREIGN KEY (user_id)
    REFERENCES public.users (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION
    NOT VALID;


ALTER TABLE IF EXISTS public.withdrawals
    ADD CONSTRAINT "FK_withdrawals_orders" FOREIGN KEY (order_id)
    REFERENCES public.orders (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION
    NOT VALID;


ALTER TABLE IF EXISTS public.withdrawals
    ADD CONSTRAINT "FK_withdrawals_users" FOREIGN KEY (user_id)
    REFERENCES public.users (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION
    NOT VALID;


ALTER TABLE IF EXISTS public.balances
    ADD FOREIGN KEY (user_id)
    REFERENCES public.users (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION
    NOT VALID;

END;`
	var (
		attempts = defaultAttempts
		err      error
		pg       *postgres.Postgres
	)

	for attempts > 0 {
		pg, err = postgres.New(dbURI)
		if err == nil {
			l.Error("migrate connection failed", zap.Error(err), zap.Int("postgres is trying to connect, attempts left", attempts))
			continue
		}

		_, err = pg.Pool.Exec(context.Background(), sql)
		if err == nil {
			break
		}
		l.Error("migration failed", zap.Error(err))

		time.Sleep(defaultTimeout)
		attempts--
	}

	defer pg.Close()

	l.Info("migrate: up success")
}
