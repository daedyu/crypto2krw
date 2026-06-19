import Foundation

// MARK: - API Response Models (Codable)

struct UserProfile: Decodable {
    let id: String
    let email: String
    let kycStatus: String?
    let status: String
    let createdAt: String
}

struct BalanceItem: Decodable, Identifiable {
    var id: String { currency }
    let currency: String
    let availableBalance: String
    let lockedBalance: String
}

struct WalletItem: Decodable, Identifiable {
    var id: String { address }
    let currency: String
    let address: String
    let network: String
    let paymentPriority: Int
}

struct TransactionItem: Decodable, Identifiable {
    let id: String
    let type: String
    let merchantName: String?
    let amountKrw: String?
    let usedCurrency: String
    let usedAmount: String
    let appliedRate: String?
    let status: String
    let createdAt: String
}

struct TransactionList: Decodable {
    let items: [TransactionItem]
    let total: Int
}

struct RateMap: Decodable {
    let sol: String
    let eth: String
    let usdt: String

    enum CodingKeys: String, CodingKey {
        case sol = "SOL"
        case eth = "ETH"
        case usdt = "USDT"
    }
}

// MARK: - Payment DTOs

struct CreateQRRequest: Encodable {
    let merchantId: String
    let amountKrw: String
}

struct QRSession: Decodable {
    let token: String
    let amountKrw: String
    let expiresAt: String
    let qrPayload: String
}

struct PayQRRequest: Encodable {
    let currency: String
}

struct PayQRResult: Decodable {
    let transactionId: String
    let merchantId: String
    let amountKrw: String
    let usedCurrency: String
    let usedAmount: String
    let appliedRate: String
    let remainingBalance: String
}

// MARK: - TransactionItem → Transaction mapping

extension TransactionItem {
    func toTransaction() -> Transaction? {
        guard let currency = Currency(apiValue: usedCurrency),
              let usedAmountD = Double(usedAmount) else { return nil }
        let txType: TransactionType = (type == "DEPOSIT") ? .deposit : .payment
        let txStatus: TransactionStatus
        switch status {
        case "COMPLETED": txStatus = .success
        case "PENDING":   txStatus = .pending
        default:          txStatus = .failed
        }
        let amountKrwD   = Double(amountKrw   ?? "0") ?? 0
        let appliedRateD = Double(appliedRate  ?? "0") ?? 0
        let date = ISO8601DateFormatter().date(from: createdAt) ?? Date()
        return Transaction(
            id: id,
            type: txType,
            merchantName: merchantName,
            amountKrw: amountKrwD,
            usedCurrency: currency,
            usedAmount: usedAmountD,
            appliedRate: appliedRateD,
            status: txStatus,
            createdAt: date
        )
    }
}

// MARK: - UserService

@MainActor
final class UserService: ObservableObject {
    static let shared = UserService()
    private let api = APIClient.shared

    private init() {}

    func getProfile() async throws -> UserProfile {
        try await api.request("/api/v1/users/me")
    }

    func getBalances() async throws -> [BalanceItem] {
        try await api.request("/api/v1/users/me/balances")
    }

    func getWallets() async throws -> [WalletItem] {
        try await api.request("/api/v1/users/me/wallets")
    }

    func getTransactions(limit: Int = 20, offset: Int = 0) async throws -> [TransactionItem] {
        try await api.request("/api/v1/users/me/transactions?limit=\(limit)&offset=\(offset)")
    }

    func getRates() async throws -> RateMap {
        try await api.request("/api/v1/rates", requiresAuth: false)
    }

    func createQRSession(merchantId: String, amountKRW: String) async throws -> QRSession {
        try await api.request(
            "/api/v1/payment/qr",
            method: "POST",
            body: CreateQRRequest(merchantId: merchantId, amountKrw: amountKRW)
        )
    }

    func payQR(token: String, currency: String) async throws -> PayQRResult {
        try await api.request(
            "/api/v1/payment/qr/\(token)/pay",
            method: "POST",
            body: PayQRRequest(currency: currency)
        )
    }
}
