package inter_tx

import (
	"fmt"

	proto "github.com/gogo/protobuf/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	"github.com/defund-labs/defund/x/broker/keeper"
	"github.com/defund-labs/defund/x/broker/types"

	ibctransfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v3/modules/core/05-port/types"
	host "github.com/cosmos/ibc-go/v3/modules/core/24-host"
	ibcexported "github.com/cosmos/ibc-go/v3/modules/core/exported"
	liquiditytypes "github.com/tendermint/liquidity/x/liquidity/types"
)

var _ porttypes.IBCModule = IBCModule{}

// IBCModule implements the ICS26 interface for interchain accounts controller chains
type IBCModule struct {
	keeper keeper.Keeper
}

// NewIBCModule creates a new IBCModule given the keeper
func NewIBCModule(k keeper.Keeper) IBCModule {
	return IBCModule{
		keeper: k,
	}
}

// OnChanOpenInit performs basic validation of channel initialization.
// The channel order must be UNORDERED to match the Transfer module,
// the counterparty port identifier must be either transfer or ibcaccount,
// the channel version must be equal to the version in the types package,
// there must not be an active channel for the specfied port identifier,
// and the interchain accounts module must be able to claim the channel
// capability.
func (im IBCModule) OnChanOpenInit(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	chanCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	version string,
) error {
	return im.keeper.ClaimCapability(ctx, chanCap, host.ChannelCapabilityPath(portID, channelID))
}

// OnChanOpenTry implements the IBCModule interface
func (im IBCModule) OnChanOpenTry(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID,
	channelID string,
	chanCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	counterpartyVersion string,
) (string, error) {
	return "", nil
}

// OnChanOpenAck implements the IBCModule interface
func (im IBCModule) OnChanOpenAck(
	ctx sdk.Context,
	portID,
	channelID string,
	counterpartyChannelID string,
	counterpartyVersion string,
) error {
	return nil
}

// OnChanOpenConfirm implements the IBCModule interface
func (im IBCModule) OnChanOpenConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return nil
}

// OnChanCloseInit implements the IBCModule interface
func (im IBCModule) OnChanCloseInit(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	// Disallow user-initiated channel closing for broker channels
	return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "user cannot close channel")
}

// OnChanCloseConfirm implements the IBCModule interface
func (im IBCModule) OnChanCloseConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return nil
}

// OnRecvPacket implements the IBCModule interface. A successful acknowledgement
// is returned if the packet data is succesfully decoded and the receive application
// logic returns without error.
func (im IBCModule) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) ibcexported.Acknowledgement {
	return channeltypes.NewErrorAcknowledgement("cannot receive packet via interchain accounts authentication module")
}

// OnAcknowledgementPacket implements the IBCModule interface
func (im IBCModule) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	var ack channeltypes.Acknowledgement
	if err := channeltypes.SubModuleCdc.UnmarshalJSON(acknowledgement, &ack); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-27 packet acknowledgement: %v", err)
	}
	ctx.Logger().Info(fmt.Sprintf("Received Acknowledgement In Broker Module on Source Port %s and Dest Port %s", packet.SourcePort, packet.DestinationPort))
	return nil
}

// OnTimeoutPacket implements the IBCModule interface.
func (im IBCModule) OnTimeoutPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	return nil
}

// NegotiateAppVersion implements the IBCModule interface
func (im IBCModule) NegotiateAppVersion(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionID string,
	portID string,
	counterparty channeltypes.Counterparty,
	proposedVersion string,
) (string, error) {
	return "", nil
}

func (im IBCModule) handleMsgData(ctx sdk.Context, msgData *sdk.MsgData, packet channeltypes.Packet, ackErr bool, timeout bool) error {
	switch msgData.MsgType {
	case sdk.MsgTypeURL(&liquiditytypes.MsgSwapWithinBatch{}):
		msgResponse := &liquiditytypes.MsgSwapWithinBatch{}
		if err := proto.Unmarshal(msgData.Data, msgResponse); err != nil {
			return sdkerrors.Wrapf(types.ErrHandlingICAMsg, "cannot unmarshal send response message for swap: %s", err.Error())
		}
		err := im.keeper.HandleICASwap(ctx, msgResponse, packet, ackErr, timeout)
		if err != nil {
			return sdkerrors.Wrapf(types.ErrHandlingICAMsg, "error handling ica swap logic: %s", err.Error())
		}

		return nil
	case sdk.MsgTypeURL(&ibctransfertypes.MsgTransfer{}):
		msgResponse := &ibctransfertypes.MsgTransfer{}
		if err := proto.Unmarshal(msgData.Data, msgResponse); err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrJSONUnmarshal, "cannot unmarshal send response message for transfer: %s", err.Error())
		}
		err := im.keeper.HandleICASend(ctx, msgResponse, packet, ackErr, timeout)
		if err != nil {
			return sdkerrors.Wrapf(types.ErrHandlingICAMsg, "error handling ica transfer logic: %s", err.Error())
		}
	default:
		return nil
	}
	return nil
}
