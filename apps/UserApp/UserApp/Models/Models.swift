import Foundation

// MARK: - Enums

enum Currency: String, CaseIterable, Identifiable {
    case USDT, SOL, ETH
    var id: String { rawValue }

    var fullName: String {
        switch self {
        case .USDT: return "Tether"
        case .SOL:  return "Solana"
        case .ETH:  return "Ethereum"
        }
    }

    var symbol: String {
        switch self {
        case .USDT: return "$"
        case .SOL:  return "◎"
        case .ETH:  return "Ξ"
        }
    }

    var accentHex: String {
        switch self {
        case .USDT: return "#26A17B"
        case .SOL:  return "#9945FF"
        case .ETH:  return "#627EEA"
        }
    }

    var backgroundHex: String {
        switch self {
        case .USDT: return "#D4EDE5"
        case .SOL:  return "#E4D9FF"
        case .ETH:  return "#D4DEF5"
        }
    }
}

enum NetworkType: String {
    case trc20 = "TRC-20"
    case erc20 = "ERC-20"
    case sol   = "SOL"
}

enum TransactionType: String {
    case payment = "PAYMENT"
    case deposit = "DEPOSIT"
}

enum TransactionStatus: String {
    case success = "SUCCESS"
    case pending = "PENDING"
    case failed  = "FAILED"
}

enum DepositStatus: String {
    case confirmed = "CONFIRMED"
    case pending   = "PENDING"
}

// MARK: - Models

struct CoinBalance: Identifiable {
    let id = UUID()
    let currency: Currency
    let amount: Double
    let krwValue: Double
    let address: String
}

struct Rate: Identifiable {
    let id = UUID()
    let currency: Currency
    let krwRate: Double
}

struct Transaction: Identifiable {
    let id: String
    let type: TransactionType
    let merchantName: String?
    let amountKrw: Double
    let usedCurrency: Currency
    let usedAmount: Double
    let appliedRate: Double
    let status: TransactionStatus
    let createdAt: Date
}

struct DepositNetwork: Identifiable {
    let id = UUID()
    let network: NetworkType
    let address: String
    let minDeposit: Double
    let minDepositUnit: String
    let confirmations: Int
    let arrivalTime: String
    let networkDescription: String
}

struct CoinDepositInfo: Identifiable {
    let id = UUID()
    let currency: Currency
    let networks: [DepositNetwork]
}

// MARK: - API Mapping

extension Currency {
    init?(apiValue: String) {
        switch apiValue {
        case "SOL":                          self = .SOL
        case "ETH":                          self = .ETH
        case "USDT", "USDT_ERC20", "USDT_TRC20": self = .USDT
        default: return nil
        }
    }
}

struct DepositRecord: Identifiable {
    let id: String
    let txHash: String
    let currency: Currency
    let network: NetworkType
    let amount: Double
    let status: DepositStatus
    let confirmCount: Int
    let requiredConfirmCount: Int
    let createdAt: Date
}
