export const useDashboardData = () => {
  // Mock data structure, ready to be replaced with real API calls using useFetch or $fetch

  const getUserStats = async () => {
    return {
      totalBalance: 12450.75,
      realizedProfit: 3240.50,
      totalGasFeePaid: 1620.25,
      activeLayersCount: 18
    }
  }

  const getExchangeBindings = async () => {
    return [
      { id: 1, name: 'Binance', logo: '/UserDashboard/Binance_logo.svg', status: 'connected', lastSynced: '2026-07-18T10:30:00Z', balance: 8450.75 },
      { id: 2, name: 'OKX', logo: '/UserDashboard/OKX_logo.svg', logoDark: '/UserDashboard/OKX_logo_dark.svg', status: 'connected', lastSynced: '2026-07-18T10:30:00Z', balance: 4000.00 },
      { id: 3, name: 'Bybit', logo: '/UserDashboard/Bybit_logo.svg', logoDark: '/UserDashboard/Bybit_logo_dark.svg', status: 'disconnected', lastSynced: '2026-07-01T08:00:00Z', balance: 0.00 },
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
      },
      {
        id: 'L-104',
        pair: 'BNB/USDT',
        entryPrice: 580.00,
        currentPrice: 592.35,
        allocationPct: 8,
        allocatedUsdt: 996.06,
        unrealizedPnl: 21.22, // USDT
        unrealizedPnlPct: 2.13,
        status: 'open',
        openedAt: '2026-07-17T18:45:00Z'
      },
      {
        id: 'L-105',
        pair: 'XRP/USDT',
        entryPrice: 0.58,
        currentPrice: 0.56,
        allocationPct: 4,
        allocatedUsdt: 498.03,
        unrealizedPnl: -17.17, // USDT
        unrealizedPnlPct: -3.45,
        status: 'open',
        openedAt: '2026-07-18T06:20:00Z'
      },
      {
        id: 'L-106',
        pair: 'DOGE/USDT',
        entryPrice: 0.12,
        currentPrice: 0.13,
        allocationPct: 6,
        allocatedUsdt: 747.05,
        unrealizedPnl: 62.25, // USDT
        unrealizedPnlPct: 8.33,
        status: 'open',
        openedAt: '2026-07-18T10:05:00Z'
      },
      {
        id: 'L-107',
        pair: 'ADA/USDT',
        entryPrice: 0.42,
        currentPrice: 0.44,
        allocationPct: 5,
        allocatedUsdt: 622.54,
        unrealizedPnl: 29.64, // USDT
        unrealizedPnlPct: 4.76,
        status: 'open',
        openedAt: '2026-07-18T10:40:00Z'
      },
      {
        id: 'L-108',
        pair: 'AVAX/USDT',
        entryPrice: 31.25,
        currentPrice: 30.70,
        allocationPct: 7,
        allocatedUsdt: 871.55,
        unrealizedPnl: -15.34, // USDT
        unrealizedPnlPct: -1.76,
        status: 'open',
        openedAt: '2026-07-18T11:15:00Z'
      },
      {
        id: 'L-109',
        pair: 'LINK/USDT',
        entryPrice: 14.80,
        currentPrice: 15.42,
        allocationPct: 9,
        allocatedUsdt: 1120.57,
        unrealizedPnl: 46.93, // USDT
        unrealizedPnlPct: 4.19,
        status: 'open',
        openedAt: '2026-07-18T11:50:00Z'
      },
      {
        id: 'L-110',
        pair: 'MATIC/USDT',
        entryPrice: 0.72,
        currentPrice: 0.71,
        allocationPct: 3,
        allocatedUsdt: 373.52,
        unrealizedPnl: -5.19, // USDT
        unrealizedPnlPct: -1.39,
        status: 'open',
        openedAt: '2026-07-18T12:25:00Z'
      },
      {
        id: 'L-111',
        pair: 'DOT/USDT',
        entryPrice: 6.15,
        currentPrice: 6.39,
        allocationPct: 6,
        allocatedUsdt: 747.05,
        unrealizedPnl: 29.14, // USDT
        unrealizedPnlPct: 3.90,
        status: 'open',
        openedAt: '2026-07-18T13:00:00Z'
      },
      {
        id: 'L-112',
        pair: 'LTC/USDT',
        entryPrice: 86.40,
        currentPrice: 88.10,
        allocationPct: 5,
        allocatedUsdt: 622.54,
        unrealizedPnl: 12.25, // USDT
        unrealizedPnlPct: 1.97,
        status: 'open',
        openedAt: '2026-07-18T13:35:00Z'
      },
      {
        id: 'L-113',
        pair: 'ATOM/USDT',
        entryPrice: 7.20,
        currentPrice: 7.06,
        allocationPct: 4,
        allocatedUsdt: 498.03,
        unrealizedPnl: -9.69, // USDT
        unrealizedPnlPct: -1.94,
        status: 'open',
        openedAt: '2026-07-18T14:10:00Z'
      },
      {
        id: 'L-114',
        pair: 'UNI/USDT',
        entryPrice: 9.85,
        currentPrice: 10.22,
        allocationPct: 6,
        allocatedUsdt: 747.05,
        unrealizedPnl: 28.07, // USDT
        unrealizedPnlPct: 3.76,
        status: 'open',
        openedAt: '2026-07-18T14:45:00Z'
      },
      {
        id: 'L-115',
        pair: 'NEAR/USDT',
        entryPrice: 5.45,
        currentPrice: 5.61,
        allocationPct: 7,
        allocatedUsdt: 871.55,
        unrealizedPnl: 25.58, // USDT
        unrealizedPnlPct: 2.94,
        status: 'open',
        openedAt: '2026-07-18T15:20:00Z'
      },
      {
        id: 'L-116',
        pair: 'APT/USDT',
        entryPrice: 8.30,
        currentPrice: 8.05,
        allocationPct: 5,
        allocatedUsdt: 622.54,
        unrealizedPnl: -18.75, // USDT
        unrealizedPnlPct: -3.01,
        status: 'open',
        openedAt: '2026-07-18T15:55:00Z'
      },
      {
        id: 'L-117',
        pair: 'ARB/USDT',
        entryPrice: 1.28,
        currentPrice: 1.34,
        allocationPct: 4,
        allocatedUsdt: 498.03,
        unrealizedPnl: 23.34, // USDT
        unrealizedPnlPct: 4.69,
        status: 'open',
        openedAt: '2026-07-18T16:30:00Z'
      },
      {
        id: 'L-118',
        pair: 'OP/USDT',
        entryPrice: 2.05,
        currentPrice: 2.11,
        allocationPct: 3,
        allocatedUsdt: 373.52,
        unrealizedPnl: 10.93, // USDT
        unrealizedPnlPct: 2.93,
        status: 'open',
        openedAt: '2026-07-18T17:05:00Z'
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
