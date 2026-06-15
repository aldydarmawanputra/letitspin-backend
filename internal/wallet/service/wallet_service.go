package service

import (
	"context"
	"errors"
	"fmt"

	"let-it-spin/internal/model"
	"let-it-spin/internal/wallet/repository"

	"github.com/google/uuid"
)

type WalletService struct {
	walletRepo *repository.WalletRepository
}

func NewWalletService(walletRepo *repository.WalletRepository) *WalletService {
	return &WalletService{
		walletRepo: walletRepo,
	}
}

type DepositRequest struct {
	UserID      uuid.UUID
	Amount      int64
	ReferenceID *string
	Description *string
}

type WithdrawRequest struct {
	UserID      uuid.UUID
	Amount      int64
	ReferenceID *string
	Description *string
}

func (s *WalletService) Deposit(ctx context.Context, req DepositRequest) (*model.Wallet, *model.Transaction, error) {
	if req.Amount <= 0 {
		return nil, nil, errors.New("amount must be greater than 0")
	}

	tx, err := s.walletRepo.BeginTx(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer s.walletRepo.RollbackTx(tx)

	wallet, err := s.walletRepo.GetOrCreateWallet(ctx, req.UserID)
	if err != nil {
		return nil, nil, err
	}

	balanceBefore := wallet.Balance
	balanceAfter := wallet.Balance + req.Amount

	err = s.walletRepo.UpdateBalance(ctx, tx, wallet.ID, balanceAfter)
	if err != nil {
		return nil, nil, err
	}

	transaction := &model.Transaction{
		ID:            uuid.New(),
		UserID:        req.UserID,
		WalletID:      wallet.ID,
		Amount:        req.Amount,
		Type:          model.TransactionTypeDeposit,
		ReferenceID:   req.ReferenceID,
		Description:   req.Description,
		BalanceBefore: balanceBefore,
		BalanceAfter:  balanceAfter,
	}

	err = s.walletRepo.CreateTransaction(ctx, tx, transaction)
	if err != nil {
		return nil, nil, err
	}

	if err := s.walletRepo.CommitTx(tx); err != nil {
		return nil, nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	wallet.Balance = balanceAfter
	return wallet, transaction, nil
}

func (s *WalletService) Withdraw(ctx context.Context, req WithdrawRequest) (*model.Wallet, *model.Transaction, error) {
	if req.Amount <= 0 {
		return nil, nil, errors.New("amount must be greater than 0")
	}

	tx, err := s.walletRepo.BeginTx(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer s.walletRepo.RollbackTx(tx)

	wallet, err := s.walletRepo.GetWalletByUserID(ctx, req.UserID)
	if err != nil {
		return nil, nil, err
	}

	if wallet.Balance < req.Amount {
		return nil, nil, errors.New("insufficient balance")
	}

	balanceBefore := wallet.Balance
	balanceAfter := wallet.Balance - req.Amount

	err = s.walletRepo.UpdateBalance(ctx, tx, wallet.ID, balanceAfter)
	if err != nil {
		return nil, nil, err
	}

	transaction := &model.Transaction{
		ID:            uuid.New(),
		UserID:        req.UserID,
		WalletID:      wallet.ID,
		Amount:        -req.Amount,
		Type:          model.TransactionTypeWithdraw,
		ReferenceID:   req.ReferenceID,
		Description:   req.Description,
		BalanceBefore: balanceBefore,
		BalanceAfter:  balanceAfter,
	}

	err = s.walletRepo.CreateTransaction(ctx, tx, transaction)
	if err != nil {
		return nil, nil, err
	}

	if err := s.walletRepo.CommitTx(tx); err != nil {
		return nil, nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	wallet.Balance = balanceAfter
	return wallet, transaction, nil
}

func (s *WalletService) GetBalance(ctx context.Context, userID uuid.UUID) (*model.Wallet, error) {
	wallet, err := s.walletRepo.GetWalletByUserID(ctx, userID)
	if err != nil {
		return s.walletRepo.GetOrCreateWallet(ctx, userID)
	}
	return wallet, nil
}

func (s *WalletService) GetTransactions(ctx context.Context, userID uuid.UUID, page, limit int) ([]model.Transaction, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	transactions, err := s.walletRepo.GetTransactions(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return transactions, len(transactions), nil
}

func (s *WalletService) UpdateBalanceDirect(ctx context.Context, userID uuid.UUID, newBalance int64) error {
	wallet, err := s.walletRepo.GetWalletByUserID(ctx, userID)
	if err != nil {
		return err
	}

	tx, err := s.walletRepo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer s.walletRepo.RollbackTx(tx)

	err = s.walletRepo.UpdateBalance(ctx, tx, wallet.ID, newBalance)
	if err != nil {
		return err
	}

	if err := s.walletRepo.CommitTx(tx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
