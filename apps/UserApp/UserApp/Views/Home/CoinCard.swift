import SwiftUI

struct CoinCard: View {
    let balance: CoinBalance

    var body: some View {
        ZStack(alignment: .topTrailing) {
            // Decorative symbol
            Text(balance.currency.symbol)
                .font(.system(size: 96, weight: .black))
                .foregroundStyle(balance.currency.accentColor.opacity(0.15))
                .offset(x: 12, y: -8)

            VStack(alignment: .leading, spacing: 0) {
                // Top: name
                VStack(alignment: .leading, spacing: 3) {
                    Text(balance.currency.rawValue)
                        .font(.system(size: 24, weight: .black))
                        .foregroundStyle(Color.appLabel)
                    Text(balance.currency.fullName)
                        .font(.system(size: 14, weight: .medium))
                        .foregroundStyle(Color.appSecondary)
                }

                Spacer()

                // Bottom: values
                VStack(alignment: .leading, spacing: 4) {
                    Text(balance.krwValue.krwFormatted)
                        .font(.system(size: 26, weight: .bold))
                        .foregroundStyle(balance.currency.accentColor)
                    Text("\(balance.amount.coinFormatted) \(balance.currency.rawValue)")
                        .font(.system(size: 14, weight: .medium))
                        .foregroundStyle(Color.appSecondary)
                }
            }
            .padding(24)
        }
        .frame(maxWidth: .infinity, minHeight: 170, alignment: .leading)
        .background(balance.currency.backgroundColor)
        .clipShape(RoundedRectangle(cornerRadius: 24, style: .continuous))
        .clipped()
    }
}

#Preview {
    CoinCard(balance: MockData.balances[0])
        .padding()
}
