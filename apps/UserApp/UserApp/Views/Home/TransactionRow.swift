import SwiftUI

struct TransactionRow: View {
    let tx: Transaction

    private var iconName: String {
        tx.type == .deposit ? "arrow.down" : "bag"
    }

    private var iconColor: Color {
        tx.type == .deposit ? Color(hex: "#26A17B") : Color(hex: "#9945FF")
    }

    private var iconBg: Color {
        tx.type == .deposit ? Color(hex: "#D4EDE5") : Color(hex: "#E4D9FF")
    }

    private var title: String {
        tx.merchantName ?? (tx.type == .deposit ? "\(tx.usedCurrency.rawValue) 입금" : "결제")
    }

    var body: some View {
        HStack(spacing: 12) {
            // Icon
            ZStack {
                RoundedRectangle(cornerRadius: 12, style: .continuous)
                    .fill(iconBg)
                    .frame(width: 44, height: 44)
                Image(systemName: iconName)
                    .font(.system(size: 16, weight: .semibold))
                    .foregroundStyle(iconColor)
            }

            // Info
            VStack(alignment: .leading, spacing: 3) {
                Text(title)
                    .font(.system(size: 15, weight: .semibold))
                    .foregroundStyle(Color.appLabel)
                Text(tx.createdAt.shortFormatted)
                    .font(.system(size: 12))
                    .foregroundStyle(Color.appSecondary)
            }

            Spacer()

            // Amount
            VStack(alignment: .trailing, spacing: 3) {
                Text("\(tx.type == .deposit ? "+" : "-")\(tx.amountKrw.krwFormatted)")
                    .font(.system(size: 16, weight: .bold))
                    .foregroundStyle(tx.type == .deposit ? iconColor : Color.appLabel)
                Text("\(tx.usedAmount.coinFormatted) \(tx.usedCurrency.rawValue)")
                    .font(.system(size: 12))
                    .foregroundStyle(Color.appSecondary)
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
