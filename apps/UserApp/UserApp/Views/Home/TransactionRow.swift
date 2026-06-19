import SwiftUI

struct TransactionRow: View {
    let tx: Transaction

    private var isDeposit: Bool { tx.type == .deposit }

    private var iconName: String { isDeposit ? "arrow.down" : "arrow.up" }

    private var accentColor: Color { isDeposit ? Color(hex: "#26A17B") : Color(hex: "#9945FF") }

    private var title: String {
        tx.merchantName ?? (isDeposit ? "\(tx.usedCurrency.rawValue) 입금" : "\(tx.usedCurrency.rawValue) 출금")
    }

    var body: some View {
        HStack(spacing: 12) {
            // 아이콘
            ZStack {
                RoundedRectangle(cornerRadius: 12, style: .continuous)
                    .fill(accentColor.opacity(0.15))
                    .frame(width: 44, height: 44)
                Image(systemName: iconName)
                    .font(.system(size: 15, weight: .semibold))
                    .foregroundStyle(accentColor)
            }

            // 내용
            VStack(alignment: .leading, spacing: 3) {
                Text(title)
                    .font(.system(size: 15, weight: .semibold))
                    .foregroundStyle(.primary)
                Text(tx.createdAt.shortFormatted)
                    .font(.system(size: 12))
                    .foregroundStyle(.secondary)
            }

            Spacer()

            // 금액
            VStack(alignment: .trailing, spacing: 3) {
                Text("\(isDeposit ? "+" : "-")\(tx.amountKrw.krwFormatted)")
                    .font(.system(size: 16, weight: .bold))
                    .foregroundStyle(isDeposit ? accentColor : .primary)
                Text("\(tx.usedAmount.coinFormatted) \(tx.usedCurrency.rawValue)")
                    .font(.system(size: 12))
                    .foregroundStyle(.secondary)
            }
        }
        .padding(.vertical, 14)
    }
}

#Preview {
    VStack(spacing: 0) {
        ForEach(MockData.transactions) { tx in
            TransactionRow(tx: tx)
            Divider().padding(.leading, 72)
        }
    }
    .padding(.horizontal, 16)
}
