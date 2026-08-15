package evmd

import (
	"crypto/ed25519"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	xidtypes "github.com/cosmos/evm/x/xid/types"
)

// xidAttestKeyFile is the node-local file (under <home>/config) holding this
// validator's ed25519 xID attestation private key as a 32-byte hex seed. The
// matching pubkey is registered on-chain via MsgRegisterAttestKey. Absent => this
// node does not attest (empty vote extensions).
const xidAttestKeyFile = "xid_attest_key"

// loadAttestPrivKey loads the node's ed25519 attestation key, or nil if absent.
func loadAttestPrivKey(homePath string) ed25519.PrivateKey {
	if homePath == "" {
		return nil
	}
	bz, err := os.ReadFile(filepath.Join(homePath, "config", xidAttestKeyFile))
	if err != nil {
		return nil
	}
	seed, err := hex.DecodeString(strings.TrimSpace(string(bz)))
	if err != nil || len(seed) != ed25519.SeedSize {
		return nil
	}
	return ed25519.NewKeyFromSeed(seed)
}

// registerAttestationHandlers wires the ABCI++ vote-extension handlers for xID
// digest attestation. MUST be called AFTER configureEVMMempool, since it wraps the
// EVM mempool's PrepareProposal (captured on app.evmPrepareProposal). No custom
// ProcessProposal is needed: baseapp skips the injected non-tx bytes during
// execution, and PreBlocker independently verifies every attestation signature, so
// the default accept-all ProcessProposal is safe.
func (app *EVMD) registerAttestationHandlers() {
	app.SetExtendVoteHandler(app.extendVoteHandler())
	app.SetVerifyVoteExtensionHandler(app.verifyVoteExtensionHandler())
	app.SetPrepareProposal(app.prepareProposalWithVoteExts())
}

// extendVoteHandler signs the current xID state digest with this node's attest key
// over AttestationSignBytes(chain_id, height, block_time, digest). The block_time is
// the proposed block's time (only ExtendVote receives it), carried in the payload so
// the verifier can reconstruct the sign-bytes.
func (app *EVMD) extendVoteHandler() sdk.ExtendVoteHandler {
	return func(ctx sdk.Context, req *abci.RequestExtendVote) (*abci.ResponseExtendVote, error) {
		empty := &abci.ResponseExtendVote{VoteExtension: []byte{}}
		if app.attestPrivKey == nil {
			return empty, nil
		}
		sd, found := app.XIDKeeper.GetStateDigest(ctx)
		if !found || sd.Digest == "" {
			return empty, nil
		}
		height := uint64(req.Height)
		blockTime := req.Time.Unix()
		msg := xidtypes.AttestationSignBytes(ctx.ChainID(), height, blockTime, sd.Digest)
		sig := ed25519.Sign(app.attestPrivKey, msg)
		pub := app.attestPrivKey.Public().(ed25519.PublicKey)
		ext := xidtypes.AttestationVoteExtension{
			Height:        height,
			BlockTime:     blockTime,
			Digest:        sd.Digest,
			Ed25519Pubkey: hex.EncodeToString(pub),
			Signature:     hex.EncodeToString(sig),
		}
		bz, err := ext.Marshal()
		if err != nil {
			return empty, nil
		}
		return &abci.ResponseExtendVote{VoteExtension: bz}, nil
	}
}

// verifyVoteExtensionHandler rejects a peer's extension only if it is malformed, or
// (for a validator that registered an attest key) its pubkey/height/digest are wrong
// or its signature fails. Deterministic: it reads only committed state + the request.
func (app *EVMD) verifyVoteExtensionHandler() sdk.VerifyVoteExtensionHandler {
	accept := &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_ACCEPT}
	reject := &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_REJECT}
	return func(ctx sdk.Context, req *abci.RequestVerifyVoteExtension) (*abci.ResponseVerifyVoteExtension, error) {
		if len(req.VoteExtension) == 0 {
			return accept, nil // empty = a non-attesting validator
		}
		var ext xidtypes.AttestationVoteExtension
		if err := ext.Unmarshal(req.VoteExtension); err != nil {
			return reject, nil
		}
		valcons := sdk.ConsAddress(req.ValidatorAddress).String()
		regPub, ok := app.XIDKeeper.GetAttestKey(ctx, valcons)
		if !ok {
			return accept, nil // unregistered: not trusted, but not our place to reject
		}
		if ext.Ed25519Pubkey != regPub || ext.Height != uint64(req.Height) {
			return reject, nil
		}
		if sd, found := app.XIDKeeper.GetStateDigest(ctx); !found || ext.Digest != sd.Digest {
			return reject, nil
		}
		if !verifyAttestSig(ctx.ChainID(), ext, regPub) {
			return reject, nil
		}
		return accept, nil
	}
}

// verifyAttestSig verifies ext's ed25519 signature over its sign-bytes using pubHex.
func verifyAttestSig(chainID string, ext xidtypes.AttestationVoteExtension, pubHex string) bool {
	pub, err := hex.DecodeString(pubHex)
	if err != nil || len(pub) != ed25519.PublicKeySize {
		return false
	}
	sig, err := hex.DecodeString(ext.Signature)
	if err != nil || len(sig) != ed25519.SignatureSize {
		return false
	}
	msg := xidtypes.AttestationSignBytes(chainID, ext.Height, ext.BlockTime, ext.Digest)
	return ed25519.Verify(pub, msg, sig)
}

// prepareProposalWithVoteExts wraps the EVM mempool's PrepareProposal, prepending
// the previous height's ExtendedCommitInfo as tx[0] when vote extensions are active.
// baseapp skips it during execution (it is not an sdk.Tx); PreBlocker consumes it.
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

// processXidVoteExtensions consumes tx[0] (the injected ExtendedCommitInfo) when
// vote extensions are active: verifies each attestation against the registered
// pubkey, uses REAL staking power (never the injected power), picks the canonical
// (>=majority-power) block_time per digest, and persists the signed attestations.
// Called from PreBlocker before ModuleManager.PreBlock.
func (app *EVMD) processXidVoteExtensions(ctx sdk.Context, txs [][]byte) {
	if len(txs) == 0 {
		return
	}
	cp := app.GetConsensusParams(ctx)
	if cp.Abci == nil || cp.Abci.VoteExtensionsEnableHeight == 0 ||
		ctx.BlockHeight() <= cp.Abci.VoteExtensionsEnableHeight {
		return // extensions not active yet -> tx[0] is a normal tx, leave it alone
	}
	var ec abci.ExtendedCommitInfo
	if err := ec.Unmarshal(txs[0]); err != nil {
		return
	}
	chainID := ctx.ChainID()

	type dk struct {
		digest string
		bt     int64
	}
	powerByDK := map[dk]uint64{}
	type verifiedExt struct {
		ext   xidtypes.AttestationVoteExtension
		power uint64
	}
	var verified []verifiedExt

	for _, vote := range ec.Votes {
		if vote.BlockIdFlag != cmtproto.BlockIDFlagCommit || len(vote.VoteExtension) == 0 {
			continue
		}
		var ext xidtypes.AttestationVoteExtension
		if err := ext.Unmarshal(vote.VoteExtension); err != nil {
			continue
		}
		valcons := sdk.ConsAddress(vote.Validator.Address)
		regPub, ok := app.XIDKeeper.GetAttestKey(ctx, valcons.String())
		if !ok || ext.Ed25519Pubkey != regPub {
			continue
		}
		if !verifyAttestSig(chainID, ext, regPub) {
			continue
		}
		validator, err := app.StakingKeeper.GetValidatorByConsAddr(ctx, valcons)
		if err != nil || validator.GetStatus() != stakingtypes.Bonded {
			continue
		}
		power := uint64(validator.GetConsensusPower(sdk.DefaultPowerReduction))
		if power == 0 {
			continue
		}
		ext.ValidatorConsAddr = valcons.String()
		verified = append(verified, verifiedExt{ext, power})
		powerByDK[dk{ext.Digest, ext.BlockTime}] += power
	}

	// Canonical block_time per digest = the (digest, block_time) with the most
	// verified power. Honest validators (>=2/3) all sign the same time (PBTS), so a
	// validator that signed a different time contributes to a losing bucket and is
	// not persisted (the client would reject it against the canonical time anyway).
	majTime := map[string]int64{}
	majPow := map[string]uint64{}
	for k, p := range powerByDK {
		if p > majPow[k.digest] {
			majPow[k.digest] = p
			majTime[k.digest] = k.bt
		}
	}
	for _, v := range verified {
		if v.ext.BlockTime == majTime[v.ext.Digest] {
			app.XIDKeeper.RecordSignedAttestation(ctx, v.ext, v.power)
		}
	}
	for digest, bt := range majTime {
		app.XIDKeeper.SetDigestBlockTime(ctx, digest, bt)
	}
}
