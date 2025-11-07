package domain

import (
	"context"
)

// TransactionManager defines the interface for managing database transactions
// 트랜잭션 관리를 위한 인터페이스
type TransactionManager interface {
	// Begin starts a new transaction and returns a context with transaction
	// 새로운 트랜잭션을 시작하고 트랜잭션이 포함된 context를 반환
	Begin(ctx context.Context) (context.Context, error)

	// Commit commits the transaction in the context
	// context의 트랜잭션을 커밋
	Commit(ctx context.Context) error

	// Rollback rolls back the transaction in the context
	// context의 트랜잭션을 롤백
	Rollback(ctx context.Context) error

	// WithTransaction executes a function within a transaction
	// 트랜잭션 내에서 함수를 실행
	WithTransaction(ctx context.Context, fn func(context.Context) error) error
}

// TransactionalRepository extends a repository with transaction support
// 트랜잭션 지원을 추가한 Repository 확장 인터페이스
type TransactionalRepository interface {
	// WithTransaction executes repository operations within a transaction
	// 트랜잭션 내에서 repository 작업을 실행
	WithTransaction(ctx context.Context, fn func(context.Context) error) error
}
