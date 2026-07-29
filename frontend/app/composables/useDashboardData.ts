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

export interface TradeHistory {
  id: string
  pair: string
  layerNumber: number
  exchangeName: string
  exchangeDisplayName: string
  layerLabel: string
  exitPrice: number
  pnl: number
  gasFee: number
  closedAt: string
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

interface ApiTradeHistory {
  id: string
  pair: string
  layerNumber?: string | number
  exchangeName?: string
  exchangeDisplayName?: string
  layerLabel?: string
  exitPrice: string | number
  pnl: string | number
  gasFee: string | number
  closedAt: string
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

  const getHistory = async (): Promise<TradeHistory[]> => {
    try {
      const history = await $fetch<ApiTradeHistory[]>(`${apiBase}/user/history/trades`, {
        headers: authHeaders()
      })

      return history.map((item) => {
        const layerNumber = numberValue(item.layerNumber)
        const exchangeDisplayName = item.exchangeDisplayName?.trim() || item.exchangeName?.trim() || 'Exchange'
        const layerLabel = item.layerLabel?.trim() || (layerNumber > 0 ? `L${layerNumber} ${exchangeDisplayName}` : item.id)

        return {
          id: item.id,
          pair: item.pair,
          layerNumber,
          exchangeName: item.exchangeName || '',
          exchangeDisplayName,
          layerLabel,
          exitPrice: numberValue(item.exitPrice),
          pnl: numberValue(item.pnl),
          gasFee: numberValue(item.gasFee),
          closedAt: item.closedAt
        }
      })
    } catch (error) {
      console.warn('Failed to fetch trade history:', error)
      return []
    }
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
