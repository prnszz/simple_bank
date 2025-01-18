package db

import (
	"context"
	"database/sql"
	"fmt"
)

// Store provides all functions to execute db queries and transactions
// Because each query only do a specific task, queries don't support transactions
// we need to compose the queries to support transactions
// which we are going to use composition to do that
// all queries are composed and going to be stored in the Store struct
type Store struct {
	*Queries
	db *sql.DB // db connection
}

// NewStore creates a new Store
func NewStore(db *sql.DB) *Store {
	return &Store{
		db:      db,
		Queries: New(db),
	}
}

// execTx executes a function within a database transaction
// This make sure that the function is executed atomically
// Parameters:
//   - ctx: context for the transaction
//   - fn: the function to execute within the transaction, takes a Queries object and returns error
//
// Returns:
//   - error if the transaction fails, nil on success
//
// Note: This is an internal method (unexported) as it handles low-level transaction
// logic. External packages should use higher-level public methods that compose
// this functionality.
func (store *Store) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil) // begin transaction
	if err != nil {
		return err
	}

	q := New(tx) // create a new query with the transaction

	// apply the function with the query
	// if the function returns an error, rollback the transaction
	// if the rollback fails, return the transaction error and rollback error
	// if the function returns nil, commit the transaction
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

// TransferTxParams contains the input parameters of the transfer transaction
type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

// TransferTxResult is the result of the transfer transaction
type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

// var txKey = struct{}{}

// TransferTx performs a money transfer from one account to the other
// It creates a transfer record, add account entries,
// and update account balances within a single database transaction
func (store *Store) TransferTx(ctx context.Context, arg TransferTxParams,) (TransferTxResult, error) {
	var result TransferTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		// txName := ctx.Value(txKey)

		//1. Create a transfer record
		// fmt.Println(">> Executing transfer transaction:", txName)
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams(arg))
		if err != nil {
			return err
		}

		//2. Create the account entries
		// fmt.Println(">> Executing entry 1:", txName)
		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount,
		})
		if err != nil {
			return err
		}

		// fmt.Println(">> Executing entry 2:", txName)
		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})
		if err != nil {
			return err
		}

		//TODO: update account balance
		// the query GetAccount just a normal select query, 
		// so it doesn't block the transaction from the other queries on the same table
		// thus, we need to lock the account row using SELECT ... FOR UPDATE
		// fmt.Println(">> Get Account 1 for update", txName)
		// account1, err:= q.GetAccountForUpdate(ctx, arg.FromAccountID)
		// if err != nil {
		// 	return err
		// }

		// // fmt.Println(">> Excute update for Account 1", txName)
		// result.FromAccount, err = q.UpdateAccount(ctx, UpdateAccountParams{
		// 	ID: arg. FromAccountID,
		// 	Balance: account1.Balance - arg.Amount,
		// })
		// if err != nil {
		// 	return err
		// }

		// // fmt.Println(">> Get Account 2 for update", txName)
		// account2, err := q.GetAccountForUpdate(ctx, arg.ToAccountID)
		// if err != nil {
		// 	return err
		// }

		// // fmt.Println(">> Excute update for Account 2", txName)
		// result.ToAccount, err = q.UpdateAccount(ctx, UpdateAccountParams{
		// 	ID: arg.ToAccountID,
		// 	Balance: account2.Balance + arg.Amount,
		// })
		// if err != nil {
		// 	return err
		// }
		

		if arg.FromAccountID < arg.ToAccountID {
		result.FromAccount, result.ToAccount, err = addMoney(ctx, q, arg.FromAccountID, -arg.Amount, arg.ToAccountID, arg.Amount)	
	} else {
		result.ToAccount, result.FromAccount, err = addMoney(ctx, q, arg.ToAccountID, arg.Amount, arg.FromAccountID, -arg.Amount)	
		}
    if err != nil {
      return err
    }
		return nil
	})

	return result, err
}


func addMoney(
	ctx context.Context,
	q * Queries,
	accountID1 int64,
	amount1 int64,
	accountID2 int64,
	amount2 int64,
) (account1 Account, account2 Account, err error) {
	account1, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
    ID: accountID1,
    Amount: amount1,
  })
  if err != nil {
    return
  }

  account2, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
    ID: accountID2,
    Amount: amount2,
  })
  return
}