package ethclient

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type client struct {
	metrics *metrics
	logger  *zerolog.Logger
	cfg     *Config

	m *ethclient.Client // main
	b *ethclient.Client // backup
}

type Client interface {
	bind.ContractBackend
	ethereum.ChainReader
	ethereum.TransactionReader
	ethereum.ChainStateReader
	ethereum.ChainSyncReader
	ethereum.ContractCaller
	ethereum.LogFilterer
	ethereum.TransactionSender
	ethereum.GasPricer
	ethereum.PendingStateReader
	ethereum.PendingContractCaller
	ethereum.GasEstimator
}

func New(
	appName string,
	chain string,
	cfg *Config,
) (Client, error) {
	logger := zerolog.DefaultContextLogger
	if logger == nil {
		l := log.With().Caller().Logger()
		logger = &l
	}
	logger.Info().Msgf("setting up rpc client for app %s on chain %s", appName, chain)
	if err := cfg.Valid(); err != nil {
		return nil, err
	}
	m, err := ethclient.Dial(cfg.RpcUrl)
	if err != nil {
		return nil, err
	}
	b, err := ethclient.Dial(cfg.FailoverRpcUrl)
	if err != nil {
		return nil, err
	}
	c := client{
		logger: logger,
		cfg:    cfg,
		m:      m,
		b:      b,
	}
	if cfg.EnablePrometheus {
		logger.Info().Msgf("enabling rpc metrics")
		c.metrics = newMetrics(appName, chain)
		c.metrics.Register()
	}
	return &c, nil
}

func (c *client) shouldFailover(err error) bool {
	if err == context.DeadlineExceeded || err == context.Canceled {
		return false
	}
	return true
}

func (c *client) BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error) {
	t := time.Now()
	r, err := c.m.BalanceAt(ctx, account, blockNumber)
	c.metrics.Observe("BalanceAt", t, c.cfg.RpcName, err == nil)

	if err != nil {
		if !c.shouldFailover(err) {
			return r, err
		}

		// use failover rpc client
		t = time.Now()
		r, err = c.b.BalanceAt(ctx, account, blockNumber)
		c.metrics.Observe("BalanceAt", t, c.cfg.FailoverRpcName, err == nil)
		return r, err
	}
	return r, nil
}

func (c *client) BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error) {
	t := time.Now()
	r, err := c.m.BlockByHash(ctx, hash)
	c.metrics.Observe("BlockByHash", t, c.cfg.RpcName, err == nil)

	if err != nil {
		if !c.shouldFailover(err) {
			return r, err
		}

		// use failover rpc client
		t = time.Now()
		r, err = c.b.BlockByHash(ctx, hash)
		c.metrics.Observe("BlockByHash", t, c.cfg.FailoverRpcName, err == nil)
		return r, err
	}
	return r, nil
}

func (c *client) BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
	t := time.Now()
	r, err := c.m.BlockByNumber(ctx, number)
	c.metrics.Observe("BlockByNumber", t, c.cfg.RpcName, err == nil)

	if err != nil {
		if !c.shouldFailover(err) {
			return r, err
		}

		// use failover rpc client
		t = time.Now()
		r, err = c.b.BlockByNumber(ctx, number)
		c.metrics.Observe("BlockByNumber", t, c.cfg.FailoverRpcName, err == nil)
		return r, err
	}
	return r, nil
}

func (c *client) BlockNumber(ctx context.Context) (uint64, error) {
	t := time.Now()
	r, err := c.m.BlockNumber(ctx)
	c.metrics.Observe("BlockNumber", t, c.cfg.RpcName, err == nil)

	if err != nil {
		if !c.shouldFailover(err) {
			return r, err
		}

		// use failover rpc client
		t = time.Now()
		r, err = c.b.BlockNumber(ctx)
		c.metrics.Observe("BlockNumber", t, c.cfg.FailoverRpcName, err == nil)
		return r, err
	}
	return r, nil
}

func (c *client) CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	t := time.Now()
	r, err := c.m.CallContract(ctx, msg, blockNumber)
	c.metrics.Observe("CallContract", t, c.cfg.RpcName, err == nil)

	if err != nil {
		if !c.shouldFailover(err) {
			return r, err
		}

		// use failover rpc client
		t = time.Now()
		r, err = c.b.CallContract(ctx, msg, blockNumber)
		c.metrics.Observe("CallContract", t, c.cfg.FailoverRpcName, err == nil)
		return r, err
	}
	return r, nil
}

func (c *client) CallContractAtHash(ctx context.Context, msg ethereum.CallMsg, blockHash common.Hash) ([]byte, error) {
	t := time.Now()
	r, err := c.m.CallContractAtHash(ctx, msg, blockHash)
	c.metrics.Observe("CallContractAtHash", t, c.cfg.RpcName, err == nil)

	if err != nil {
		if !c.shouldFailover(err) {
			return r, err
		}

		// use failover rpc client
		t = time.Now()
		r, err = c.b.CallContractAtHash(ctx, msg, blockHash)
		c.metrics.Observe("CallContractAtHash", t, c.cfg.FailoverRpcName, err == nil)
		return r, err
	}
	return r, nil
}

func (c *client) ChainID(ctx context.Context) (*big.Int, error) {
	t := time.Now()
	r, err := c.m.ChainID(ctx)
	c.metrics.Observe("ChainID", t, c.cfg.RpcName, err == nil)

	if err != nil {
		if !c.shouldFailover(err) {
			return r, err
		}

		// use failover rpc client
		t = time.Now()
		r, err = c.b.ChainID(ctx)
		c.metrics.Observe("ChainID", t, c.cfg.FailoverRpcName, err == nil)
		return r, err
	}
	return r, nil
}

func (c *client) Close() {
	c.m.Close()
	c.b.Close()
}

func (c *client) CodeAt(ctx context.Context, account common.Address, blockNumber *big.Int) ([]byte, error) {
	t := time.Now()
	r, err := c.m.CodeAt(ctx, account, blockNumber)
	c.metrics.Observe("CodeAt", t, c.cfg.RpcName, err == nil)

	if err != nil {
		if !c.shouldFailover(err) {
			return r, err
		}

		// use failover rpc client
		t = time.Now()
		r, err = c.b.CodeAt(ctx, account, blockNumber)
		c.metrics.Observe("CodeAt", t, c.cfg.FailoverRpcName, err == nil)
		return r, err
	}
	return r, nil
}

func (c *client) EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error) {
	t := time.Now()
	r, err := c.m.EstimateGas(ctx, msg)
	c.metrics.Observe("EstimateGas", t, c.cfg.RpcName, err == nil)

	if err != nil {
		if !c.shouldFailover(err) {
			return r, err
		}

		// use failover rpc client
		t = time.Now()
		r, err = c.b.EstimateGas(ctx, msg)
		c.metrics.Observe("EstimateGas", t, c.cfg.FailoverRpcName, err == nil)
		return r, err
	}
	return r, nil
}

func (c *client) FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error) {
	t := time.Now()
	r, err := c.m.FilterLogs(ctx, q)
	c.metrics.Observe("FilterLogs", t, c.cfg.RpcName, err == nil)

	if err != nil {
		if !c.shouldFailover(err) {
			return r, err
		}

		// use failover rpc client
		t = time.Now()
		r, err = c.b.FilterLogs(ctx, q)
		c.metrics.Observe("FilterLogs", t, c.cfg.FailoverRpcName, err == nil)
		return r, err
	}
	return r, nil
}

func (c *client) HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error) {
	t := time.Now()
	r, err := c.m.HeaderByHash(ctx, hash)
	c.metrics.Observe("HeaderByHash", t, c.cfg.RpcName, err == nil)

	if err != nil {
		if !c.shouldFailover(err) {
			return r, err
		}

		// use failover rpc client
		t = time.Now()
		r, err = c.b.HeaderByHash(ctx, hash)
		c.metrics.Observe("HeaderByHash", t, c.cfg.FailoverRpcName, err == nil)
		return r, err
	}
	return r, nil
}

func (c *client) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	t := time.Now()
	r, err := c.m.HeaderByNumber(ctx, number)
	c.metrics.Observe("HeaderByNumber", t, c.cfg.RpcName, err == nil)

	if err != nil {
		if !c.shouldFailover(err) {
			return r, err
		}

		// use failover rpc client
		t = time.Now()
		r, err = c.b.HeaderByNumber(ctx, number)
		c.metrics.Observe("HeaderByNumber", t, c.cfg.FailoverRpcName, err == nil)
		return r, err
	}
	return r, nil
}

func (c *client) NetworkID(ctx context.Context) (*big.Int, error) {
	t := time.Now()
	r, err := c.m.NetworkID(ctx)
	c.metrics.Observe("NetworkID", t, c.cfg.RpcName, err == nil)

	if err != nil {
		if !c.shouldFailover(err) {
			return r, err
		}

		// use failover rpc client
		t = time.Now()
		r, err = c.b.NetworkID(ctx)
		c.metrics.Observe("NetworkID", t, c.cfg.FailoverRpcName, err == nil)
		return r, err
	}
	return r, nil
}

func (c *client) NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error) {
	t := time.Now()
	r, err := c.m.NonceAt(ctx, account, blockNumber)
	c.metrics.Observe("NonceAt", t, c.cfg.RpcName, err == nil)

	if err != nil {
		if !c.shouldFailover(err) {
			return r, err
		}

		// use failover rpc client
		t = time.Now()
		r, err = c.b.NonceAt(ctx, account, blockNumber)
		c.metrics.Observe("NonceAt", t, c.cfg.FailoverRpcName, err == nil)
		return r, err
	}
	return r, nil
}

func (c *client) PeerCount(ctx context.Context) (uint64, error) {
	t := time.Now()
	r, err := c.m.PeerCount(ctx)
	c.metrics.Observe("PeerCount", t, c.cfg.RpcName, err == nil)

	if err != nil {
		if !c.shouldFailover(err) {
			return r, err
		}

		// use failover rpc client
		t = time.Now()
		r, err = c.b.PeerCount(ctx)
		c.metrics.Observe("PeerCount", t, c.cfg.FailoverRpcName, err == nil)
		return r, err
	}
	return r, nil
}

func (c *client) PendingBalanceAt(ctx context.Context, account common.Address) (*big.Int, error) {
	t := time.Now()
	r, err := c.m.PendingBalanceAt(ctx, account)
	c.metrics.Observe("PendingBalanceAt", t, c.cfg.RpcName, err == nil)

	if err != nil {
		if !c.shouldFailover(err) {
			return r, err
		}

		// use failover rpc client
		t = time.Now()
		r, err = c.b.PendingBalanceAt(ctx, account)
		c.metrics.Observe("PendingBalanceAt", t, c.cfg.FailoverRpcName, err == nil)
		return r, err
	}
	return r, nil
}

func (c *client) PendingCallContract(ctx context.Context, msg ethereum.CallMsg) ([]byte, error) {
	t := time.Now()
	r, err := c.m.PendingCallContract(ctx, msg)
	c.metrics.Observe("PendingCallContract", t, c.cfg.RpcName, err == nil)

	if err != nil {
		if !c.shouldFailover(err) {
			return r, err
		}

		// use failover rpc client
		t = time.Now()
		r, err = c.b.PendingCallContract(ctx, msg)
		c.metrics.Observe("PendingCallContract", t, c.cfg.FailoverRpcName, err == nil)
		return r, err
	}
	return r, nil
}

func (c *client) PendingCodeAt(ctx context.Context, account common.Address) ([]byte, error) {
	t := time.Now()
	r, err := c.m.PendingCodeAt(ctx, account)
	c.metrics.Observe("PendingCodeAt", t, c.cfg.RpcName, err == nil)

	if err != nil {
		if !c.shouldFailover(err) {
			return r, err
		}

		// use failover rpc client
		t = time.Now()
		r, err = c.b.PendingCodeAt(ctx, account)
		c.metrics.Observe("PendingCodeAt", t, c.cfg.FailoverRpcName, err == nil)
		return r, err
	}
	return r, nil
}

func (c *client) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	t := time.Now()
	r, err := c.m.PendingNonceAt(ctx, account)
	c.metrics.Observe("PendingNonceAt", t, c.cfg.RpcName, err == nil)

	if err != nil {
		if !c.shouldFailover(err) {
			return r, err
		}

		// use failover rpc client
		t = time.Now()
		r, err = c.b.PendingNonceAt(ctx, account)
		c.metrics.Observe("PendingNonceAt", t, c.cfg.FailoverRpcName, err == nil)
		return r, err
	}
	return r, nil
}

func (c *client) PendingStorageAt(ctx context.Context, account common.Address, key common.Hash) ([]byte, error) {
	t := time.Now()
	r, err := c.m.PendingStorageAt(ctx, account, key)
	c.metrics.Observe("PendingStorageAt", t, c.cfg.RpcName, err == nil)

	if err != nil {
		if !c.shouldFailover(err) {
			return r, err
		}

		// use failover rpc client
		t = time.Now()
		r, err = c.b.PendingStorageAt(ctx, account, key)
		c.metrics.Observe("PendingStorageAt", t, c.cfg.FailoverRpcName, err == nil)
		return r, err
	}
	return r, nil
}

func (c *client) PendingTransactionCount(ctx context.Context) (uint, error) {
	t := time.Now()
	r, err := c.m.PendingTransactionCount(ctx)
	c.metrics.Observe("PendingTransactionCount", t, c.cfg.RpcName, err == nil)

	if err != nil {
		if !c.shouldFailover(err) {
			return r, err
		}

		// use failover rpc client
		t = time.Now()
		r, err = c.b.PendingTransactionCount(ctx)
		c.metrics.Observe("PendingTransactionCount", t, c.cfg.FailoverRpcName, err == nil)
		return r, err
	}
	return r, nil
}

func (c *client) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	t := time.Now()
	err := c.m.SendTransaction(ctx, tx)
	c.metrics.Observe("SendTransaction", t, c.cfg.RpcName, err == nil)

	if err != nil {
		if !c.shouldFailover(err) {
			return err
		}

		// use failover rpc client
		t = time.Now()
		err = c.b.SendTransaction(ctx, tx)
		c.metrics.Observe("SendTransaction", t, c.cfg.FailoverRpcName, err == nil)
		return err
	}
	return nil
}

func (c *client) StorageAt(ctx context.Context, account common.Address, key common.Hash, blockNumber *big.Int) ([]byte, error) {
	t := time.Now()
	r, err := c.m.StorageAt(ctx, account, key, blockNumber)
	c.metrics.Observe("StorageAt", t, c.cfg.RpcName, err == nil)

	if err != nil {
		if !c.shouldFailover(err) {
			return r, err
		}

		// use failover rpc client
		t = time.Now()
		r, err = c.b.StorageAt(ctx, account, key, blockNumber)
		c.metrics.Observe("StorageAt", t, c.cfg.FailoverRpcName, err == nil)
		return r, err
	}
	return r, nil
}

func (c *client) SubscribeFilterLogs(ctx context.Context, q ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error) {
	t := time.Now()
	r, err := c.m.SubscribeFilterLogs(ctx, q, ch)
	c.metrics.Observe("SubscribeFilterLogs", t, c.cfg.RpcName, err == nil)

	if err != nil {
		if !c.shouldFailover(err) {
			return r, err
		}

		// use failover rpc client
		t = time.Now()
		r, err = c.b.SubscribeFilterLogs(ctx, q, ch)
		c.metrics.Observe("SubscribeFilterLogs", t, c.cfg.FailoverRpcName, err == nil)
		return r, err
	}
	return r, nil
}

func (c *client) SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) (ethereum.Subscription, error) {
	t := time.Now()
	r, err := c.m.SubscribeNewHead(ctx, ch)
	c.metrics.Observe("SubscribeNewHead", t, c.cfg.RpcName, err == nil)

	if err != nil {
		if !c.shouldFailover(err) {
			return r, err
		}

		// use failover rpc client
		t = time.Now()
		r, err = c.b.SubscribeNewHead(ctx, ch)
		c.metrics.Observe("SubscribeNewHead", t, c.cfg.FailoverRpcName, err == nil)
		return r, err
	}
	return r, nil
}

func (c *client) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	t := time.Now()
	r, err := c.m.SuggestGasPrice(ctx)
	c.metrics.Observe("SuggestGasPrice", t, c.cfg.RpcName, err == nil)

	if err != nil {
		if !c.shouldFailover(err) {
			return r, err
		}

		// use failover rpc client
		t = time.Now()
		r, err = c.b.SuggestGasPrice(ctx)
		c.metrics.Observe("SuggestGasPrice", t, c.cfg.FailoverRpcName, err == nil)
		return r, err
	}
	return r, nil
}

func (c *client) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
	t := time.Now()
	r, err := c.m.SuggestGasTipCap(ctx)
	c.metrics.Observe("SuggestGasTipCap", t, c.cfg.RpcName, err == nil)

	if err != nil {
		if !c.shouldFailover(err) {
			return r, err
		}

		// use failover rpc client
		t = time.Now()
		r, err = c.b.SuggestGasTipCap(ctx)
		c.metrics.Observe("SuggestGasTipCap", t, c.cfg.FailoverRpcName, err == nil)
		return r, err
	}
	return r, nil
}

func (c *client) SyncProgress(ctx context.Context) (*ethereum.SyncProgress, error) {
	t := time.Now()
	r, err := c.m.SyncProgress(ctx)
	c.metrics.Observe("SyncProgress", t, c.cfg.RpcName, err == nil)

	if err != nil {
		if !c.shouldFailover(err) {
			return r, err
		}

		// use failover rpc client
		t = time.Now()
		r, err = c.b.SyncProgress(ctx)
		c.metrics.Observe("SyncProgress", t, c.cfg.FailoverRpcName, err == nil)
		return r, err
	}
	return r, nil
}

func (c *client) TransactionByHash(ctx context.Context, hash common.Hash) (tx *types.Transaction, isPending bool, err error) {
	t := time.Now()
	r1, r2, err := c.m.TransactionByHash(ctx, hash)
	c.metrics.Observe("BalanceAtTransactionByHash", t, c.cfg.RpcName, err == nil)

	if err != nil {
		if !c.shouldFailover(err) {
			return r1, r2, err
		}

		// use failover rpc client
		t = time.Now()
		r1, r2, err = c.b.TransactionByHash(ctx, hash)
		c.metrics.Observe("BalanceAtTransactionByHash", t, c.cfg.FailoverRpcName, err == nil)
		return r1, r2, err
	}
	return r1, r2, nil
}

func (c *client) TransactionCount(ctx context.Context, blockHash common.Hash) (uint, error) {
	t := time.Now()
	r, err := c.m.TransactionCount(ctx, blockHash)
	c.metrics.Observe("TransactionCount", t, c.cfg.RpcName, err == nil)

	if err != nil {
		if !c.shouldFailover(err) {
			return r, err
		}

		// use failover rpc client
		t = time.Now()
		r, err = c.b.TransactionCount(ctx, blockHash)
		c.metrics.Observe("TransactionCount", t, c.cfg.FailoverRpcName, err == nil)
		return r, err
	}
	return r, nil
}

func (c *client) TransactionInBlock(ctx context.Context, blockHash common.Hash, index uint) (*types.Transaction, error) {
	t := time.Now()
	r, err := c.m.TransactionInBlock(ctx, blockHash, index)
	c.metrics.Observe("TransactionInBlock", t, c.cfg.RpcName, err == nil)

	if err != nil {
		if !c.shouldFailover(err) {
			return r, err
		}

		// use failover rpc client
		t = time.Now()
		r, err = c.b.TransactionInBlock(ctx, blockHash, index)
		c.metrics.Observe("TransactionInBlock", t, c.cfg.FailoverRpcName, err == nil)
		return r, err
	}
	return r, nil
}

func (c *client) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	t := time.Now()
	r, err := c.m.TransactionReceipt(ctx, txHash)
	c.metrics.Observe("TransactionReceipt", t, c.cfg.RpcName, err == nil)

	if err != nil {
		if !c.shouldFailover(err) {
			return r, err
		}

		// use failover rpc client
		t = time.Now()
		r, err = c.b.TransactionReceipt(ctx, txHash)
		c.metrics.Observe("TransactionReceipt", t, c.cfg.FailoverRpcName, err == nil)
		return r, err
	}
	return r, nil
}

func (c *client) TransactionSender(ctx context.Context, tx *types.Transaction, block common.Hash, index uint) (common.Address, error) {
	t := time.Now()
	r, err := c.m.TransactionSender(ctx, tx, block, index)
	c.metrics.Observe("TransactionSender", t, c.cfg.RpcName, err == nil)

	if err != nil {
		if !c.shouldFailover(err) {
			return r, err
		}

		// use failover rpc client
		t = time.Now()
		r, err = c.b.TransactionSender(ctx, tx, block, index)
		c.metrics.Observe("TransactionSender", t, c.cfg.FailoverRpcName, err == nil)
		return r, err
	}
	return r, nil
}
