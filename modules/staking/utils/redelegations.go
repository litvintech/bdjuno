package utils

import (
	"context"
	"sync"

	"github.com/forbole/bdjuno/database"
	"github.com/forbole/bdjuno/modules/utils"
	"github.com/forbole/bdjuno/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/rs/zerolog/log"
)

// ConvertRedelegationResponse converts the given response into a slice of BDJuno redelegation objects
func ConvertRedelegationResponse(
	height int64, bondDenom string, response stakingtypes.RedelegationResponse,
) []types.Redelegation {
	var delegations []types.Redelegation
	for _, entry := range response.Entries {
		delegations = append(delegations, types.NewRedelegation(
			response.Redelegation.DelegatorAddress,
			response.Redelegation.ValidatorSrcAddress,
			response.Redelegation.ValidatorDstAddress,
			sdk.NewCoin(bondDenom, entry.Balance),
			entry.RedelegationEntry.CompletionTime,
			height,
		))
	}
	return delegations
}

// --------------------------------------------------------------------------------------------------------------------

// UpdateValidatorsRedelegations updates the redelegations for all the validators provided
func UpdateValidatorsRedelegations(
	height int64, bondDenom string, validators []stakingtypes.Validator, client stakingtypes.QueryClient, db *database.Db,
) {
	log.Debug().Str("module", "staking").Int64("height", height).
		Msg("updating validators redelegations")

	var wg sync.WaitGroup
	for _, val := range validators {
		wg.Add(1)
		go getRedelegations(val.OperatorAddress, bondDenom, height, client, db, &wg)
	}
	wg.Wait()
}

func getRedelegations(
	validatorAddress string, bondDenom string, height int64, client stakingtypes.QueryClient, db *database.Db, wg *sync.WaitGroup,
) {
	defer wg.Done()

	header := utils.GetHeightRequestHeader(height)

	var nextKey []byte
	var stop = false
	for !stop {
		res, err := client.Redelegations(
			context.Background(),
			&stakingtypes.QueryRedelegationsRequest{
				SrcValidatorAddr: validatorAddress,
				Pagination: &query.PageRequest{
					Key:   nextKey,
					Limit: 100, // Query 100 unbonding delegations at a time
				},
			},
			header,
		)
		if err != nil {
			log.Error().Str("module", "staking").Err(err).Int64("height", height).
				Msg("error while getting validators redelegations")
			return
		}

		var delegations []types.Redelegation
		for _, delegation := range res.RedelegationResponses {
			redelegations := ConvertRedelegationResponse(height, bondDenom, delegation)
			delegations = append(delegations, redelegations...)
		}
		err = db.SaveRedelegations(delegations)
		if err != nil {
			log.Error().Str("module", "staking").Err(err).Int64("height", height).
				Msg("error while saving validators redelegations")
			return
		}

		nextKey = res.Pagination.NextKey
		stop = len(res.Pagination.NextKey) == 0
	}
}

// --------------------------------------------------------------------------------------------------------------------

// DeleteRedelegation returns a function that when called removes the given redelegation from the database.
func DeleteRedelegation(redelegation types.Redelegation, db *database.Db) func() {
	return func() {
		// Remove existing redelegations
		err := db.DeleteRedelegation(redelegation)
		if err != nil {
			log.Error().Str("module", "staking").Err(err).
				Str("operation", "update redelegations").
				Msg("error while removing delegator redelegations")
			return
		}
	}
}
