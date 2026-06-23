package evmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"cosmossdk.io/math"
	"github.com/gorilla/mux"

	"github.com/cosmos/evm/x/epixmint/types"

	"github.com/cosmos/cosmos-sdk/client"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

const (
	displayDenom = "epix"
	baseDenom    = "aepix"
	epixDecimals = 18
)

// aepixPerEpix is 10^18 — conversion factor from base denom (aepix) to display denom (epix).
var aepixPerEpix = math.NewInt(1000000000000000000)

// supplyJSONResponse is the JSON envelope returned by the /supply/* endpoints.
// Shape matches CoinGecko's expected supply schema: a denom-tagged decimal string
// rendered at full precision (18 decimals).
type supplyJSONResponse struct {
	Denom    string `json:"denom"`
	Decimals int    `json:"decimals"`
	Amount   string `json:"amount"`     // full-precision decimal EPIX (e.g. "39500000.123456789012345678")
	AmountAepix string `json:"amount_aepix"` // raw base-denom integer for clients that prefer it
}

// formatEpix renders an aepix integer as a full-precision EPIX decimal string.
// e.g. 1500000000000000000 -> "1.500000000000000000"
func formatEpix(aepix math.Int) string {
	if aepix.IsNegative() {
		aepix = math.ZeroInt()
	}
	whole := aepix.Quo(aepixPerEpix)
	frac := aepix.Mod(aepixPerEpix)
	// Pad fractional to 18 digits.
	return fmt.Sprintf("%s.%s", whole.String(), padLeft(frac.String(), epixDecimals))
}

func padLeft(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return strings.Repeat("0", width-len(s)) + s
}

// totalSupplyAepix queries total supply in aepix.
func totalSupplyAepix(ctx context.Context, qc types.QueryClient) (math.Int, error) {
	resp, err := qc.SupplyOf(ctx, &types.QuerySupplyOfRequest{Denom: baseDenom})
	if err != nil {
		return math.Int{}, err
	}
	return resp.Supply, nil
}

// communityPoolAepix queries the distribution community pool's aepix balance, truncated to Int.
func communityPoolAepix(ctx context.Context, dc distrtypes.QueryClient) (math.Int, error) {
	resp, err := dc.CommunityPool(ctx, &distrtypes.QueryCommunityPoolRequest{})
	if err != nil {
		return math.Int{}, err
	}
	return resp.Pool.AmountOf(baseDenom).TruncateInt(), nil
}

// circulatingSupplyAepix = total supply - community pool, floored at zero.
func circulatingSupplyAepix(ctx context.Context, qc types.QueryClient, dc distrtypes.QueryClient) (math.Int, error) {
	total, err := totalSupplyAepix(ctx, qc)
	if err != nil {
		return math.Int{}, err
	}
	pool, err := communityPoolAepix(ctx, dc)
	if err != nil {
		return math.Int{}, err
	}
	circ := total.Sub(pool)
	if circ.IsNegative() {
		circ = math.ZeroInt()
	}
	return circ, nil
}

// SupplyAPIHandler creates a handler for simple supply queries compatible with trackers.
// Supports query parameters: ?q=totalcoins, ?q=circulatingsupply, ?q=maxsupply.
// Returns plain text (whole EPIX integer) for backward compatibility.
func SupplyAPIHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		if query == "" {
			writeTextErrorResponse(w, http.StatusBadRequest, "Missing query parameter 'q'")
			return
		}

		queryClient := types.NewQueryClient(clientCtx)
		distrClient := distrtypes.NewQueryClient(clientCtx)
		ctx := context.Background()

		switch query {
		case "totalcoins":
			// xID registration fees are already removed via bank BurnCoins, so they do not
			// need to be subtracted here.
			total, err := totalSupplyAepix(ctx, queryClient)
			if err != nil {
				writeTextErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to query supply: %s", err.Error()))
				return
			}
			writeText(w, total.Quo(aepixPerEpix).String())

		case "circulatingsupply":
			circ, err := circulatingSupplyAepix(ctx, queryClient, distrClient)
			if err != nil {
				writeTextErrorResponse(w, http.StatusInternalServerError, err.Error())
				return
			}
			writeText(w, circ.Quo(aepixPerEpix).String())

		case "maxsupply":
			resp, err := queryClient.MaxSupply(ctx, &types.QueryMaxSupplyRequest{})
			if err != nil {
				writeTextErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to query max supply: %s", err.Error()))
				return
			}
			writeText(w, resp.MaxSupply.Quo(aepixPerEpix).String())

		default:
			writeTextErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Unsupported query: %s. Supported queries: totalcoins, circulatingsupply, maxsupply", query))
		}
	}
}

// SupplyJSONHandler returns supply data in CoinGecko-compatible JSON.
// kind is one of: "total", "circulating", "max".
func SupplyJSONHandler(clientCtx client.Context, kind string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		queryClient := types.NewQueryClient(clientCtx)
		distrClient := distrtypes.NewQueryClient(clientCtx)
		ctx := context.Background()

		var amountAepix math.Int
		switch kind {
		case "total":
			a, err := totalSupplyAepix(ctx, queryClient)
			if err != nil {
				writeJSONError(w, http.StatusInternalServerError, err.Error())
				return
			}
			amountAepix = a
		case "circulating":
			a, err := circulatingSupplyAepix(ctx, queryClient, distrClient)
			if err != nil {
				writeJSONError(w, http.StatusInternalServerError, err.Error())
				return
			}
			amountAepix = a
		case "max":
			resp, err := queryClient.MaxSupply(ctx, &types.QueryMaxSupplyRequest{})
			if err != nil {
				writeJSONError(w, http.StatusInternalServerError, err.Error())
				return
			}
			amountAepix = resp.MaxSupply
		default:
			writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("unknown supply kind: %s", kind))
			return
		}

		writeJSON(w, http.StatusOK, supplyJSONResponse{
			Denom:       displayDenom,
			Decimals:    epixDecimals,
			Amount:      formatEpix(amountAepix),
			AmountAepix: amountAepix.String(),
		})
	}
}

func writeText(w http.ResponseWriter, body string) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, body)
}

// writeTextErrorResponse writes a plain text error response
func writeTextErrorResponse(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(status)
	fmt.Fprintf(w, "Error: %s", message)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{"error": message})
}

// RegisterSupplyAPI registers the supply API routes.
func RegisterSupplyAPI(router *mux.Router, clientCtx client.Context) {
	// Legacy plain-text endpoint (tracker compatibility).
	router.HandleFunc("/api.dws", SupplyAPIHandler(clientCtx)).Methods("GET")

	// JSON endpoints (CoinGecko-compatible). Decimals included; full-precision amounts.
	router.HandleFunc("/supply/total", SupplyJSONHandler(clientCtx, "total")).Methods("GET")
	router.HandleFunc("/supply/circulating", SupplyJSONHandler(clientCtx, "circulating")).Methods("GET")
	router.HandleFunc("/supply/max", SupplyJSONHandler(clientCtx, "max")).Methods("GET")
}
