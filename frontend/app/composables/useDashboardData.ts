export const useDashboardData = () => {
  // Mock data structure, ready to be replaced with real API calls using useFetch or $fetch

  const getUserStats = async () => {
    return {
      totalBalance: 12450.75,
      realizedProfit: 3240.50,
      totalGasFeePaid: 1620.25,
      activeLayersCount: 3
    }
  }

  const getExchangeBindings = async () => {
    return [
      { id: 1, name: 'Binance', logo: '/UserDashboard/Binance_logo.svg', status: 'connected', lastSynced: '2026-07-18T10:30:00Z', balance: 8450.75 },
      { id: 2, name: 'OKX', logo: '/UserDashboard/OKX_logo.svg', status: 'connected', lastSynced: '2026-07-18T10:30:00Z', balance: 4000.00 },
      { id: 3, name: 'Bybit', logo: '/UserDashboard/Bybit_logo.svg', status: 'disconnected', lastSynced: '2026-07-01T08:00:00Z', balance: 0.00 },
      { id: 4, name: 'Tokocrypto', logo: '/UserDashboard/Tokocrypto_logo.svg', status: 'disconnected', lastSynced: null, balance: 0.00 }
    ]
  }

  const getActiveLayers = async () => {
    return [
      {
        id: 'L-101',
        pair: 'BTC/USDT',
        entryPrice: 62450.00,
        currentPrice: 63100.50,
        allocationPct: 10,
        allocatedUsdt: 845.07,
        unrealizedPnl: 8.79, // USDT
        unrealizedPnlPct: 1.04,
        status: 'open',
        openedAt: '2026-07-18T08:15:00Z'
      },
      {
        id: 'L-102',
        pair: 'ETH/USDT',
        entryPrice: 3450.25,
        currentPrice: 3410.00,
        allocationPct: 5,
        allocatedUsdt: 422.53,
        unrealizedPnl: -4.93, // USDT
        unrealizedPnlPct: -1.16,
        status: 'open',
        openedAt: '2026-07-18T09:00:00Z'
      },
      {
        id: 'L-103',
        pair: 'SOL/USDT',
        entryPrice: 145.50,
        currentPrice: 151.20,
        allocationPct: 15,
        allocatedUsdt: 1267.61,
        unrealizedPnl: 49.65, // USDT
        unrealizedPnlPct: 3.91,
        status: 'open',
        openedAt: '2026-07-17T14:20:00Z'
      }
    ]
  }

  const getHistory = async () => {
    return [
      {
        id: 'L-099',
        pair: 'BNB/USDT',
        exitPrice: 580.00,
        pnl: 120.50,
        gasFee: 60.25,
        closedAt: '2026-07-16T11:00:00Z'
      },
      {
        id: 'L-098',
        pair: 'ADA/USDT',
        exitPrice: 0.42,
        pnl: -20.00,
        gasFee: -10.00, // rebate
        closedAt: '2026-07-15T09:30:00Z'
      }
    ]
  }

  return {
    getUserStats,
    getExchangeBindings,
    getActiveLayers,
    getHistory
  }
}
