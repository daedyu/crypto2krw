import SwiftUI
import CoreImage.CIFilterBuiltins

// MARK: - Coin Select

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
                                Text(balance.currency.rawValue).font(.headline)
                                Text(balance.currency.fullName).font(.caption).foregroundStyle(.secondary)
                            }
                            Spacer()
                            VStack(alignment: .trailing, spacing: 3) {
                                Text(balance.krwValue.krwFormatted)
                                    .font(.subheadline).bold()
                                    .foregroundStyle(balance.currency.accentColor)
                                Text("\(balance.amount.coinFormatted) \(balance.currency.rawValue)")
                                    .font(.caption).foregroundStyle(.secondary)
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
// 유저가 이 QR 코드를 가맹점 POS 카메라에 보여주면 결제 처리됨

struct QRDisplayView: View {
    let balance: CoinBalance

    @State private var qrImage: Image?
    @State private var timeLeft = 25
    private let total           = 25

    var body: some View {
        ScrollView {
            VStack(spacing: 24) {

                // QR 카드
                VStack(spacing: 16) {
                    Group {
                        if TokenStore.accessToken == nil {
                            // 로그인 안 된 상태
                            VStack(spacing: 12) {
                                Image(systemName: "lock.fill")
                                    .font(.system(size: 48))
                                    .foregroundStyle(.secondary)
                                Text("로그인 후 사용 가능합니다")
                                    .font(.subheadline)
                                    .foregroundStyle(.secondary)
                            }
                            .frame(width: 240, height: 240)
                        } else if let img = qrImage {
                            img
                                .interpolation(.none)
                                .resizable()
                                .scaledToFit()
                                .frame(width: 240, height: 240)
                        } else {
                            ProgressView()
                                .frame(width: 240, height: 240)
                        }
                    }

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

                // 타이머
                VStack(spacing: 8) {
                    ProgressView(value: Double(timeLeft), total: Double(total))
                        .tint(balance.currency.accentColor)
                    Text("\(timeLeft)초 후 자동 갱신")
                        .font(.caption).foregroundStyle(.secondary)
                        .frame(maxWidth: .infinity, alignment: .trailing)
                }

                // 잔액 카드
                GroupBox {
                    HStack {
                        VStack(alignment: .leading, spacing: 4) {
                            Text("보유 잔액").font(.caption).foregroundStyle(.secondary)
                            Text(balance.krwValue.krwFormatted)
                                .font(.title3).bold()
                                .foregroundStyle(balance.currency.accentColor)
                        }
                        Spacer()
                        VStack(alignment: .trailing, spacing: 4) {
                            Text("코인 수량").font(.caption).foregroundStyle(.secondary)
                            Text("\(balance.amount.coinFormatted) \(balance.currency.rawValue)")
                                .font(.subheadline).bold()
                        }
                    }
                }

                Text("가맹점 POS 카메라에 이 QR을 보여주세요")
                    .font(.footnote).foregroundStyle(.secondary)
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

    // QR 페이로드: 유저 액세스 토큰 + 코인 포함
    // POS가 이 토큰으로 결제를 처리함
    private func makePayload() -> String? {
        guard let at = TokenStore.accessToken, !at.isEmpty else { return nil }
        return "crypto2krw://pay?coin=\(balance.currency.rawValue)&at=\(at)"
    }

    // CoreImage로 QR 네이티브 생성 (외부 서비스 불필요)
    private func generateQR(from string: String) -> Image? {
        let context = CIContext()
        let filter  = CIFilter.qrCodeGenerator()
        filter.setValue(Data(string.utf8), forKey: "inputMessage")
        filter.setValue("M", forKey: "inputCorrectionLevel")
        guard let output = filter.outputImage else { return nil }
        let scale  = 480.0 / output.extent.size.width
        let scaled = output.transformed(by: CGAffineTransform(scaleX: scale, y: scale))
        guard let cg = context.createCGImage(scaled, from: scaled.extent) else { return nil }
        return Image(cg, scale: 1.0, label: Text("QR"))
    }

    private func refresh() {
        if let payload = makePayload() {
            qrImage = generateQR(from: payload)
        } else {
            qrImage = nil   // 로그인 필요 상태
        }
        timeLeft = total
    }
}

#Preview {
    NavigationStack { QRCoinSelectView() }
}
