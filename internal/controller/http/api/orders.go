package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/Xacor/gophermart/internal/controller/usecase"
	"github.com/Xacor/gophermart/internal/entity"
	"github.com/Xacor/gophermart/internal/utils/converter"
	"github.com/Xacor/gophermart/internal/utils/jwt"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type ordersRoutes struct {
	o usecase.Orderer
	l *zap.Logger
}

func newOrdersRoutes(handler chi.Router, o usecase.Orderer, l *zap.Logger) {
	or := &ordersRoutes{o, l}
	handler.Post("/", or.PostOrder)
	handler.Get("/", or.GetOrders)

}

func (or *ordersRoutes) PostOrder(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		or.l.Error("error read body", zap.Error(err), zap.Int("body length", len(body)))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	number := string(body)
	userID := jwt.GetUserIDFromCtx(r.Context())

	err = or.o.CreateOrder(r.Context(), number, userID)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidLuhn) {
			or.l.Error("error create order", zap.Error(err), zap.String("number", number))
			w.WriteHeader(http.StatusUnprocessableEntity)
			return

		} else if errors.Is(err, usecase.ErrAnothersOrder) {
			or.l.Error("error create order", zap.Error(err))
			w.WriteHeader(http.StatusConflict)
			return

		} else if errors.Is(err, usecase.ErrAlredyUploaded) {
			or.l.Error("error create order", zap.Error(err))
			w.WriteHeader(http.StatusOK)
			return

		}
		or.l.Error("another error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

type orderResp struct {
	Number     string        `json:"number,omitempty"`
	UserID     int           `json:"-"`
	Status     entity.Status `json:"status,omitempty"`
	Accrual    float64       `json:"accrual,omitempty"`
	UploadedAt time.Time     `json:"uploaded_at,omitempty"`
}

func (or *ordersRoutes) GetOrders(w http.ResponseWriter, r *http.Request) {
	userID := jwt.GetUserIDFromCtx(r.Context())
	orders, err := or.o.GetOrders(r.Context(), userID)
	if err != nil {
		or.l.Error("error get orders", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var ordersResp []orderResp
	for _, v := range orders {
		ordersResp = append(ordersResp, orderResp{
			Number:     v.Number,
			UserID:     v.UserID,
			Status:     v.Status,
			Accrual:    converter.IntToFloat(v.Accrual),
			UploadedAt: v.UploadedAt,
		})
	}

	json, err := json.Marshal(ordersResp)
	if err != nil {
		or.l.Error("error marshling orders", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}
