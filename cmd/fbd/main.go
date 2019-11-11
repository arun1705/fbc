package main

import (
	"encoding/json"
	"io"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/cli"
	"github.com/tendermint/tendermint/libs/log"

	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/blockkungpao/fbc/app"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"

	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	"github.com/cosmos/cosmos-sdk/x/staking"

)

func main() {
	cdc := app.MakeCodec()

	// TODO: set custom bech32 prefixes for peggy
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("firstbloodaccaddr", "firstbloodaccpub")
	config.SetBech32PrefixForValidator("firstbloodvaladdr", "firstbloodvalpub")
	config.SetBech32PrefixForConsensusNode("firstbloodconsaddr", "firstbloodconspub")

	config.Seal()

	ctx := server.NewDefaultContext()
	cobra.EnableCommandSorting = false

	rootCmd := &cobra.Command{
		Use:               "fbd",
		Short:             "Ethereum Bridge App Daemon (server)",
		PersistentPreRunE: server.PersistentPreRunEFn(ctx),
	}

	rootCmd.AddCommand(genutilcli.InitCmd(ctx, cdc, app.ModuleBasics, app.DefaultNodeHome))
	rootCmd.AddCommand(genutilcli.CollectGenTxsCmd(ctx, cdc, auth.GenesisAccountIterator{}, app.DefaultNodeHome))
	rootCmd.AddCommand(genutilcli.GenTxCmd(ctx, cdc, app.ModuleBasics, staking.AppModuleBasic{},
		auth.GenesisAccountIterator{}, app.DefaultNodeHome, app.DefaultCLIHome))
	rootCmd.AddCommand(genutilcli.ValidateGenesisCmd(ctx, cdc, app.ModuleBasics))
	rootCmd.AddCommand(AddGenesisAccountCmd(ctx, cdc, app.DefaultNodeHome, app.DefaultCLIHome))

	server.AddCommands(ctx, cdc, rootCmd, newApp, exportAppStateAndTMValidators)

	// prepare and add flags
	executor := cli.PrepareBaseCmd(rootCmd, "EB", app.DefaultNodeHome)
	err := executor.Execute()
	if err != nil {
		panic(err)
	}
}

func newApp(logger log.Logger, db dbm.DB, traceStore io.Writer) abci.Application {
	return app.NewEthereumBridgeApp(
		logger, db, true,
		baseapp.SetMinGasPrices(viper.GetString(server.FlagMinGasPrices)),
		baseapp.SetHaltHeight(uint64(viper.GetInt(server.FlagHaltHeight))),
	)
}

func exportAppStateAndTMValidators(
	logger log.Logger, db dbm.DB, traceStore io.Writer, height int64, forZeroHeight bool, jailWhiteList []string,
) (json.RawMessage, []tmtypes.GenesisValidator, error) {

	if height != -1 {
		ebApp := app.NewEthereumBridgeApp(logger, db, false)
		err := ebApp.LoadHeight(height)
		if err != nil {
			return nil, nil, err
		}
		return ebApp.ExportAppStateAndValidators(forZeroHeight, jailWhiteList)
	}
	ebApp := app.NewEthereumBridgeApp(logger, db, true)
	return ebApp.ExportAppStateAndValidators(forZeroHeight, jailWhiteList)
}
