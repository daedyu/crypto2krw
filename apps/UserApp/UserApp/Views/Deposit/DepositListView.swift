import SwiftUI

struct DepositListView: View {
    private let balances = MockData.balances

    var body: some View {
        List {
            Section {
                ForEach(balances) { balance in
                    NavigationLink {
                        DepositDetailView(currency: balance.currency)
                    } label: {
                        HStack(spacing: 14) {
                            CoinBadge(currency: balance.currency, size: 40)

                            VStack(alignment: .leading, spacing: 3) {
                                Text(balance.currency.rawValue)
                                    .font(.headline)
                                Text(balance.currency.fullName)
                                    .font(.caption)
                                    .foregroundStyle(.secondary)
                            }

                            Spacer()

                            VStack(alignment: .trailing, spacing: 3) {
                                Text(balance.krwValue.krwFormatted)
                                    .font(.subheadline).bold()
                                    .foregroundStyle(balance.currency.accentColor)
                                Text("\(balance.amount.coinFormatted) \(balance.currency.rawValue)")
                                    .font(.caption)
                                    .foregroundStyle(.secondary)
                            }
                        }
                        .padding(.vertical, 4)
                    }
                }
            } header: {
                Text("입금할 코인을 선택하세요")
            }
        }
        .navigationTitle("입금")
    }
}

#Preview {
    NavigationStack {
        DepositListView()
    }
}
