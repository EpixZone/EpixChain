package evmd

import (
	"encoding/hex"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/protoio"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	xidtypes "github.com/cosmos/evm/x/xid/types"
)

// registerAttestationHandlers wires the ABCI++ vote-extension handlers for xID
// digest attestation. MUST be called AFTER configureEVMMempool (it wraps the EVM
// mempool's PrepareProposal). No separate attest key and no per-validator
// registration are needed: every validator's ExtendVote payload is signed by its
// CONSENSUS key by CometBFT (the ExtensionSignature), so validators only upgrade.
// The client verifies that signature against the pinned consensus validator set.
func (app *EVMD) registerAttestationHandlers() {
	app.SetExtendVoteHandler(app.extendVoteHandler())
	app.SetVerifyVoteExtensionHandler(app.verifyVoteExtensionHandler())
	app.SetPrepareProposal(app.prepareProposalWithVoteExts())
}

// extendVoteHandler returns the current xID state digest (+ height/block_time for
// freshness) as the vote-extension payload. It does NOT sign — CometBFT signs the
// whole vote extension with the validator's consensus key.
func (app *EVMD) extendVoteHandler() sdk.ExtendVoteHandler {
	return func(ctx sdk.Context, req *abci.RequestExtendVote) (*abci.ResponseExtendVote, error) {
		empty := &abci.ResponseExtendVote{VoteExtension: []byte{}}
		sd, found := app.XIDKeeper.GetStateDigest(ctx)
		if !found || sd.Digest == "" {
			return empty, nil
		}
		ext := xidtypes.AttestationVoteExtension{
			Height:    uint64(req.Height),
			BlockTime: req.Time.Unix(),
			Digest:    sd.Digest,
		}
		bz, err := ext.Marshal()
		if err != nil {
			return empty, nil
		}
		return &abci.ResponseExtendVote{VoteExtension: bz}, nil
	}
}

// verifyVoteExtensionHandler rejects a malformed extension, or one whose digest
// isn't the current one / height isn't this vote's. Deterministic (reads only
// committed state + the request). The signature itself is verified by CometBFT.
func (app *EVMD) verifyVoteExtensionHandler() sdk.VerifyVoteExtensionHandler {
	accept := &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_ACCEPT}
	reject := &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_REJECT}
	return func(ctx sdk.Context, req *abci.RequestVerifyVoteExtension) (*abci.ResponseVerifyVoteExtension, error) {
		if len(req.VoteExtension) == 0 {
			return accept, nil
		}
		var ext xidtypes.AttestationVoteExtension
		if err := ext.Unmarshal(req.VoteExtension); err != nil {
			return reject, nil
		}
		if ext.Height != uint64(req.Height) {
			return reject, nil
		}
		if sd, found := app.XIDKeeper.GetStateDigest(ctx); !found || ext.Digest != sd.Digest {
			return reject, nil
		}
		return accept, nil
	}
}

// prepareProposalWithVoteExts wraps the EVM mempool's PrepareProposal, prepending
// the previous height's ExtendedCommitInfo as tx[0] when vote extensions are active.
// baseapp skips it during execution (not an sdk.Tx); PreBlocker consumes it.
func (app *EVMD) prepareProposalWithVoteExts() sdk.PrepareProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
		var injected [][]byte
		if len(req.LocalLastCommit.Votes) > 0 {
			if bz, err := req.LocalLastCommit.Marshal(); err == nil {
				injected = [][]byte{bz}
			}
		}
		inner := app.evmPrepareProposal
		if inner == nil {
			return &abci.ResponsePrepareProposal{Txs: append(injected, req.Txs...)}, nil
		}
		reqCopy := *req
		if len(injected) > 0 {
			reqCopy.MaxTxBytes -= int64(len(injected[0])) + 8
			if reqCopy.MaxTxBytes < 0 {
				reqCopy.MaxTxBytes = 0
			}
		}
		resp, err := inner(ctx, &reqCopy)
		if err != nil {
			return nil, err
		}
		return &abci.ResponsePrepareProposal{Txs: append(injected, resp.Txs...)}, nil
	}
}

// canonicalVoteExtSignBytes reproduces exactly what CometBFT signed for a vote
// extension: MarshalDelimited(CanonicalVoteExtension{extension, height, round,
// chain_id}). The client reproduces the same bytes to verify the signature.
func canonicalVoteExtSignBytes(extension []byte, height int64, round int64, chainID string) ([]byte, error) {
	cve := cmtproto.CanonicalVoteExtension{
		Extension: extension,
		Height:    height,
		Round:     round,
		ChainId:   chainID,
	}
	return protoio.MarshalDelimited(&cve)
}

// processXidVoteExtensions consumes tx[0] (the injected ExtendedCommitInfo) when
// vote extensions are active: verifies each validator's CometBFT ExtensionSignature
// against its CONSENSUS pubkey (mirrors baseapp.ValidateVoteExtensions), uses REAL
// staking power, picks the canonical (>=majority-power) block_time per digest, and
// persists the signed attestations. Runs in PreBlocker before ModuleManager.PreBlock.
func (app *EVMD) processXidVoteExtensions(ctx sdk.Context, txs [][]byte) {
	if len(txs) == 0 {
		return
	}
	cp := app.GetConsensusParams(ctx)
	if cp.Abci == nil || cp.Abci.VoteExtensionsEnableHeight == 0 ||
		ctx.BlockHeight() <= cp.Abci.VoteExtensionsEnableHeight {
		return // extensions not active yet -> tx[0] is a normal tx
	}
	var ec abci.ExtendedCommitInfo
	if err := ec.Unmarshal(txs[0]); err != nil {
		return
	}
	chainID := ctx.ChainID()
	signedHeight := ctx.BlockHeight() - 1 // the extensions were signed at the previous height
	round := int64(ec.Round)

	type dk struct {
		digest string
		bt     int64
	}
	powerByDK := map[dk]uint64{}
	type verifiedExt struct {
		valcons string
		pubkey  string
		digest  string
		bt      int64
		ext     []byte
		sig     []byte
		power   uint64
	}
	var verified []verifiedExt

	for _, vote := range ec.Votes {
		if vote.BlockIdFlag != cmtproto.BlockIDFlagCommit ||
			len(vote.VoteExtension) == 0 || len(vote.ExtensionSignature) == 0 {
			continue
		}
		var ext xidtypes.AttestationVoteExtension
		if err := ext.Unmarshal(vote.VoteExtension); err != nil {
			continue
		}
		valcons := sdk.ConsAddress(vote.Validator.Address)
		validator, err := app.StakingKeeper.GetValidatorByConsAddr(ctx, valcons)
		if err != nil || validator.GetStatus() != stakingtypes.Bonded {
			continue
		}
		consPub, err := validator.ConsPubKey()
		if err != nil {
			continue
		}
		signBytes, err := canonicalVoteExtSignBytes(vote.VoteExtension, signedHeight, round, chainID)
		if err != nil || !consPub.VerifySignature(signBytes, vote.ExtensionSignature) {
			continue
		}
		power := uint64(validator.GetConsensusPower(sdk.DefaultPowerReduction))
		if power == 0 {
			continue
		}
		verified = append(verified, verifiedExt{
			valcons: valcons.String(),
			pubkey:  hex.EncodeToString(consPub.Bytes()),
			digest:  ext.Digest,
			bt:      ext.BlockTime,
			ext:     vote.VoteExtension,
			sig:     vote.ExtensionSignature,
			power:   power,
		})
		powerByDK[dk{ext.Digest, ext.BlockTime}] += power
	}

	// Canonical block_time per digest = the (digest, block_time) with the most
	// verified power (honest validators, >=2/3, all sign the same time under PBTS).
	majTime := map[string]int64{}
	majPow := map[string]uint64{}
	for k, p := range powerByDK {
		if p > majPow[k.digest] {
			majPow[k.digest] = p
			majTime[k.digest] = k.bt
		}
	}
	for _, v := range verified {
		if v.bt != majTime[v.digest] {
			continue
		}
		app.XIDKeeper.RecordSignedAttestation(ctx, xidtypes.Attestation{
			ValidatorAddr:     v.valcons,
			ValidatorConsAddr: v.valcons,
			Digest:            v.digest,
			Signature:         hex.EncodeToString(v.sig),
			Ed25519Pubkey:     v.pubkey,
			Height:            uint64(signedHeight),
			Round:             uint64(round),
			VoteExtension:     v.ext,
			VotingPower:       v.power,
		})
	}
	for digest, bt := range majTime {
		app.XIDKeeper.SetDigestBlockTime(ctx, digest, bt)
	}
}
