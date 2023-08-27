package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/Xacor/gophermart/internal/entity"
	"go.uber.org/zap"
)

type WithdrawUseCase struct {
	withdrawalsRepo WithdrawalsRepo
	balanceRepo     BalanceRepo
	l               *zap.Logger
}

var ErrInsufficientBalance = errors.New("insufficient balance")

func NewWithdrawUseCase(withdrawalsRepo WithdrawalsRepo, balanceRepo BalanceRepo, logger *zap.Logger) *WithdrawUseCase {
	return &WithdrawUseCase{withdrawalsRepo, balanceRepo, logger}
}

func (w *WithdrawUseCase) Withdraw(ctx context.Context, withdraw entity.Withdraw, userID int) error {
	balance, err := w.balanceRepo.Get(ctx, userID)
	if err != nil {
		return fmt.Errorf("error get user balance: %v", err)
	}
	if balance.Current < withdraw.Sum {
		return ErrInsufficientBalance
	}

	return w.withdrawalsRepo.Create(ctx, withdraw)
}

func (w *WithdrawUseCase) ListWithdrawals(ctx context.Context, userID int) ([]entity.Withdraw, error) {
	return w.withdrawalsRepo.GetByUserID(ctx, userID)
}
