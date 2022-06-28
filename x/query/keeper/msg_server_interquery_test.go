package keeper_test

import (
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	keepertest "github.com/defund-labs/defund/testutil/keeper"
	"github.com/defund-labs/defund/x/query/keeper"
	"github.com/defund-labs/defund/x/query/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func TestInterqueryMsgServerCreate(t *testing.T) {
	k, ctx := keepertest.QueryKeeper(t)
	srv := keeper.NewMsgServerImpl(*k)
	wctx := sdk.WrapSDKContext(ctx)
	creator := "A"
	for i := 0; i < 5; i++ {
		expected := &types.MsgCreateInterquery{Creator: creator,
			Storeid: strconv.Itoa(i),
		}
		_, err := srv.CreateInterquery(wctx, expected)
		require.NoError(t, err)
		rst, found := k.GetInterquery(ctx,
			expected.Storeid,
		)
		require.True(t, found)
		require.Equal(t, expected.Creator, rst.Creator)
	}
}

func TestInterqueryMsgServerResult(t *testing.T) {
	creator := "A"

	for _, tc := range []struct {
		desc    string
		request *types.MsgCreateInterqueryResult
		err     error
	}{
		{
			desc: "Completed",
			request: &types.MsgCreateInterqueryResult{Creator: creator,
				Storeid: strconv.Itoa(0),
			},
		},
		{
			desc: "Unauthorized",
			request: &types.MsgCreateInterqueryResult{Creator: "B",
				Storeid: strconv.Itoa(0),
			},
			err: sdkerrors.ErrUnauthorized,
		},
		{
			desc: "KeyNotFound",
			request: &types.MsgCreateInterqueryResult{Creator: creator,
				Storeid: strconv.Itoa(100000),
			},
			err: sdkerrors.ErrKeyNotFound,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			k, ctx := keepertest.QueryKeeper(t)
			srv := keeper.NewMsgServerImpl(*k)
			wctx := sdk.WrapSDKContext(ctx)
			expected := &types.MsgCreateInterquery{Creator: creator,
				Storeid: strconv.Itoa(0),
			}
			_, err := srv.CreateInterquery(wctx, expected)
			require.NoError(t, err)

			_, err = srv.CreateInterqueryResult(wctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				rst, found := k.GetInterquery(ctx,
					expected.Storeid,
				)
				require.True(t, found)
				require.Equal(t, expected.Creator, rst.Creator)
			}
		})
	}
}

func TestInterqueryMsgServerTimeout(t *testing.T) {
	creator := "A"

	for _, tc := range []struct {
		desc    string
		request *types.MsgCreateInterqueryTimeout
		err     error
	}{
		{
			desc: "Completed",
			request: &types.MsgCreateInterqueryTimeout{Creator: creator,
				Storeid: strconv.Itoa(0),
			},
		},
		{
			desc: "Unauthorized",
			request: &types.MsgCreateInterqueryTimeout{Creator: "B",
				Storeid: strconv.Itoa(0),
			},
			err: sdkerrors.ErrUnauthorized,
		},
		{
			desc: "KeyNotFound",
			request: &types.MsgCreateInterqueryTimeout{Creator: creator,
				Storeid: strconv.Itoa(100000),
			},
			err: sdkerrors.ErrKeyNotFound,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			k, ctx := keepertest.QueryKeeper(t)
			srv := keeper.NewMsgServerImpl(*k)
			wctx := sdk.WrapSDKContext(ctx)

			_, err := srv.CreateInterquery(wctx, &types.MsgCreateInterquery{Creator: creator,
				Storeid: strconv.Itoa(0),
			})
			require.NoError(t, err)
			_, err = srv.CreateInterqueryTimeout(wctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				_, found := k.GetInterquery(ctx,
					tc.request.Storeid,
				)
				require.False(t, found)
			}
		})
	}
}
