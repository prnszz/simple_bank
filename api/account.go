package api

import (
	"database/sql"
	"net/http"
	db "simple_bank/db/sqlc"

	"github.com/gin-gonic/gin"
)

// createAccountRequest defines the request payload for creating a new account
type createAccountRequest struct {
	Owner    string `json:"owner" binding:"required"`
	Currency string `json:"currency" binding:"required,oneof=USD EUR"`
}

// createAccount handles the request to create a new account
func (server *Server) createAccount(ctx *gin.Context) {
	var req createAccountRequest
	// Bind the JSON payload to the createAccountRequest struct
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Prepare the parameters for creating a new account
	arg := db.CreateAccountParams{
		Owner:    req.Owner,
		Currency: req.Currency,
		Balance:  0,
	}

	// Call the store to create a new account
	account, err := server.store.CreateAccount(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Respond with the created account
	ctx.JSON(http.StatusOK, account)
}


// getAccountRequest defines the request payload for retrieving an account by ID
type getAccountRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

// getAccount handles the request to retrieve an account by ID
func (server *Server) getAccount(ctx *gin.Context) {
	var req getAccountRequest
	// Bind the URI parameter to the getAccountRequest struct
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Call the store to get the account by ID
	account, err := server.store.GetAccount(ctx, req.ID)
	if err != nil {
		// If no account is found, respond with a 404 status code
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		// For other errors, respond with a 500 status code
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Respond with the retrieved account
	ctx.JSON(http.StatusOK, account)
}

// listAccountRequest defines the request payload for listing accounts with pagination
type listAccountRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

// listAccount handles the request to list accounts with pagination
func (server *Server) listAccount(ctx *gin.Context) {
	var req listAccountRequest
	// Bind the query parameters to the listAccountRequest struct
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Prepare the parameters for listing accounts
	arg := db.ListAccountsParams{
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}

	// Call the store to list the accounts
	accounts, err := server.store.ListAccounts(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Respond with the list of accounts
	ctx.JSON(http.StatusOK, accounts)
}