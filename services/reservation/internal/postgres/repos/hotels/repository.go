package hotels

import (
	"context"
	"database/sql"
	"uuid"

	"lab2/platform/observability/metrics"

	"github.com/samber/oops"
	"lab2/reservation/ent"
	"lab2/reservation/ent/hotel"
	hotels "lab2/reservation/internal/models/entities"
	hotelconverter "lab2/reservation/internal/postgres/converters/hotels"
)

type Repository struct{ client *ent.Client }

func New(client *ent.Client) *Repository { return &Repository{client: client} }

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (_ hotels.Hotel, err error) {
	defer metrics.Instrument(metrics.PreInstrument(ctx, r.Get), &err)
	h, err := r.client.Hotel.Query().Where(hotel.HotelUIDEQ(id)).Only(ctx)
	if ent.IsNotFound(err) {
		return hotels.Hotel{}, oops.In("repository").Code("not_found").
			With("entity", "hotel").
			Public("not found").
			Wrap(hotels.ErrNotFound)
	}
	if err != nil {
		return hotels.Hotel{}, oops.FromContext(ctx).In("hotels.postgres").
			With("operation", "get").
			Wrapf(err, "operation failed")
	}
	return hotelconverter.Hotel(h), nil
}

func (r *Repository) List(ctx context.Context, page, size int) (_ hotels.Page, err error) {
	defer metrics.Instrument(metrics.PreInstrument(ctx, r.List), &err)
	tx, err := r.client.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return hotels.Page{}, oops.FromContext(ctx).In("hotels.postgres").
			With("operation", "begin_snapshot").
			Wrapf(err, "operation failed")
	}
	result, queryErr := list(ctx, tx, page, size)
	if queryErr != nil {
		return hotels.Page{}, oops.In("hotels.postgres").With("operation", "list").Join(queryErr, tx.Rollback())
	}
	if err := tx.Commit(); err != nil {
		return hotels.Page{}, oops.FromContext(ctx).In("hotels.postgres").
			With("operation", "commit_snapshot").
			Wrapf(err, "operation failed")
	}
	return result, nil
}

func list(ctx context.Context, tx *ent.Tx, page, size int) (hotels.Page, error) {
	count, err := tx.Hotel.Query().Count(ctx)
	if err != nil {
		return hotels.Page{}, oops.FromContext(ctx).In("hotels.postgres").
			With("operation", "count").
			Wrapf(err, "operation failed")
	}
	rows, err := tx.Hotel.Query().Order(ent.Asc(hotel.FieldID)).Offset((page - 1) * size).Limit(size).All(ctx)
	if err != nil {
		return hotels.Page{}, oops.FromContext(ctx).In("hotels.postgres").
			With("operation", "list").
			Wrapf(err, "operation failed")
	}
	items := make([]hotels.Hotel, 0, len(rows))
	for _, h := range rows {
		items = append(items, hotelconverter.Hotel(h))
	}
	return hotels.Page{Page: page, PageSize: size, TotalElements: count, Items: items}, nil
}
