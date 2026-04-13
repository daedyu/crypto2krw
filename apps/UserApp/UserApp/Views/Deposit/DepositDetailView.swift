import SwiftUI

struct DepositDetailView: View {
    let currency: Currency

    private var depositInfo: CoinDepositInfo {
        MockData.depositInfo.first { $0.currency == currency }!
    }
    private var records: [DepositRecord] {
        MockData.depositRecords.filter { $0.currency == currency }
    }

    @State private var selectedNetwork: DepositNetwork? = nil
    @State private var copied = false

    private var network: DepositNetwork {
        selectedNetwork ?? depositInfo.networks[0]
    }

    var body: some View {
        List {
            // Network picker
            if depositInfo.networks.count > 1 {
                Section("네트워크") {
                    Picker("네트워크", selection: Binding(
                        get: { network.network.rawValue },
                        set: { val in selectedNetwork = depositInfo.networks.first { $0.network.rawValue == val } }
                    )) {
                        ForEach(depositInfo.networks) { n in
                            Text(n.network.rawValue).tag(n.network.rawValue)
                        }
                    }
                    .pickerStyle(.segmented)
                    .listRowInsets(.init(top: 8, leading: 16, bottom: 8, trailing: 16))
                }
            }

            // QR
            Section {
                VStack(spacing: 16) {
                    if let url = qrURL {
                        AsyncImage(url: url) { phase in
                            switch phase {
                            case .success(let img):
                                img.resizable().scaledToFit()
                                    .frame(width: 200, height: 200)
                            default:
                                ProgressView().frame(width: 200, height: 200)
                            }
                        }
                    }
                    Text(network.networkDescription)
                        .font(.caption)
                        .foregroundStyle(.secondary)
                        .multilineTextAlignment(.center)
                }
                .frame(maxWidth: .infinity)
                .padding(.vertical, 8)
            }

            // Address
            Section("입금 주소") {
                Text(network.address)
                    .font(.system(.footnote, design: .monospaced))
                    .textSelection(.enabled)

                Button {
                    UIPasteboard.general.string = network.address
                    copied = true
                    DispatchQueue.main.asyncAfter(deadline: .now() + 2) { copied = false }
                } label: {
                    Label(
                        copied ? "복사됨" : "주소 복사",
                        systemImage: copied ? "checkmark" : "doc.on.doc"
                    )
                }
                .buttonStyle(.bordered)
                .tint(currency.accentColor)

                ShareLink(item: network.address) {
                    Label("주소 공유", systemImage: "square.and.arrow.up")
                }
                .buttonStyle(.bordered)
            }

            // Info
            Section("입금 정보") {
                LabeledContent("최소 입금") {
                    Text("\(network.minDeposit.coinFormatted) \(network.minDepositUnit)")
                }
                LabeledContent("확인 횟수") {
                    Text("\(network.confirmations)회")
                }
                LabeledContent("도착 예상") {
                    Text(network.arrivalTime)
                }
            }

            // Warning
            Section {
                Label {
                    Text("반드시 **\(network.network.rawValue)** 네트워크로만 입금하세요. 다른 네트워크 입금 시 자산을 잃을 수 있습니다.")
                        .font(.footnote)
                } icon: {
                    Image(systemName: "exclamationmark.triangle.fill")
                        .foregroundStyle(.orange)
                }
            }

            // History
            if !records.isEmpty {
                Section("입금 내역") {
                    ForEach(records) { record in
                        HStack(spacing: 12) {
                            Image(systemName: record.status == .confirmed ? "checkmark.circle.fill" : "clock.fill")
                                .foregroundStyle(record.status == .confirmed ? .green : .orange)
                                .font(.title3)

                            VStack(alignment: .leading, spacing: 3) {
                                Text("\(record.amount.coinFormatted) \(currency.rawValue)")
                                    .font(.subheadline).bold()
                                Text(record.createdAt.dateFormatted)
                                    .font(.caption)
                                    .foregroundStyle(.secondary)
                            }

                            Spacer()

                            Text(record.status == .confirmed
                                 ? "완료"
                                 : "\(record.confirmCount)/\(record.requiredConfirmCount)")
                                .font(.caption).bold()
                                .foregroundStyle(record.status == .confirmed ? .green : .orange)
                        }
                        .padding(.vertical, 2)
                    }
                }
            }
        }
        .navigationTitle("\(currency.rawValue) 입금")
        .navigationBarTitleDisplayMode(.inline)
    }

    private var qrURL: URL? {
        let addr = network.address
        guard let encoded = addr.addingPercentEncoding(withAllowedCharacters: .urlQueryAllowed)
        else { return nil }
        return URL(string: "https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=\(encoded)")
    }
}

#Preview {
    NavigationStack {
        DepositDetailView(currency: .USDT)
    }
}
