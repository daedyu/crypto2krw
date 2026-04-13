import Foundation

// MARK: - Mock Data

enum MockData {
    static let rates: [Rate] = [
        Rate(currency: .USDT, krwRate: 1385),
        Rate(currency: .SOL,  krwRate: 192000),
        Rate(currency: .ETH,  krwRate: 4820000),
    ]

    static let balances: [CoinBalance] = [
        CoinBalance(currency: .USDT, amount: 320.5,
                    krwValue: 320.5 * 1385,
                    address: "TQn9Y2khEsLJW1ChVWFMSMeRDow5KcbLSE"),
        CoinBalance(currency: .SOL,  amount: 2.41,
                    krwValue: 2.41 * 192000,
                    address: "7xKXtg2CW87d97TXJSDpbD5jBkheTqA83TZRuJosgAsU"),
        CoinBalance(currency: .ETH,  amount: 0.12,
                    krwValue: 0.12 * 4820000,
                    address: "0x742d35Cc6634C0532925a3b844Bc454e4438f44e"),
    ]

    static let transactions: [Transaction] = [
        Transaction(id: "tx_001", type: .payment,
                    merchantName: "스타벅스 강남점",
                    amountKrw: 8500, usedCurrency: .USDT, usedAmount: 6.14,
                    appliedRate: 1385, status: .success,
                    createdAt: iso("2026-04-06T14:23:00Z")),
        Transaction(id: "tx_002", type: .deposit,
                    merchantName: nil,
                    amountKrw: 443200, usedCurrency: .USDT, usedAmount: 320,
                    appliedRate: 1385, status: .success,
                    createdAt: iso("2026-04-05T09:10:00Z")),
        Transaction(id: "tx_003", type: .payment,
                    merchantName: "올리브영 홍대점",
                    amountKrw: 32000, usedCurrency: .USDT, usedAmount: 23.1,
                    appliedRate: 1385, status: .success,
                    createdAt: iso("2026-04-04T18:45:00Z")),
        Transaction(id: "tx_004", type: .payment,
                    merchantName: "파리바게뜨",
                    amountKrw: 12500, usedCurrency: .SOL, usedAmount: 0.065,
                    appliedRate: 192000, status: .success,
                    createdAt: iso("2026-04-03T11:30:00Z")),
        Transaction(id: "tx_005", type: .deposit,
                    merchantName: nil,
                    amountKrw: 576000, usedCurrency: .SOL, usedAmount: 3.0,
                    appliedRate: 192000, status: .success,
                    createdAt: iso("2026-04-01T08:00:00Z")),
    ]

    static let depositInfo: [CoinDepositInfo] = [
        CoinDepositInfo(currency: .USDT, networks: [
            DepositNetwork(network: .trc20,
                           address: "TQn9Y2khEsLJW1ChVWFMSMeRDow5KcbLSE",
                           minDeposit: 10, minDepositUnit: "USDT",
                           confirmations: 20, arrivalTime: "약 5분",
                           networkDescription: "TRON 네트워크 (수수료 저렴)"),
            DepositNetwork(network: .erc20,
                           address: "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
                           minDeposit: 50, minDepositUnit: "USDT",
                           confirmations: 12, arrivalTime: "약 3분",
                           networkDescription: "Ethereum 네트워크"),
        ]),
        CoinDepositInfo(currency: .SOL, networks: [
            DepositNetwork(network: .sol,
                           address: "7xKXtg2CW87d97TXJSDpbD5jBkheTqA83TZRuJosgAsU",
                           minDeposit: 0.1, minDepositUnit: "SOL",
                           confirmations: 32, arrivalTime: "약 30초",
                           networkDescription: "Solana 네트워크"),
        ]),
        CoinDepositInfo(currency: .ETH, networks: [
            DepositNetwork(network: .erc20,
                           address: "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
                           minDeposit: 0.005, minDepositUnit: "ETH",
                           confirmations: 12, arrivalTime: "약 3분",
                           networkDescription: "Ethereum 네트워크"),
        ]),
    ]

    static let depositRecords: [DepositRecord] = [
        DepositRecord(id: "dep_001",
                      txHash: "a3f4b2c1d5e6f789...",
                      currency: .USDT, network: .trc20,
                      amount: 320, status: .confirmed,
                      confirmCount: 20, requiredConfirmCount: 20,
                      createdAt: iso("2026-04-05T09:10:00Z")),
        DepositRecord(id: "dep_002",
                      txHash: "b7c8d9e0f1a2b3c4...",
                      currency: .SOL, network: .sol,
                      amount: 3.0, status: .confirmed,
                      confirmCount: 32, requiredConfirmCount: 32,
                      createdAt: iso("2026-04-01T08:00:00Z")),
        DepositRecord(id: "dep_003",
                      txHash: "c1d2e3f4a5b6c7d8...",
                      currency: .USDT, network: .erc20,
                      amount: 100, status: .pending,
                      confirmCount: 7, requiredConfirmCount: 12,
                      createdAt: iso("2026-04-07T21:30:00Z")),
    ]

    private static func iso(_ string: String) -> Date {
        let formatter = ISO8601DateFormatter()
        return formatter.date(from: string) ?? Date()
    }
}
