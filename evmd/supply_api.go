package evmd

import (
	"context"
	"fmt"
	"net/http"

	"cosmossdk.io/math"
	"github.com/gorilla/mux"

	"github.com/cosmos/evm/x/epixmint/types"

	"github.com/cosmos/cosmos-sdk/client"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// aepixPerEpix is 10^18 — conversion factor from base denom (aepix) to display denom (epix).
var aepixPerEpix = math.NewInt(1000000000000000000)

// SupplyAPIHandler creates a handler for simple supply queries compatible with trackers
// Supports query parameters like ?q=totalcoins and ?q=circulatingsupply
func SupplyAPIHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the query parameter
		query := r.URL.Query().Get("q")
		if query == "" {
			writeTextErrorResponse(w, http.StatusBadRequest, "Missing query parameter 'q'")
			return
		}

		// Create EpixMint query client
		queryClient := types.NewQueryClient(clientCtx)

		switch query {
		case "totalcoins":
			// Total supply in EPIX (display units). xID registration fees are already
			// removed via bank BurnCoins, so they do not need to be subtracted here.
			req := &types.QuerySupplyOfRequest{
				Denom: "epix",
			}

			resp, err := queryClient.SupplyOf(context.Background(), req)
			if err != nil {
				writeTextErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to query supply: %s", err.Error()))
				return
			}

			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "%s", resp.Supply.String())

		case "circulatingsupply":
			// Circulating = total supply (aepix) - community pool (aepix), converted to EPIX.
			// Community pool funds are held by the distribution module and are not in circulation.
			supplyResp, err := queryClient.SupplyOf(context.Background(), &types.QuerySupplyOfRequest{
				Denom: "aepix",
			})
			if err != nil {
				writeTextErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to query supply: %s", err.Error()))
				return
			}

			distrClient := distrtypes.NewQueryClient(clientCtx)
			poolResp, err := distrClient.CommunityPool(context.Background(), &distrtypes.QueryCommunityPoolRequest{})
			if err != nil {
				writeTextErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to query community pool: %s", err.Error()))
				return
			}

			// Community pool is DecCoins; take the aepix amount and truncate to Int.
			poolAepix := poolResp.Pool.AmountOf("aepix").TruncateInt()

			circulatingAepix := supplyResp.Supply.Sub(poolAepix)
			if circulatingAepix.IsNegative() {
				circulatingAepix = math.ZeroInt()
			}
			circulatingEpix := circulatingAepix.Quo(aepixPerEpix)

			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "%s", circulatingEpix.String())

		case "maxsupply":
			// Get maximum supply
			req := &types.QueryMaxSupplyRequest{}

			resp, err := queryClient.MaxSupply(context.Background(), req)
			if err != nil {
				writeTextErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to query max supply: %s", err.Error()))
				return
			}

			// Convert from aepix to epix (divide by 10^18)
			epixMaxSupply := resp.MaxSupply.Quo(aepixPerEpix)

			// Return just the number as plain text
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "%s", epixMaxSupply.String())

		default:
			writeTextErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Unsupported query: %s. Supported queries: totalcoins, circulatingsupply, maxsupply", query))
		}
	}
}

// writeTextErrorResponse writes a plain text error response
func writeTextErrorResponse(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(status)
	fmt.Fprintf(w, "Error: %s", message)
}

// RegisterSupplyAPI registers the supply API routes
func RegisterSupplyAPI(router *mux.Router, clientCtx client.Context) {
	// Register the handler for supply queries
	router.HandleFunc("/api.dws", SupplyAPIHandler(clientCtx)).Methods("GET")
}
