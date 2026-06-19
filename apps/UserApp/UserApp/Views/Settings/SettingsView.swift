import SwiftUI

struct SettingsView: View {
    @Environment(AuthManager.self) private var auth
    @State private var wallets:  [WalletItem] = []
    @State private var expanded: String? = nil   // wallet.currency 문자열 키
    @State private var copied:   String? = nil

    var body: some View {
        List {
            Section("입금 주소") {
                if wallets.isEmpty {
                    Text("입금 주소를 불러오는 중...")
                        .foregroundStyle(.secondary)
                        .font(.subheadline)
                } else {
                    ForEach(wallets) { wallet in
                        let displayCurrency = Currency(apiValue: wallet.currency) ?? .USDT
                        DisclosureGroup(isExpanded: Binding(
                            get: { expanded == wallet.currency },
                            set: { expanded = $0 ? wallet.currency : nil }
                        )) {
                            VStack(alignment: .leading, spacing: 12) {
                                Text(wallet.address)
                                    .font(.system(.footnote, design: .monospaced))
                                    .foregroundStyle(.primary)
                                    .textSelection(.enabled)

                                Button {
                                    UIPasteboard.general.string = wallet.address
                                    copied = wallet.currency
                                    DispatchQueue.main.asyncAfter(deadline: .now() + 2) {
                                        copied = nil
                                    }
                                } label: {
                                    Label(
                                        copied == wallet.currency ? "복사됨" : "주소 복사",
                                        systemImage: copied == wallet.currency ? "checkmark" : "doc.on.doc"
                                    )
                                }
                                .buttonStyle(.bordered)
                                .tint(displayCurrency.accentColor)
                            }
                            .padding(.top, 4)
                        } label: {
                            HStack(spacing: 12) {
                                CoinBadge(currency: displayCurrency, size: 32)
                                VStack(alignment: .leading, spacing: 2) {
                                    Text(walletDisplayName(wallet))
                                        .font(.headline)
                                    Text(wallet.network)
                                        .font(.caption)
                                        .foregroundStyle(.secondary)
                                }
                            }
                        }
                    }
                }
            }

            Section("계정") {
                Button(role: .destructive) {
                    Task { await auth.logout() }
                } label: {
                    Label("로그아웃", systemImage: "rectangle.portrait.and.arrow.right")
                }
            }
        }
        .navigationTitle("설정")
        .task {
            if let items = try? await UserService.shared.getWallets() {
                wallets = items
            }
        }
    }

    private func walletDisplayName(_ wallet: WalletItem) -> String {
        switch wallet.currency {
        case "SOL":        return "Solana (SOL)"
        case "ETH":        return "Ethereum (ETH)"
        case "USDT_ERC20": return "USDT (ERC-20)"
        case "USDT_TRC20": return "USDT (TRC-20)"
        default:           return wallet.currency
        }
    }
}

#Preview {
    NavigationStack {
        SettingsView()
            .environment(AuthManager())
    }
}
