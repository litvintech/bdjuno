package auth

import (
        "github.com/forbole/bdjuno/database"
        authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)


// FastSync downloads the x/auth state at the given height, and stores it inside the database
func FastSync(height int64, client authtypes.QueryClient, db *database.Db) error {
        return nil
}

