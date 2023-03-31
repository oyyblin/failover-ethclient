# failover-ethclient

ethclient but with a failover RPC endpoint + prometheus metrics

## Note

Hobby project, failover criteria is immature.

## Usage

Environment variables:

```
- ETHEREUM_RPCNAME=nodereal
- ETHEREUM_RPCURL=https://eth-mainnet.nodereal.io/v1/<omitted>
- ETHEREUM_FAILOVERRPCNAME=alchemy
- ETHEREUM_FAILOVERRPCURL=https://eth-mainnet.g.alchemy.com/v2/<omitted>
```

Code:

```golang
cfg := ethclient.ConfigFromEnvPrefix("ethereum")
ethClient, err := ethclient.New("my-app", "ethereum", cfg)
if err != nil {
    log.Fatal().Err(err).Msg("failed to init ethclient")
}
```

You'll then be able to query the following metrics:

```
rpc_request_total{app="my-app", success="true", chain="ethereum", client="nodereal"}
rpc_request_total{app="my-app", success="false", chain="ethereum",  client="nodereal"}
rpc_request_total{app="my-app", success="true", chain="ethereum", client="alchemy"}
rpc_request_total{app="my-app", success="false", chain="ethereum",  client="alchemy"}
```
