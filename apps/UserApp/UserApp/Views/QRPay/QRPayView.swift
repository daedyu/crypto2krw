import SwiftUI

// MARK: - Coin Select (Tab root)

struct QRCoinSelectView: View {
    private let balances = MockData.balances

    var body: some View {
        List {
            Section {
                ForEach(balances) { balance in
                    NavigationLink {
                        QRDisplayView(balance: balance)
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
                Text("결제에 사용할 코인을 선택하세요")
            }
        }
        .navigationTitle("QR 결제")
    }
}

// MARK: - QR Display

struct QRDisplayView: View {
    let balance: CoinBalance

    @State private var qrToken  = ""
    @State private var timeLeft = 30
    private let total = 30

    var body: some View {
        ScrollView {
            VStack(spacing: 24) {
                // QR Card
                VStack(spacing: 16) {
                    if let url = qrURL {
                        AsyncImage(url: url) { phase in
                            switch phase {
                            case .success(let img):
                                img.resizable().scaledToFit()
                                    .frame(width: 220, height: 220)
                            default:
                                ProgressView()
                                    .frame(width: 220, height: 220)
                            }
                        }
                    }

                    // Coin label
                    Label {
                        Text("\(balance.currency.rawValue) · \(balance.currency.fullName)")
                            .font(.subheadline).bold()
                            .foregroundStyle(balance.currency.accentColor)
                    } icon: {
                        CoinBadge(currency: balance.currency, size: 22)
                    }
                }
                .padding(28)
                .background(.background.secondary)
                .clipShape(RoundedRectangle(cornerRadius: 24, style: .continuous))
                .shadow(color: .black.opacity(0.06), radius: 16, y: 4)

                // Timer
                VStack(spacing: 8) {
                    ProgressView(value: Double(timeLeft), total: Double(total))
                        .tint(balance.currency.accentColor)
                    Text("\(timeLeft)초 후 자동 갱신")
                        .font(.caption)
                        .foregroundStyle(.secondary)
                        .frame(maxWidth: .infinity, alignment: .trailing)
                }

                // Balance card
                GroupBox {
                    HStack {
                        VStack(alignment: .leading, spacing: 4) {
                            Text("보유 잔액")
                                .font(.caption)
                                .foregroundStyle(.secondary)
                            Text(balance.krwValue.krwFormatted)
                                .font(.title3).bold()
                                .foregroundStyle(balance.currency.accentColor)
                        }
                        Spacer()
                        VStack(alignment: .trailing, spacing: 4) {
                            Text("코인 수량")
                                .font(.caption)
                                .foregroundStyle(.secondary)
                            Text("\(balance.amount.coinFormatted) \(balance.currency.rawValue)")
                                .font(.subheadline).bold()
                        }
                    }
                }

                Text("가맹점에게 이 QR 코드를 보여주세요")
                    .font(.footnote)
                    .foregroundStyle(.secondary)
                    .multilineTextAlignment(.center)
            }
            .padding()
        }
        .navigationTitle("\(balance.currency.rawValue) 결제")
        .navigationBarTitleDisplayMode(.inline)
        .onAppear { refresh() }
        .onReceive(Timer.publish(every: 1, on: .main, in: .common).autoconnect()) { _ in
            guard timeLeft > 0 else { refresh(); return }
            timeLeft -= 1
        }
    }

    private var qrURL: URL? {
        guard !qrToken.isEmpty,
              let encoded = qrToken.addingPercentEncoding(withAllowedCharacters: .urlQueryAllowed)
        else { return nil }
        return URL(string: "https://api.qrserver.com/v1/create-qr-code/?size=220x220&data=\(encoded)")
    }

    private func refresh() {
        qrToken  = "crypto2krw://pay?coin=\(balance.currency.rawValue)&ts=\(Int(Date().timeIntervalSince1970))"
        timeLeft = total
    }
}

#Preview {
    NavigationStack {
        QRCoinSelectView()
    }
}
