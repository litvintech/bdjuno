package consensus

import (
	"fmt"

	"github.com/forbole/bdjuno/database"

	"github.com/rs/zerolog/log"
	tmtypes "github.com/tendermint/tendermint/types"
)

func HandleGenesis(genesisDoc *tmtypes.GenesisDoc, db *database.Db) error {
	log.Debug().Str("module", "consensus").Msg("parsing genesis")

	// Save the genesis time
	err := db.SaveGenesisData(genesisDoc)
	if err != nil {
		return fmt.Errorf("error while storing genesis time: %s", err)
	}

	return nil
}
