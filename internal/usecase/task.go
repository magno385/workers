package usecase

import (
	"context"
	"errors"
	"fmt"
	"tt-newbotsacc/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TaskUseCase struct {
	pool *pgxpool.Pool
	queries *repository.Queries
}

func NewTaskUseCase(pool *pgxpool.Pool) *TaskUseCase {
	queries := repository.New(pool)
	return &TaskUseCase{
		pool: pool,
		queries: queries,
	}
}

func (uc *TaskUseCase) ClaimNewTask(ctx context.Context) (*repository.GetPendingTaskRow, error) {

	tx, err := uc.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	qtx := uc.queries.WithTx(tx)

	task, err := qtx.GetPendingTask(ctx) 
	if errors.Is(err, pgx.ErrNoRows) {
    	fmt.Println("None of task in queue")
    	return nil, nil
	} else if err != nil {
		return nil, err
	}

	updStatus := repository.UpdateTaskStatusParams{
		Status: repository.NullTaskStatus{TaskStatus: repository.TaskStatusProcessing, Valid: true},
		ErrorMessage: pgtype.Text{Valid: false},
		ID: task.ID,
	}

	_, err = qtx.UpdateTaskStatus(ctx, updStatus)
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	task.Status = repository.NullTaskStatus{TaskStatus: repository.TaskStatusProcessing, Valid: true}

	return &task, nil
}

func (uc *TaskUseCase) CompleteTask(ctx context.Context, id pgtype.UUID) error {
	updStatus := repository.UpdateTaskStatusParams{
		Status: repository.NullTaskStatus{TaskStatus: repository.TaskStatusCompleted, Valid: true},
		ErrorMessage: pgtype.Text{Valid: false},
		ID: id,
	}

	_, err := uc.queries.UpdateTaskStatus(ctx, updStatus)
	if err != nil {
		return err
	} 

	return nil
}

func (uc *TaskUseCase) FailTask(ctx context.Context, id pgtype.UUID, errMsg string) error {
	updStatus := repository.UpdateTaskStatusParams{
		Status: repository.NullTaskStatus{TaskStatus: repository.TaskStatusFailed, Valid: true},
		ErrorMessage: pgtype.Text{String: errMsg, Valid: true},
		ID: id,
	}

	_, err := uc.queries.UpdateTaskStatus(ctx,updStatus)
	if err != nil {
		return err
	}

	return nil
}