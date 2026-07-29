type ExchangeAccountMode = 'real' | 'demo' | 'testnet'

export interface ExchangeBinding {
  id: string
  exchange: string
  bindingId?: string
  name: string
  logo: string
  logoDark?: string
  status: 'connected' | 'disconnected'
  accountMode: ExchangeAccountMode
  lastSynced: string | null
  balance: number
  hasApi: boolean
}

export interface ExchangeCredentialSummary {
  id: string
  exchange: string
  accountMode: ExchangeAccountMode
  status: string
  maskedApiKey: string
  hasApiSecret: boolean
  hasPassphrase: boolean
  permissionScope: string
  lastVerifiedAt?: string
  createdAt: string
  updatedAt: string
}

export interface ActiveLayer {
  id: string
  pair: string
  layerNumber: number
  exchangeName: string
  exchangeDisplayName: string
  layerLabel: string
  entryPrice: number
  currentPrice: number
  allocationPct: number
  allocatedUsdt: number
  unrealizedPnl: number
  unrealizedPnlPct: number
  status: string
  openedAt: string
}

interface ApiExchangeBinding {
  id: string
  name: string
  accountMode?: string
  status: string
  lastSynced: string | null
  balance: string | number
  hasApi: boolean
}

interface ApiActiveLayer {
  id: string
  pair: string
  layerNumber?: string | number
  exchangeName?: string
  exchangeDisplayName?: string
  layerLabel?: string
  entryPrice: string | number
  currentPrice: string | number
  allocationPct: string | number
  allocatedUsdt: string | number
  unrealizedPnl: string | number
  unrealizedPnlPct: string | number
  status: string
  openedAt: string
}

interface BindExchangePayload {
  exchange: string
  apiKey: string
  apiSecret: string
  passphrase?: string
}

interface GetExchangeBindingsOptions {
  includeUnbound?: boolean
}

export const useDashboardData = () => {
  const config = useRuntimeConfig()
  const apiBase = config.public.apiBase
  const tokenCookie = useCookie<string | null>('auth_token')
  const exchangeCatalog: ExchangeBinding[] = [
    { id: 'binance', exchange: 'binance', name: 'Binance', logo: '/UserDashboard/Binance_logo.svg', status: 'disconnected', accountMode: 'real', lastSynced: null, balance: 0, hasApi: false },
    { id: 'okx', exchange: 'okx', name: 'OKX', logo: '/UserDashboard/OKX_logo.svg', logoDark: '/UserDashboard/OKX_logo_dark.svg', status: 'disconnected', accountMode: 'real', lastSynced: null, balance: 0, hasApi: false },
    { id: 'bybit', exchange: 'bybit', name: 'Bybit', logo: '/UserDashboard/Bybit_logo.svg', logoDark: '/UserDashboard/Bybit_logo_dark.svg', status: 'disconnected', accountMode: 'real', lastSynced: null, balance: 0, hasApi: false },
    { id: 'tokocrypto', exchange: 'tokocrypto', name: 'Tokocrypto', logo: '/UserDashboard/Tokocrypto_logo.svg', status: 'disconnected', accountMode: 'real', lastSynced: null, balance: 0, hasApi: false }
  ]

  const authHeaders = () => ({
    Authorization: `Bearer ${tokenCookie.value}`
  })

  const numberValue = (value: string | number | null | undefined) => {
    const parsed = Number(value ?? 0)
    return Number.isFinite(parsed) ? parsed : 0
  }

  const normalizeExchangeKey = (value: string) => {
    return value.toLowerCase().replace(/\s+/g, '')
  }

  const toExchangeStatus = (status: string): 'connected' | 'disconnected' => {
    return status.toLowerCase() === 'connected' || status.toLowerCase() === 'active' ? 'connected' : 'disconnected'
  }

  const toAccountMode = (mode: string | null | undefined): ExchangeAccountMode => {
    const normalizedMode = mode?.toLowerCase()
    if (normalizedMode === 'demo' || normalizedMode === 'testnet') return normalizedMode
    return 'real'
  }

  const formatApiError = (error: unknown, fallback: string) => {
    if (typeof error === 'object' && error !== null && 'data' in error) {
      const data = (error as { data?: { error?: string, message?: string } }).data
      return data?.error || data?.message || fallback
    }
    if (error instanceof Error) {
      return error.message
    }
    return fallback
  }

  const getUserStats = async () => {
    try {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const stats = await $fetch<any>(`${apiBase}/user/stats`, {
        headers: {
          Authorization: `Bearer ${tokenCookie.value}`
        }
      })
      // Convert string numbers to numbers for the frontend
      return {
        totalBalance: parseFloat(stats.totalBalance || '0'),
        realizedProfit: parseFloat(stats.realizedProfit || '0'),
        totalGasFeePaid: parseFloat(stats.totalGasFeePaid || '0'),
        activeLayersCount: stats.activeLayersCount || 0,
        gasFeeDepositStatus: stats.gasFeeDepositStatus || 'none',
        gasFeeDepositTxId: stats.gasFeeDepositTxId || ''
      }
    } catch (error) {
      console.warn('Failed to fetch user stats, falling back to mock:', error)
      return {
        totalBalance: 12450.75,
        realizedProfit: 3240.50,
        totalGasFeePaid: 1620.25,
        activeLayersCount: 18,
        gasFeeDepositStatus: 'none',
        gasFeeDepositTxId: ''
      }
    }
  }

  const getGasFeeAccount = async () => {
    return await $fetch(`${apiBase}/user/gas-fee`, {
      headers: authHeaders()
    })
  }

  const createGasFeeDeposit = async (payload: { amount: number | string, asset?: string, txId: string }) => {
    return await $fetch(`${apiBase}/user/gas-fee/deposits`, {
      method: 'POST',
      headers: authHeaders(),
      body: {
        amount: String(payload.amount),
        asset: payload.asset || 'USDT',
        tx_id: payload.txId
      }
    })
  }

  const getExchangeBindings = async (options: GetExchangeBindingsOptions = {}): Promise<ExchangeBinding[]> => {
    try {
      const bindings = await $fetch<ApiExchangeBinding[]>(`${apiBase}/user/exchange-bindings`, {
        headers: authHeaders()
      })
      const bindingsByExchange = new Map(
        bindings.map(binding => [normalizeExchangeKey(binding.name), binding])
      )

      const mergedBindings = exchangeCatalog.map((catalogExchange) => {
        const binding = bindingsByExchange.get(catalogExchange.exchange)
        if (!binding) return { ...catalogExchange }

        return {
          ...catalogExchange,
          bindingId: binding.id,
          status: toExchangeStatus(binding.status),
          accountMode: toAccountMode(binding.accountMode),
          lastSynced: binding.lastSynced,
          balance: numberValue(binding.balance),
          hasApi: binding.hasApi
        }
      })
      return options.includeUnbound ? mergedBindings : mergedBindings.filter(exchange => exchange.hasApi)
    } catch (error) {
      console.warn('Failed to fetch exchange bindings:', error)
      return options.includeUnbound ? exchangeCatalog.map(exchange => ({ ...exchange })) : []
    }
  }

  const bindExchange = async (payload: BindExchangePayload): Promise<ExchangeCredentialSummary> => {
    const body: Record<string, string> = {
      exchange: payload.exchange,
      api_key: payload.apiKey,
      api_secret: payload.apiSecret,
      permission_scope: 'trade_only'
    }

    if (payload.passphrase?.trim()) {
      body.api_passphrase = payload.passphrase.trim()
    }

    try {
      return await $fetch<ExchangeCredentialSummary>(`${apiBase}/user/exchange-bindings`, {
        method: 'POST',
        headers: authHeaders(),
        body
      })
    } catch (error) {
      throw new Error(formatApiError(error, 'Failed to bind exchange'), { cause: error })
    }
  }

  const getExchangeBindingCredentials = async (exchange: string): Promise<ExchangeCredentialSummary> => {
    try {
      return await $fetch<ExchangeCredentialSummary>(`${apiBase}/user/exchange-bindings/${encodeURIComponent(exchange)}/credentials`, {
        headers: authHeaders()
      })
    } catch (error) {
      throw new Error(formatApiError(error, 'Failed to read exchange credential'), { cause: error })
    }
  }

  const updateExchangeBindingStatus = async (exchange: string, status: 'connected' | 'disconnected'): Promise<ExchangeCredentialSummary> => {
    try {
      return await $fetch<ExchangeCredentialSummary>(`${apiBase}/user/exchange-bindings/${encodeURIComponent(exchange)}/status`, {
        method: 'PATCH',
        headers: authHeaders(),
        body: { status }
      })
    } catch (error) {
      throw new Error(formatApiError(error, 'Failed to update exchange status'), { cause: error })
    }
  }

  const deleteExchangeBinding = async (exchange: string): Promise<ExchangeCredentialSummary> => {
    try {
      return await $fetch<ExchangeCredentialSummary>(`${apiBase}/user/exchange-bindings/${encodeURIComponent(exchange)}`, {
        method: 'DELETE',
        headers: authHeaders()
      })
    } catch (error) {
      throw new Error(formatApiError(error, 'Failed to delete exchange API'), { cause: error })
    }
  }

  const getActiveLayers = async (): Promise<ActiveLayer[]> => {
    try {
      const layers = await $fetch<ApiActiveLayer[]>(`${apiBase}/user/layers`, {
        headers: authHeaders()
      })

      return layers.map((layer) => {
        const layerNumber = numberValue(layer.layerNumber)
        const exchangeDisplayName = layer.exchangeDisplayName?.trim() || layer.exchangeName?.trim() || 'Exchange'
        const layerLabel = layer.layerLabel?.trim() || (layerNumber > 0 ? `L${layerNumber} ${exchangeDisplayName}` : exchangeDisplayName)

        return {
          id: layer.id,
          pair: layer.pair,
          layerNumber,
          exchangeName: layer.exchangeName || '',
          exchangeDisplayName,
          layerLabel,
          entryPrice: numberValue(layer.entryPrice),
          currentPrice: numberValue(layer.currentPrice),
          allocationPct: numberValue(layer.allocationPct),
          allocatedUsdt: numberValue(layer.allocatedUsdt),
          unrealizedPnl: numberValue(layer.unrealizedPnl),
          unrealizedPnlPct: numberValue(layer.unrealizedPnlPct),
          status: layer.status,
          openedAt: layer.openedAt
        }
      })
    } catch (error) {
      console.warn('Failed to fetch active layers:', error)
      return []
    }
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
      },
      {
        id: 'L-097',
        pair: 'BTC/USDT',
        exitPrice: 62120.50,
        pnl: 88.75,
        gasFee: 42.10,
        closedAt: '2026-07-14T16:45:00Z'
      },
      {
        id: 'L-096',
        pair: 'ETH/USDT',
        exitPrice: 3525.40,
        pnl: 64.20,
        gasFee: 30.35,
        closedAt: '2026-07-14T12:15:00Z'
      },
      {
        id: 'L-095',
        pair: 'SOL/USDT',
        exitPrice: 148.80,
        pnl: -12.45,
        gasFee: 18.80,
        closedAt: '2026-07-13T18:20:00Z'
      },
      {
        id: 'L-094',
        pair: 'XRP/USDT',
        exitPrice: 0.61,
        pnl: 31.10,
        gasFee: 9.45,
        closedAt: '2026-07-13T10:05:00Z'
      },
      {
        id: 'L-093',
        pair: 'DOGE/USDT',
        exitPrice: 0.13,
        pnl: 17.65,
        gasFee: 7.20,
        closedAt: '2026-07-12T21:40:00Z'
      },
      {
        id: 'L-092',
        pair: 'AVAX/USDT',
        exitPrice: 32.05,
        pnl: -28.90,
        gasFee: 16.75,
        closedAt: '2026-07-12T14:25:00Z'
      },
      {
        id: 'L-091',
        pair: 'LINK/USDT',
        exitPrice: 15.05,
        pnl: 39.80,
        gasFee: 12.35,
        closedAt: '2026-07-11T19:10:00Z'
      },
      {
        id: 'L-090',
        pair: 'DOT/USDT',
        exitPrice: 6.28,
        pnl: 11.25,
        gasFee: 8.90,
        closedAt: '2026-07-11T08:35:00Z'
      },
      {
        id: 'L-089',
        pair: 'LTC/USDT',
        exitPrice: 84.95,
        pnl: -15.70,
        gasFee: 14.45,
        closedAt: '2026-07-10T17:55:00Z'
      },
      {
        id: 'L-088',
        pair: 'ATOM/USDT',
        exitPrice: 7.34,
        pnl: 22.40,
        gasFee: 6.30,
        closedAt: '2026-07-10T11:20:00Z'
      },
      {
        id: 'L-087',
        pair: 'UNI/USDT',
        exitPrice: 9.62,
        pnl: -9.35,
        gasFee: 5.85,
        closedAt: '2026-07-09T20:15:00Z'
      },
      {
        id: 'L-086',
        pair: 'NEAR/USDT',
        exitPrice: 5.72,
        pnl: 36.65,
        gasFee: 10.20,
        closedAt: '2026-07-09T13:45:00Z'
      },
      {
        id: 'L-085',
        pair: 'APT/USDT',
        exitPrice: 8.11,
        pnl: -18.10,
        gasFee: 11.75,
        closedAt: '2026-07-08T18:05:00Z'
      },
      {
        id: 'L-084',
        pair: 'ARB/USDT',
        exitPrice: 1.31,
        pnl: 14.95,
        gasFee: 4.60,
        closedAt: '2026-07-08T09:30:00Z'
      },
      {
        id: 'L-083',
        pair: 'OP/USDT',
        exitPrice: 2.18,
        pnl: 19.55,
        gasFee: 5.15,
        closedAt: '2026-07-07T22:50:00Z'
      },
      {
        id: 'L-082',
        pair: 'MATIC/USDT',
        exitPrice: 0.75,
        pnl: 12.70,
        gasFee: 3.90,
        closedAt: '2026-07-07T15:10:00Z'
      }
    ]
  }

  return {
    getUserStats,
    getGasFeeAccount,
    createGasFeeDeposit,
    getExchangeBindings,
    bindExchange,
    getExchangeBindingCredentials,
    updateExchangeBindingStatus,
    deleteExchangeBinding,
    getActiveLayers,
    getHistory
  }
}
