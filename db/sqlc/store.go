package db

import (
	"context"
	"database/sql"
	"fmt"
)

//Provides all functions to execute db queries and transactions
type Store struct {
	*Queries
	dataB *sql.DB
}

// creates a new store
func NewStore(dataBase *sql.DB) *Store {
	return &Store{
		dataB: dataBase,
		Queries: New(dataBase),
	}
}

// executes a function within a database transaction
func (store *Store) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.dataB.BeginTx(ctx,nil)

	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)

	if err  != nil {
		if rbErr:= tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx error: %v, rb err: %v", err, rbErr)
		}
		return err
	}
	return tx.Commit()
}

// Input paramaters of the transfer transaction
type TransferTxParams struct {
	FromAccountID int64  `json:"from_account_id"`
	ToAccountID   int64  `json:"to_account_id"`
	Amount        int64  `json:"amount"`
}

//Result of transfer transaction 
type TransferTxResult struct {
	Transfer	Transfer `json:"transfer"`
	FromAccount	Account  `json:"from_account"`
	ToAccount	Account  `json:"to_account"`
	FromEntry	Entry    `json:"from_entry"`
	ToEntry		Entry    `json:"to_entry"`
}

// Transfers performs a money transfer from one account to the other
// It creates a transfer record, add account entries, and update accounts' balance within a single database transaction
func (store *Store) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
 var result TransferTxResult

 err := store.execTx(ctx, func(q *Queries) error {
	 var err error
	 result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
		 FromAccountID: arg.FromAccountID,
		 ToAccountID: arg.ToAccountID,
		 Amount: arg.Amount,
	 })

	 if err != nil {
		 return err
	 }

	 result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
		 AccountID: arg.FromAccountID,
		 Amount: -arg.Amount,
	 })
	 if err != nil {
		 return err
	 }

	 result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
		AccountID: arg.ToAccountID,
		Amount: arg.Amount,
	})
	if err != nil {
		return err
	}

	if (arg.FromAccountID<arg.ToAccountID) {
		result.FromAccount, result.ToAccount, err = addBalance(
			ctx,
			q,
			arg.FromAccountID,
			-arg.Amount,
			arg.ToAccountID,
			arg.Amount)
		if err != nil {
			return err
		}

	}else {
		result.ToAccount, result.FromAccount, err = addBalance(
			ctx,
			q,
			arg.ToAccountID,
			arg.Amount,
			arg.FromAccountID,
			-arg.Amount)
		if err != nil {
			return err
		}
	}

	 return nil
 })
 	return result, err
}

func addBalance (
	ctx context.Context,
	q *Queries,
	account1Id int64,
	account1Amount int64,
	account2Id int64, 
	account2Amount int64) (account1 Account, account2 Account, err error){

		account1,err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
			ID: account1Id, 
			Amount: account1Amount })
		if err != nil {
			return 
		}

		account2,err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
			ID: account2Id, 
			Amount: account2Amount })
			
		return
}