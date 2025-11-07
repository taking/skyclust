package postgres

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"skyclust/pkg/logger"
)

// TransactionManager implements domain.TransactionManager for PostgreSQL
// PostgreSQL용 트랜잭션 관리자 구현
type TransactionManager struct {
	db *gorm.DB
}

// NewTransactionManager creates a new transaction manager
// 새로운 트랜잭션 관리자 생성
func NewTransactionManager(db *gorm.DB) *TransactionManager {
	return &TransactionManager{db: db}
}

// transactionKey is the key type for storing transaction in context
// context에 트랜잭션을 저장하기 위한 키 타입
type transactionKey struct{}

// Begin starts a new transaction and returns a context with transaction
// 새로운 트랜잭션을 시작하고 트랜잭션이 포함된 context를 반환
func (tm *TransactionManager) Begin(ctx context.Context) (context.Context, error) {
	tx := tm.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	logger.Debug("Transaction begun")
	return context.WithValue(ctx, transactionKey{}, tx), nil
}

// Commit commits the transaction in the context
// context의 트랜잭션을 커밋
func (tm *TransactionManager) Commit(ctx context.Context) error {
	tx, ok := ctx.Value(transactionKey{}).(*gorm.DB)
	if !ok {
		return fmt.Errorf("no transaction found in context")
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	logger.Debug("Transaction committed")
	return nil
}

// Rollback rolls back the transaction in the context
// context의 트랜잭션을 롤백
func (tm *TransactionManager) Rollback(ctx context.Context) error {
	tx, ok := ctx.Value(transactionKey{}).(*gorm.DB)
	if !ok {
		return fmt.Errorf("no transaction found in context")
	}

	if err := tx.Rollback().Error; err != nil {
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}

	logger.Debug("Transaction rolled back")
	return nil
}

// WithTransaction executes a function within a transaction
// 트랜잭션 내에서 함수를 실행
func (tm *TransactionManager) WithTransaction(ctx context.Context, fn func(context.Context) error) error {
	// Begin transaction
	txCtx, err := tm.Begin(ctx)
	if err != nil {
		return err
	}

	// Execute function
	err = fn(txCtx)
	if err != nil {
		// Rollback on error
		if rollbackErr := tm.Rollback(txCtx); rollbackErr != nil {
			logger.Errorf("Failed to rollback transaction: %v", rollbackErr)
		}
		return err
	}

	// Commit on success
	if err := tm.Commit(txCtx); err != nil {
		return err
	}

	return nil
}

// GetTransaction retrieves the transaction from context, or returns the original DB
// context에서 트랜잭션을 가져오거나, 없으면 원본 DB를 반환
func GetTransaction(ctx context.Context, db *gorm.DB) *gorm.DB {
	tx, ok := ctx.Value(transactionKey{}).(*gorm.DB)
	if ok && tx != nil {
		return tx
	}
	return db.WithContext(ctx)
}
