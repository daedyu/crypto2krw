import SwiftUI

struct SettingsView: View {
    @Environment(AuthManager.self) private var auth
    @State private var expanded: Currency? = nil
    @State private var copied:   Currency? = nil

    private let balances = MockData.balances

    var body: some View {
        List {
            Section("입금 주소") {
                ForEach(balances) { balance in
                    DisclosureGroup(isExpanded: Binding(
                        get: { expanded == balance.currency },
                        set: { expanded = $0 ? balance.currency : nil }
                    )) {
                        VStack(alignment: .leading, spacing: 12) {
                            Text(balance.address)
                                .font(.system(.footnote, design: .monospaced))
                                .foregroundStyle(.primary)
                                .textSelection(.enabled)

                            Button {
                                UIPasteboard.general.string = balance.address
                                copied = balance.currency
                                DispatchQueue.main.asyncAfter(deadline: .now() + 2) {
                                    copied = nil
                                }
                            } label: {
                                Label(
                                    copied == balance.currency ? "복사됨" : "주소 복사",
                                    systemImage: copied == balance.currency ? "checkmark" : "doc.on.doc"
                                )
                            }
                            .buttonStyle(.bordered)
                            .tint(balance.currency.accentColor)
                        }
                        .padding(.top, 4)
                    } label: {
                        HStack(spacing: 12) {
                            CoinBadge(currency: balance.currency, size: 32)
                            Text(balance.currency.rawValue)
                                .font(.headline)
                        }
                    }
                }
            }

            Section("계정") {
                Button(role: .destructive) {
                    auth.logout()
                } label: {
                    Label("로그아웃", systemImage: "rectangle.portrait.and.arrow.right")
                }
            }
        }
        .navigationTitle("설정")
    }
}

#Preview {
    NavigationStack {
        SettingsView()
            .environment(AuthManager())
    }
}
